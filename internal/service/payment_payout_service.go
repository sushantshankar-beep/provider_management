package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/constants"
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

	group := make(map[string]*bucket)

	for _, svc := range services {
		if _, ok := group[svc.ProviderID.Hex()]; !ok {
			group[svc.ProviderID.Hex()] = &bucket{}
		}

		objID, _ := primitive.ObjectIDFromHex(svc.ID.Hex())

		group[svc.ProviderID.Hex()].ServiceIDs = append(
			group[svc.ProviderID.Hex()].ServiceIDs,
			objID,
		)

		group[svc.ProviderID.Hex()].Total += svc.FinalPrice
	}

	for providerID, data := range group {
		payoutID := time.Now().UnixMilli()
		base := data.Total
		commissionPercent := 20.0
		gstPercent := 18.0

		commission := base * commissionPercent / 100
		gst := commission * gstPercent / 100
		net := base - commission - gst

		providerObjID, _ := primitive.ObjectIDFromHex(providerID)

		payout := &domain.PaymentPayout{
			PayoutID:          payoutID,
			ProviderID:        providerObjID,
			ServiceIDs:        data.ServiceIDs,
			BaseAmount:        base,
			CommissionPercent: commissionPercent,
			CommissionAmount:  commission,
			GSTPercent:        gstPercent,
			GSTAmount:         gst,
			NetPayable:        net,
			Status:            "pending",
			PeriodFrom:        from,
			PeriodTo:          to,
			CreatedAt:         time.Now(),
		}

		if err := s.payoutRepo.Create(ctx, payout); err != nil {
			return err
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
		if objID, err := primitive.ObjectIDFromHex(search); err == nil {
			filter["providerId"] = objID
		} else {
			filter["$or"] = []bson.M{
				{"status": bson.M{"$regex": search, "$options": "i"}},
			}
		}
	}

	order := -1
	if sortOrder == "asc" {
		order = 1
	}

	if sortBy == "" {
		sortBy = "updatedAt"
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

	responseData := make([]map[string]any, len(payouts))
	for i, p := range payouts {
		responseData[i] = map[string]any{
			"id":                 p.ID.Hex(),
			"payout_id":          "PAY" + strconv.FormatInt(p.PayoutID, 10),
			"provider_id":        p.ProviderID.Hex(),
			"service_ids":        p.ServiceIDs,
			"base_amount":        p.BaseAmount,
			"commission_percent": p.CommissionPercent,
			"commission_amount":  p.CommissionAmount,
			"gst_percent":        p.GSTPercent,
			"gst_amount":         p.GSTAmount,
			"net_payable":        p.NetPayable,
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

		if service.IsSettled {
			continue
		}

		serviceCommission := service.FinalPrice * payout.CommissionPercent / 100
		serviceGST := serviceCommission * payout.GSTPercent / 100
		serviceNet := service.FinalPrice - serviceCommission - serviceGST

		serviceData := map[string]any{
			"id":                 service.ID,
			"booking_id":         fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":             "amc",
			"provider_id":        payout.ProviderID.Hex(),
			"service_amount":     service.FinalPrice,
			"commission_percent": payout.CommissionPercent,
			"commission_amount":  serviceCommission,
			"gst_percent":        payout.GSTPercent,
			"gst_amount":         serviceGST,
			"net_amount":         serviceNet,
			"partial_amount":     payout.PartialAmount,
			"payout_id":          fmt.Sprintf("SET%d", payout.PayoutID),
			"is_settled":         service.IsSettled,
			"settlement_id":      service.SettlementID,
			"settled_at":         service.SettledAt,
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
			"service_amount":     service.FinalPrice,
			"commission_percent": payout.CommissionPercent,
			"gst_percent":        payout.GSTPercent,
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
		"verified":            provider.Status,
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
		"total_earnings":     totalEarnings,
		"total_settled":      totalSettled,
		"pending_settlement": pendingSettlement,
		"gst":                totalGST,
		"adjustments":        adjustments,
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
					"amount":        settlement.TotalAmount,
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
						"amount":        p.NetPayable,
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
		return fmt.Errorf("invalid provider ID: %w", err)
	}

	serviceIDs := []primitive.ObjectID{}
	if req.BookingID != "" {
		bookingObjID, err := primitive.ObjectIDFromHex(req.BookingID)
		if err != nil {
			log.Printf("Warning: Invalid booking ID format: %v", err)
		} else {
			serviceIDs = append(serviceIDs, bookingObjID)
		}
	}
   


	baseAmount := req.Amount

	partialAmount := req.PartialAmount 

	commissionPercent := 20.0
	commissionAmount := baseAmount * (commissionPercent / 100)
	

	gstPercent := 18.0
	gstAmount := commissionAmount * (gstPercent / 100)

	netPayable := baseAmount - commissionAmount - gstAmount

	payoutID := time.Now().UnixMilli()


	payout := &domain.PaymentPayout{
		PayoutID:          payoutID,
		ProviderID:        providerObjID,
		ServiceIDs:        serviceIDs,
		BaseAmount:        baseAmount,
		CommissionPercent: commissionPercent,
		CommissionAmount:  commissionAmount,
		GSTPercent:        gstPercent,
		GSTAmount:         gstAmount,
		NetPayable:        netPayable,
		PartialAmount:     partialAmount,
		Status:            domain.PayoutStatusPending,
		PeriodFrom:        time.Now().Add(-24 * time.Hour),
		PeriodTo:          time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.payoutRepo.Create(ctx, payout); err != nil {
		
		return fmt.Errorf("failed to create payout: %w", err)
	}

	
	return nil
}
