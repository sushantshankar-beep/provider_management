package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"provider_management/internal/constants"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strconv"
	"strings"
	"time"
)

type PayoutService struct {
	serviceRepo    *repository.AcceptedServiceRepo
	payoutRepo     *repository.PaymentPayoutRepo
	providerRepo   *repository.ProviderRepo
	settlementRepo *repository.ProviderSettlementRepo
}

type ProviderPayoutDetails struct {
	ProviderDetails   map[string]any   `json:"provider_details"`
	AccountDetails    map[string]any   `json:"account_details"`
	EarningSummary    map[string]any   `json:"earning_summary"`
	SettlementHistory []map[string]any `json:"settlement_history"`
}

func NewPayoutService(serviceRepo *repository.AcceptedServiceRepo, payoutRepo *repository.PaymentPayoutRepo, providerRepo *repository.ProviderRepo, settlementRepo *repository.ProviderSettlementRepo) *PayoutService {
	return &PayoutService{
		serviceRepo:    serviceRepo,
		payoutRepo:     payoutRepo,
		providerRepo:   providerRepo,
		settlementRepo: settlementRepo,
	}
}

type DeductionPayoutRequest struct {
	ProviderID          string
	BookingID           string
	OriginalAmount      float64
	DeductionAmount     float64
	RemainingAmount     float64
	Reason              string
	ComplaintID         string
	ComplaintInternalID int64
}

func (s *PayoutService) CreatePayoutLast6Hours(ctx context.Context) error {
	to := time.Now()
	from := to.Add(-6 * time.Hour)

	services, err := s.serviceRepo.FindCompletedPaidBetween(ctx, from, to)
	if err != nil {
		return err
	}

	type bucket struct {
		ServiceIDs []primitive.ObjectID
		Total      float64
	}

	group := map[string]*bucket{}

	for _, svc := range services {

		if svc.IsPayoutCancelled {
			continue
		}

		pid := svc.ProviderID.Hex()
		if _, ok := group[pid]; !ok {
			group[pid] = &bucket{}
		}

		group[pid].ServiceIDs = append(group[pid].ServiceIDs, svc.ID)
		group[pid].Total += svc.FinalPrice
	}

	for providerID, data := range group {
		providerObjID, _ := primitive.ObjectIDFromHex(providerID)

		existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerObjID)
		if err != nil {
			return err
		}

		commissionPercent := 20.0
		gstPercent := 18.0

		if existing != nil {
			if existing.ServicePartialAmounts == nil {
				existing.ServicePartialAmounts = make(map[string]float64)
			}

			existing.ServiceIDs = append(existing.ServiceIDs, data.ServiceIDs...)
			existing.BaseAmount += data.Total

			totalEffectiveAmount := s.calculateEffectiveAmount(ctx, existing)

			commission := totalEffectiveAmount * commissionPercent / 100
			afterCommission := totalEffectiveAmount - commission
			gst := afterCommission * gstPercent / 100
			netPayable := afterCommission - gst

			existing.CommissionAmount = utils.RoundTo2(commission)
			existing.GSTAmount = utils.RoundTo2(gst)
			existing.NetPayable = utils.RoundTo2(netPayable)
			existing.UpdatedAt = time.Now()

			if err := s.payoutRepo.Update(ctx, existing); err != nil {
				return err
			}
		} else {
			commission := data.Total * commissionPercent / 100
			afterCommission := data.Total - commission
			gst := afterCommission * gstPercent / 100
			netPayable := afterCommission - gst

			payout := &domain.PaymentPayout{
				PayoutID:              time.Now().UnixMilli(),
				ProviderID:            providerObjID,
				ServiceIDs:            data.ServiceIDs,
				BaseAmount:            utils.RoundTo2(data.Total),
				ServicePartialAmounts: make(map[string]float64),
				CommissionPercent:     commissionPercent,
				CommissionAmount:      utils.RoundTo2(commission),
				GSTPercent:            gstPercent,
				GSTAmount:             utils.RoundTo2(gst),
				NetPayable:            utils.RoundTo2(netPayable),
				Status:                domain.PayoutStatusPending,
				PayoutType:            domain.PayoutTypeRegular,
				PeriodFrom:            from,
				PeriodTo:              to,
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			}

			if err := s.payoutRepo.Create(ctx, payout); err != nil {
				return err
			}
		}

		if err := s.serviceRepo.MarkPayoutCreated(ctx, data.ServiceIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *PayoutService) calculateEffectiveAmount(ctx context.Context, payout *domain.PaymentPayout) float64 {
	total := 0.0

	for _, serviceID := range payout.ServiceIDs {
		serviceKey := serviceID.Hex()
		if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists {
			total += partialAmt
			continue
		}
		service, err := s.serviceRepo.FindByID(ctx, serviceKey)

		if err != nil {
			continue
		}

		if service.IsPayoutCancelled {
			continue
		}

		total += service.FinalPrice
	}

	return total
}

func (s *PayoutService) GetPayouts(
	ctx context.Context,
	page, limit int64,
	search, providerID, status, periodFromStr, periodToStr, sortBy, sortOrder string,
) ([]map[string]any, int64, int64, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if providerID != "" {
		if objID, err := primitive.ObjectIDFromHex(providerID); err == nil {
			filter["providerId"] = objID
		}
	}

	if status != "" {
		switch strings.ToLower(status) {

		case "pending_settlement":
			filter["status"] = bson.M{
				"$in": []string{
					string(domain.PayoutStatusPending),
					string(domain.PayoutStatusPartiallySettled),
				},
			}
		default:
			filter["status"] = status
		}
	}

	if periodFromStr != "" && periodToStr != "" {
		from, err1 := time.Parse(time.RFC3339, periodFromStr)
		to, err2 := time.Parse(time.RFC3339, periodToStr)
		if err1 == nil && err2 == nil {
			filter["createdAt"] = bson.M{"$gte": from, "$lte": to}
		}
	}

	if search != "" {
		orFilters := []bson.M{}

		if objID, err := primitive.ObjectIDFromHex(search); err == nil {
			orFilters = append(orFilters, bson.M{"providerId": objID})
		}

		if len(search) > 3 && strings.ToUpper(search[:3]) == "PAY" {
			if payoutID, err := strconv.ParseInt(search[3:], 10, 64); err == nil {
				orFilters = append(orFilters, bson.M{"payoutId": payoutID})
			}
		}

		if payoutID, err := strconv.ParseInt(search, 10, 64); err == nil {
			orFilters = append(orFilters, bson.M{"payoutId": payoutID})
		}

		orFilters = append(orFilters, bson.M{"status": bson.M{"$regex": search, "$options": "i"}})

		if len(orFilters) > 0 {
			filter["$or"] = orFilters
		}
	}

	order := -1
	if sortOrder == "asc" {
		order = 1
	}

	if sortBy == "" {
		sortBy = "createdAt"
	} else {
		switch sortBy {
		case "payout_id":
			sortBy = "payoutId"
		case "provider_id":
			sortBy = "providerId"
		case "base_amount":
			sortBy = "baseAmount"
		case "net_payable":
			sortBy = "netPayable"
		case "status":
			sortBy = "status"
		case "created_at":
			sortBy = "createdAt"
		case "updated_at":
			sortBy = "updatedAt"
		default:
			sortBy = sortBy
		}
	}

	payouts, total, err := s.payoutRepo.GetPayouts(ctx, filter, skip, limit, sortBy, order)
	if err != nil {
		return nil, 0, 0, err
	}

	providerIDSet := make(map[primitive.ObjectID]struct{})
	for _, p := range payouts {
		providerIDSet[p.ProviderID] = struct{}{}
	}

	providerIDs := make([]primitive.ObjectID, 0, len(providerIDSet))
	for id := range providerIDSet {
		providerIDs = append(providerIDs, id)
	}

	providers, err := s.providerRepo.FindByObjectIDs(ctx, providerIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	providerMap := make(map[primitive.ObjectID]string)
	for _, pr := range providers {
		providerMap[pr.ID] = pr.Name
	}

	responseData := make([]map[string]any, len(payouts))
	for i, p := range payouts {
		responseData[i] = map[string]any{
			"id":                 p.ID.Hex(),
			"payout_id":          "PAY" + strconv.FormatInt(p.PayoutID, 10),
			"provider_id":        p.ProviderID.Hex(),
			"provider_name":      providerMap[p.ProviderID],
			"service_ids":        p.ServiceIDs,
			"base_amount":        utils.RoundTo2(p.BaseAmount),
			"commission_percent": p.CommissionPercent,
			"commission_amount":  utils.RoundTo2(p.CommissionAmount),
			"gst_percent":        p.GSTPercent,
			"gst_amount":         utils.RoundTo2(p.GSTAmount),
			"net_payable":        utils.RoundTo2(p.NetPayable),
			"status":             p.Status,
			"period_from":        p.PeriodFrom,
			"period_to":          p.PeriodTo,
			"created_at":         p.CreatedAt,
			"updated_at":         p.UpdatedAt,
		}
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return responseData, total, totalPages, nil
}

func (s *PayoutService) GetPayoutServices(ctx context.Context, payoutID string) ([]map[string]any, error) {
	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "PAY") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filter := bson.M{"payoutId": payoutIDInt}

	payouts, _, err := s.payoutRepo.GetPayouts(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}

	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	var services []map[string]any
	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
		if err != nil {
			log.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		if service.IsSettled && !service.HasComplaintAdjustment {
			continue
		}

		partialAmount := 0.0
		hasPartialAmount := false
		isDeduction := false

		if payout.ServicePartialAmounts != nil {
			if amt, exists := payout.ServicePartialAmounts[serviceID.Hex()]; exists {
				partialAmount = amt
				if amt == 0 {
					isDeduction = true
				} else {
					hasPartialAmount = true
				}
			}
		}

		var amount float64

		if isDeduction {
			amount = service.FinalPrice
			partialAmount = 0
		} else if hasPartialAmount {
			amount = partialAmount
		} else {
			amount = service.FinalPrice
			partialAmount = 0
		}

		serviceCommission := amount * payout.CommissionPercent / 100
		afterCommission := amount - serviceCommission
		serviceGST := afterCommission * payout.GSTPercent / 100
		serviceNet := afterCommission - serviceGST

		showComplaintAdjustment := false
		if service.HasComplaintAdjustment && service.IsSettled && !service.IsSettledAfterComplaint {
			showComplaintAdjustment = true
		}

		serviceData := map[string]any{
			"id":                         service.ID,
			"booking_id":                 fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":                     "amc",
			"provider_id":                payout.ProviderID.Hex(),
			"service_amount":             utils.RoundTo2(service.FinalPrice),
			"commission_percent":         payout.CommissionPercent,
			"commission_amount":          utils.RoundTo2(serviceCommission),
			"gst_percent":                payout.GSTPercent,
			"gst_amount":                 utils.RoundTo2(serviceGST),
			"net_amount":                 utils.RoundTo2(serviceNet),
			"partial_amount":             utils.RoundTo2(partialAmount),
			"payout_id":                  fmt.Sprintf("SET%d", payout.PayoutID),
			"is_settled":                 service.IsSettled,
			"settlement_id":              service.SettlementID,
			"settled_at":                 service.SettledAt,
			"has_complaint_adjustment":   service.HasComplaintAdjustment,
			"pending_settlement":         utils.RoundTo2(service.PendingDeductionAmount),
			"show_complaint_adjustment":  showComplaintAdjustment,
			"payout_status":              service.PayoutStatus,
			"isPayoutCancelled":          service.IsPayoutCancelled,
			"is_settled_after_complaint": service.IsSettledAfterComplaint,
		}

		services = append(services, serviceData)
	}

	return services, nil
}

func (s *PayoutService) GetPayoutProviderData(ctx context.Context, payoutID string) ([]map[string]any, error) {
	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "PAY") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filter := bson.M{"payoutId": payoutIDInt}

	payouts, _, err := s.payoutRepo.GetPayouts(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}
	log.Println("payoutDataaa", payouts)
	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	var services []map[string]any
	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
		if err != nil {
			log.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		serviceData := map[string]any{
			"booking_id":         fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":             "amc",
			"provider_id":        payout.ProviderID.Hex(),
			"service_amount":     utils.RoundTo2(service.FinalPrice),
			"commission_percent": payout.CommissionPercent,
			"gst_percent":        utils.RoundTo2(payout.GSTPercent),
			"partial_amount":     "₹",
			"payout_id":          fmt.Sprintf("SET%d", payout.PayoutID),
		}

		services = append(services, serviceData)
	}

	return services, nil
}

func (s *PayoutService) GetProviderPayoutDetails(ctx context.Context, payoutID string) (*ProviderPayoutDetails, error) {
	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "PAY") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filter := bson.M{"payoutId": payoutIDInt}
	payouts, _, err := s.payoutRepo.GetPayouts(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}

	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	provider, err := s.providerRepo.FindByID(ctx, payout.ProviderID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provider: %v", err)
	}

	brandNames := []string{}
	if len(provider.ProviderBrands) > 0 {
		brandNames = constants.GetBrandNamesByIDs(provider.ProviderBrands)
	}

	serviceNames := []string{}
	if len(provider.ProviderServices) > 0 {
		serviceNames = constants.GetServiceNamesByIDs(provider.ProviderServices)
	}

	providerDetails := map[string]any{
		"name":          provider.Name,
		"phone_number":  provider.Phone,
		"provider_id":   payout.ProviderID.Hex(),
		"email_id":      provider.Email,
		"mechanic_type": strings.Join(provider.VehicleType, ", "),
		"vehicle_brand": strings.Join(brandNames, ", "),
		"zone":          provider.City,
		"join_date":     provider.CreatedAt,
		"approval_date": provider.UpdatedAt,
		"service_type":  strings.Join(serviceNames, ", "),
	}

	accountDetails := map[string]any{
		"account_holder_name": "",
		"branch_name":         "",
		"ifsc_code":           "",
		"upi_id":              "",
		"gst_number":          provider.GSTNumber,
		"account_verified":    provider.Status,
	}

	if provider.BankDetails != nil {
		accountDetails["account_holder_name"] = provider.BankDetails.AccountHolderName
		accountDetails["branch_name"] = provider.BankDetails.BranchName
		accountDetails["ifsc_code"] = provider.BankDetails.IfscCode
		accountDetails["upi_id"] = provider.BankDetails.Upi
	}

	totalEarnings := 0.0
	totalSettled := 0.0
	pendingSettlement := 0.0
	totalGST := 0.0
	adjustments := 0.0

	allPayouts, _, err := s.payoutRepo.GetPayouts(ctx, bson.M{"providerId": payout.ProviderID}, 0, 1000, "createdAt", -1)
	if err == nil {
		for _, p := range allPayouts {
			totalEarnings += p.BaseAmount
			totalGST += p.GSTAmount

			if p.Status == "settled" {
				totalSettled += p.NetPayable
			} else if p.Status == "pending" {
				pendingSettlement += p.NetPayable
			}
		}
	}

	earningSummary := map[string]any{
		"total_earnings":     utils.RoundTo2(totalEarnings),
		"total_settled":      utils.RoundTo2(totalSettled),
		"pending_settlement": utils.RoundTo2(pendingSettlement),
		"gst":                utils.RoundTo2(totalGST),
		"adjustments":        utils.RoundTo2(adjustments),
	}

	settlementHistory := []map[string]any{}

	if s.settlementRepo != nil {
		settlements, _, err := s.settlementRepo.GetSettlements(ctx, bson.M{
			"providerId": payout.ProviderID,
			"status":     "settled",
		}, 0, 100, "settledAt", -1)

		if err == nil {
			for _, settlement := range settlements {
				settlementHistory = append(settlementHistory, map[string]any{
					"date":          settlement.SettledAt,
					"settlement_id": fmt.Sprintf("SET%06d", settlement.SettlementID),
					"amount":        utils.RoundTo2(settlement.TotalAmount),
					"payment_mode":  settlement.PaymentMode,
					"method":        settlement.PaymentMethod,
				})
			}
		}
	} else {
		settledPayouts, _, err := s.payoutRepo.GetPayouts(ctx, bson.M{
			"providerId": payout.ProviderID,
			"status":     "settled",
		}, 0, 100, "updatedAt", -1)

		if err == nil {
			for _, p := range settledPayouts {
				if p.SettlementID != nil {
					settlementHistory = append(settlementHistory, map[string]any{
						"date":          p.UpdatedAt,
						"settlement_id": fmt.Sprintf("UTR-IMPS-%s", p.SettlementID.Hex()[:6]),
						"amount":        utils.RoundTo2(p.NetPayable),
					})
				}
			}
		}
	}

	return &ProviderPayoutDetails{
		ProviderDetails:   providerDetails,
		AccountDetails:    accountDetails,
		EarningSummary:    earningSummary,
		SettlementHistory: settlementHistory,
	}, nil
}

func (s *PayoutService) ProcessPayout(ctx context.Context, req PayoutRequest) error {
	log.Println("ProcessPayout Request:", req)

	providerObjID, err := primitive.ObjectIDFromHex(req.ProviderID)
	if err != nil {
		log.Printf("ERROR ProcessPayout: Invalid ProviderID: %v", err)
		return err
	}

	serviceIDs := []primitive.ObjectID{}
	if req.BookingID != "" {
		if sid, err := primitive.ObjectIDFromHex(req.BookingID); err == nil {
			serviceIDs = append(serviceIDs, sid)
		}
	}

	commissionPercent := 20.0
	gstPercent := 18.0

	existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerObjID)
	if err != nil {
		log.Printf("ERROR ProcessPayout: FindPendingByProvider failed: %v", err)
		return err
	}

	if existing != nil {
		if existing.ServicePartialAmounts == nil {
			existing.ServicePartialAmounts = make(map[string]float64)
		}

		bookingExists := false
		for _, existingID := range existing.ServiceIDs {
			if len(serviceIDs) > 0 && existingID == serviceIDs[0] {
				bookingExists = true
				break
			}
		}

		if !bookingExists && len(serviceIDs) > 0 {
			existing.ServiceIDs = append(existing.ServiceIDs, serviceIDs...)

			if req.Amount > 0 {
				existing.BaseAmount += req.Amount
			}
		}

		if len(serviceIDs) > 0 {
			key := serviceIDs[0].Hex()

			if req.CancelPayout {
				existing.ServicePartialAmounts[key] = 0
				existing.IsPayoutCancelled = true

				if err := s.serviceRepo.UpdatePayoutCancellation(ctx, serviceIDs[0].Hex(), true); err != nil {
					log.Printf("ERROR: Failed to mark service as cancelled: %v", err)
				}
			} else if req.PartialAmount > 0 {
				existing.ServicePartialAmounts[key] = req.PartialAmount
			}
		}

		totalEffectiveAmount := s.calculateEffectiveAmount(ctx, existing)

		commission := totalEffectiveAmount * commissionPercent / 100
		gst := (totalEffectiveAmount - commission) * gstPercent / 100
		netPayable := totalEffectiveAmount - commission - gst

		existing.CommissionAmount = utils.RoundTo2(commission)
		existing.GSTAmount = utils.RoundTo2(gst)
		existing.NetPayable = utils.RoundTo2(netPayable)
		existing.UpdatedAt = time.Now()

		log.Printf("Updated Existing Payout: BaseAmount=%.2f EffectiveAmount=%.2f NetPayable=%.2f",
			existing.BaseAmount, totalEffectiveAmount, netPayable)

		return s.payoutRepo.Update(ctx, existing)
	}

	if req.CreateDeduction {
		var effectiveAmount float64
		if req.PartialAmount > 0 {
			effectiveAmount = req.PartialAmount
		} else {
			effectiveAmount = req.Amount
		}

		commission := effectiveAmount * commissionPercent / 100
		gst := (effectiveAmount - commission) * gstPercent / 100
		netDeduction := effectiveAmount - commission - gst

		complaintObjID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

		payout := &domain.PaymentPayout{
			PayoutID:              time.Now().UnixMilli(),
			ProviderID:            providerObjID,
			ServiceIDs:            serviceIDs,
			ComplaintID:           &complaintObjID,
			ComplaintInternalID:   &req.ComplaintInternalID,
			BaseAmount:            -req.Amount,
			ServicePartialAmounts: make(map[string]float64),
			CommissionPercent:     commissionPercent,
			CommissionAmount:      -utils.RoundTo2(commission),
			GSTPercent:            gstPercent,
			GSTAmount:             -utils.RoundTo2(gst),
			NetPayable:            -utils.RoundTo2(netDeduction),
			IsDeduction:           true,
			Status:                domain.PayoutStatusPending,
			PayoutType:            domain.PayoutTypeComplaint,
			PeriodFrom:            time.Now().Add(-24 * time.Hour),
			PeriodTo:              time.Now(),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			Remarks:               req.Reason,
		}

		log.Printf("Deduction Created: Amount=%.2f NetDeduction=-%.2f", effectiveAmount, netDeduction)

		return s.payoutRepo.Create(ctx, payout)
	}

	servicePartialAmounts := make(map[string]float64)
	var effectiveAmount float64

	if req.CancelPayout {
		if len(serviceIDs) > 0 {
			servicePartialAmounts[serviceIDs[0].Hex()] = 0
			if err := s.serviceRepo.UpdatePayoutCancellation(ctx, serviceIDs[0].Hex(), true); err != nil {
				log.Printf("ERROR: Failed to mark service as cancelled: %v", err)
			}
		}
		effectiveAmount = 0
	} else if req.PartialAmount > 0 {
		if len(serviceIDs) > 0 {
			servicePartialAmounts[serviceIDs[0].Hex()] = req.PartialAmount
		}
		effectiveAmount = req.PartialAmount
	} else {
		if req.Amount > 0 {
			effectiveAmount = req.Amount
		} else if len(serviceIDs) > 0 {
			service, err := s.serviceRepo.FindByID(ctx, serviceIDs[0].Hex())
			if err != nil {
				log.Printf("ERROR: Failed to fetch service price: %v", err)
			} else {
				effectiveAmount = service.FinalPrice
			}
		}
	}

	commission := effectiveAmount * commissionPercent / 100
	gst := (effectiveAmount - commission) * gstPercent / 100
	netPayable := effectiveAmount - commission - gst

	complaintObjID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerObjID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintObjID,
		ComplaintInternalID:   &req.ComplaintInternalID,
		BaseAmount:            req.Amount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      utils.RoundTo2(commission),
		GSTPercent:            gstPercent,
		GSTAmount:             utils.RoundTo2(gst),
		NetPayable:            utils.RoundTo2(netPayable),
		IsPayoutCancelled:     req.CancelPayout,
		Status:                domain.PayoutStatusPending,
		PayoutType:            domain.PayoutTypeComplaint,
		PeriodFrom:            time.Now().Add(-24 * time.Hour),
		PeriodTo:              time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Remarks:               req.Reason,
	}

	log.Printf("New Payout Created: BaseAmount=%.2f Effective=%.2f NetPayable=%.2f Cancel=%t",
		req.Amount, effectiveAmount, netPayable, req.CancelPayout)

	return s.payoutRepo.Create(ctx, payout)
}

func (s *PayoutService) GetProviderPendingDeductions(ctx context.Context, providerID string) (float64, error) {
	providerObjID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return 0, fmt.Errorf("invalid provider ID: %w", err)
	}

	deductions, err := s.payoutRepo.GetProviderDeductions(ctx, providerObjID)
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, d := range deductions {
		total += d.PartialAmount
	}

	return utils.RoundTo2(total), nil
}

func (s *PayoutService) ProcessDeductionPayout(ctx context.Context, req DeductionPayoutRequest) error {
	providerObjID, err := primitive.ObjectIDFromHex(req.ProviderID)
	if err != nil {
		return err
	}

	serviceIDs := []primitive.ObjectID{}
	if req.BookingID != "" {
		if sid, err := primitive.ObjectIDFromHex(req.BookingID); err == nil {
			serviceIDs = append(serviceIDs, sid)
		}
	}

	commissionPercent := 20.0
	gstPercent := 18.0

	complaintObjID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	servicePartialAmounts := make(map[string]float64)
	if len(serviceIDs) > 0 {
		servicePartialAmounts[serviceIDs[0].Hex()] = req.DeductionAmount
	}

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerObjID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintObjID,
		ComplaintInternalID:   &req.ComplaintInternalID,
		BaseAmount:            req.OriginalAmount,
		PartialAmount:         req.DeductionAmount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      0,
		GSTPercent:            gstPercent,
		GSTAmount:             0,
		NetPayable:            -req.DeductionAmount,
		PayoutType:            domain.PayoutTypeComplaint,
		IsDeduction:           true,
		Status:                domain.PayoutStatusPending,
		PeriodFrom:            time.Now().Add(-24 * time.Hour),
		PeriodTo:              time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Remarks:               req.Reason,
	}

	log.Printf("ProcessDeductionPayout - Created deduction payout: Deduction=%.2f (Negative settlement)", req.DeductionAmount)

	return s.payoutRepo.Create(ctx, payout)
}

func (s *PayoutService) AdjustExistingPayout(ctx context.Context, req AdjustPayoutRequest) error {
	providerObjID, err := primitive.ObjectIDFromHex(req.ProviderID)
	log.Println("dkjcnjcnsjbc", req)
	if err != nil {
		return fmt.Errorf("invalid provider ID: %w", err)
	}

	serviceObjID, err := primitive.ObjectIDFromHex(req.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking ID: %w", err)
	}

	existingPayout, err := s.payoutRepo.FindPendingByProviderAndService(ctx, providerObjID, serviceObjID)
	if err != nil {
		log.Printf("ERROR: Failed to find existing payout: %v", err)
		return fmt.Errorf("failed to find existing payout: %w", err)
	}

	if existingPayout == nil {
		return fmt.Errorf("no existing payout found for booking")
	}

	complaintObjID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	adjustment := domain.ComplaintAdjustment{
		ServiceID:           serviceObjID,
		ComplaintID:         complaintObjID,
		ComplaintInternalID: req.ComplaintInternalID,
		Amount:              req.NetDeduction,
		CreatedAt:           time.Now(),
	}

	servicePartialAmountsKey := fmt.Sprintf("servicePartialAmounts.%s", req.BookingID)

	if err := s.payoutRepo.UpdateFields(ctx, existingPayout.ID.Hex(), bson.M{
		"$set": bson.M{
			servicePartialAmountsKey: 0,
			"updatedAt":              time.Now(),
		},
		"$inc": bson.M{
			"netPayable": -round2(req.NetDeduction),
		},
		"$push": bson.M{
			"complaintAdjustments": adjustment,
		},
	}); err != nil {
		return fmt.Errorf("failed to update existing payout: %w", err)
	}

	if err := s.serviceRepo.UpdatePayoutStatus(ctx, req.BookingID, map[string]any{
		"payoutCreated":   false,
		"payoutCreatedAt": time.Time{},
	}); err != nil {
		log.Printf("Warning: Failed to update payout created flag: %v", err)
	}

	log.Printf("SUCCESS: Adjusted existing payout for provider %s, booking %s, deduction %.2f", req.ProviderID, req.BookingID, req.NetDeduction)
	return nil
}
