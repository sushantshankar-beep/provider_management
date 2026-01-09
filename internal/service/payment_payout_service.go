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

			totalPartialAmount := 0.0
			for _, amt := range existing.ServicePartialAmounts {
				totalPartialAmount += amt
			}
			existing.PartialAmount = totalPartialAmount

			effectiveAmount := existing.BaseAmount - existing.PartialAmount
			commission := effectiveAmount * commissionPercent / 100
			gst := commission * gstPercent / 100

			existing.CommissionAmount = utils.RoundTo2(commission)
			existing.GSTAmount = utils.RoundTo2(gst)
			existing.NetPayable = utils.RoundTo2(effectiveAmount - commission - gst)
			existing.UpdatedAt = time.Now()

			if err := s.payoutRepo.Update(ctx, existing); err != nil {
				return err
			}
		} else {

			commission := data.Total * commissionPercent / 100
			gst := commission * gstPercent / 100

			payout := &domain.PaymentPayout{
				PayoutID:              time.Now().UnixMilli(),
				ProviderID:            providerObjID,
				ServiceIDs:            data.ServiceIDs,
				BaseAmount:            utils.RoundTo2(data.Total),
				PartialAmount:         0,
				ServicePartialAmounts: make(map[string]float64),
				CommissionPercent:     commissionPercent,
				CommissionAmount:      utils.RoundTo2(commission),
				GSTPercent:            gstPercent,
				GSTAmount:             utils.RoundTo2(gst),
				NetPayable:            utils.RoundTo2(data.Total - commission - gst),
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
		filter["status"] = status
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
	log.Println("dkcnjnsajk", payouts)
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

		// Get partial amount for THIS specific booking only
		partialAmount := 0.0
		hasPartialAmount := false
		if payout.ServicePartialAmounts != nil {
			if amt, exists := payout.ServicePartialAmounts[serviceID.Hex()]; exists {
				partialAmount = amt
				hasPartialAmount = true
			}
		}

		var serviceCommission, serviceGST, serviceNet float64

		if hasPartialAmount {
			serviceCommission = partialAmount * payout.CommissionPercent / 100
			serviceGST = serviceCommission * payout.GSTPercent / 100
			serviceNet = partialAmount - serviceCommission - serviceGST
		} else {
			serviceCommission = service.FinalPrice * payout.CommissionPercent / 100
			serviceGST = serviceCommission * payout.GSTPercent / 100
			serviceNet = service.FinalPrice - serviceCommission - serviceGST
			partialAmount = 0
		}

		showComplaintAdjustment := false
		if service.HasComplaintAdjustment && service.IsSettled {
			showComplaintAdjustment = true
		}

		serviceData := map[string]any{
			"id":                        service.ID,
			"booking_id":                fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":                    "amc",
			"provider_id":               payout.ProviderID.Hex(),
			"service_amount":            utils.RoundTo2(service.FinalPrice),
			"commission_percent":        payout.CommissionPercent,
			"commission_amount":         utils.RoundTo2(serviceCommission),
			"gst_percent":               payout.GSTPercent,
			"gst_amount":                utils.RoundTo2(serviceGST),
			"net_amount":                utils.RoundTo2(serviceNet),
			"partial_amount":            utils.RoundTo2(partialAmount),
			"payout_id":                 fmt.Sprintf("SET%d", payout.PayoutID),
			"is_settled":                service.IsSettled,
			"settlement_id":             service.SettlementID,
			"settled_at":                service.SettledAt,
			"has_complaint_adjustment":  service.HasComplaintAdjustment,
			"pending_settlement":        utils.RoundTo2(service.PendingDeductionAmount),
			"show_complaint_adjustment": showComplaintAdjustment,
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
					"settlement_id": fmt.Sprintf("UP%06d", settlement.SettlementID),
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

	existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerObjID)
	if err != nil {
		return err
	}

	baseAmount := req.Amount
	partialAmount := req.PartialAmount

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
			existing.BaseAmount += baseAmount

			if partialAmount > 0 {
				existing.ServicePartialAmounts[serviceIDs[0].Hex()] = partialAmount
			}
		} else if bookingExists && len(serviceIDs) > 0 {
			if partialAmount > 0 {
				existing.ServicePartialAmounts[serviceIDs[0].Hex()] = partialAmount
			}
		}

		totalEffectiveAmount := 0.0

		for _, serviceID := range existing.ServiceIDs {
			service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
			if err != nil {
				log.Printf("Warning: Cannot fetch service %s for calculation: %v", serviceID.Hex(), err)
				continue
			}

			if partialAmt, exists := existing.ServicePartialAmounts[serviceID.Hex()]; exists && partialAmt > 0 {
				totalEffectiveAmount += partialAmt
			} else {
				totalEffectiveAmount += service.FinalPrice
			}
		}

		totalPartialAmount := 0.0
		for _, amt := range existing.ServicePartialAmounts {
			totalPartialAmount += amt
		}
		existing.PartialAmount = totalPartialAmount

		commission := totalEffectiveAmount * commissionPercent / 100
		gst := commission * gstPercent / 100

		existing.CommissionAmount = utils.RoundTo2(commission)
		existing.GSTAmount = utils.RoundTo2(gst)
		existing.NetPayable = utils.RoundTo2(totalEffectiveAmount - commission - gst)
		existing.UpdatedAt = time.Now()

		log.Printf("ProcessPayout - Updated existing payout: TotalEffective=%.2f, Commission=%.2f, GST=%.2f, NetPayable=%.2f",
			totalEffectiveAmount, commission, gst, existing.NetPayable)

		return s.payoutRepo.Update(ctx, existing)
	}

	servicePartialAmounts := make(map[string]float64)
	if partialAmount > 0 && len(serviceIDs) > 0 {
		servicePartialAmounts[serviceIDs[0].Hex()] = partialAmount
	}

	effectiveAmount := baseAmount
	if partialAmount > 0 {
		effectiveAmount = partialAmount
	}

	commission := effectiveAmount * commissionPercent / 100
	gst := commission * gstPercent / 100

	complaintObjID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerObjID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintObjID,
		ComplaintInternalID:   &req.ComplaintInternalID,
		BaseAmount:            baseAmount,
		PartialAmount:         partialAmount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      utils.RoundTo2(commission),
		GSTPercent:            gstPercent,
		GSTAmount:             utils.RoundTo2(gst),
		NetPayable:            utils.RoundTo2(effectiveAmount - commission - gst),
		PayoutType:            domain.PayoutTypeComplaint,
		IsDeduction:           true,
		Status:                domain.PayoutStatusPending,
		PeriodFrom:            time.Now().Add(-24 * time.Hour),
		PeriodTo:              time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Remarks:               req.Reason,
	}

	log.Printf("ProcessPayout - Created new payout: Effective=%.2f, Commission=%.2f, GST=%.2f, NetPayable=%.2f",
		effectiveAmount, commission, gst, payout.NetPayable)

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
