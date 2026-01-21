package service

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/utils"
	"provider_management/internal/dto"
	"provider_management/internal/domain"
	"provider_management/internal/constants"
	"provider_management/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayoutService struct {
	serviceRepo    *repository.AcceptedServiceRepo
	payoutRepo     *repository.PaymentPayoutRepo
	providerRepo   *repository.ProviderRepo
	settlementRepo *repository.ProviderSettlementRepo
}

func NewPayoutService(serviceRepo *repository.AcceptedServiceRepo, payoutRepo *repository.PaymentPayoutRepo, providerRepo *repository.ProviderRepo, settlementRepo *repository.ProviderSettlementRepo) *PayoutService {
	return &PayoutService{
		serviceRepo:    serviceRepo,
		payoutRepo:     payoutRepo,
		providerRepo:   providerRepo,
		settlementRepo: settlementRepo,
	}
}

type providerBucket struct {
	ServiceIDs []primitive.ObjectID
	Total      float64
}

func (s *PayoutService) CreatePayoutLast6Hours(ctx context.Context) error {

	to := time.Now()
	from := to.Add(-6 * time.Hour)

	services, err := s.serviceRepo.FindCompletedPaidBetween(ctx, from, to)
	if err != nil {
		return err
	}

	group := make(map[primitive.ObjectID]*providerBucket)

	for _, svc := range services {
		if svc.IsPayoutCancelled {
			continue
		}

		bucket, ok := group[svc.ProviderID]
		if !ok {
			bucket = &providerBucket{}
			group[svc.ProviderID] = bucket
		}

		bucket.ServiceIDs = append(bucket.ServiceIDs, svc.ID)
		bucket.Total += svc.FinalPrice
	}

	for providerID, bucket := range group {

		existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerID)
		if err != nil {
			return err
		}

		if existing != nil {
			s.mergeIntoExistingPayoutWithoutComplaint(ctx, existing, bucket)
			continue
		}

		if err := s.createNewPayoutWithoutComplaint(ctx, providerID, bucket, from, to); err != nil {
			return err
		}

		if err := s.serviceRepo.MarkPayoutCreated(ctx, bucket.ServiceIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *PayoutService) calculateEffectiveAmount( ctx context.Context, payout *domain.PaymentPayout ) float64 {

	serviceIDs := []primitive.ObjectID{}

	for _, id := range payout.ServiceIDs {
		if _, ok := payout.ServicePartialAmounts[id.Hex()]; !ok {
			serviceIDs = append(serviceIDs, id)
		}
	}

	services, err := s.serviceRepo.FindByIDs(ctx, serviceIDs)
	if err != nil {
		return payout.BaseAmount
	}

	total := 0.0

	for _, svc := range services {
		if !svc.IsPayoutCancelled {
			total += svc.FinalPrice
		}
	}

	for _, amt := range payout.ServicePartialAmounts {
		total += amt
	}

	return utils.RoundTo2(total)
}

func (s *PayoutService) GetProviderPayouts( ctx context.Context, filters dto.PayoutFilters, sort dto.PayoutSort, pagination dto.PaginationParams) (*dto.PayoutListResponse, error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 20
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	payouts, total, err := s.payoutRepo.GetProviderPayouts(ctx, filters, sort, pagination)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	providerMap := make(map[primitive.ObjectID]string)
	for _, pr := range providers {
		providerMap[pr.ID] = pr.Name
	}

	responseData := make([]dto.PayoutResponse,0, len(payouts))

	for _, p := range payouts {
		resp := dto.PayoutResponse{
			ID:                p.ID.Hex(),
			PayoutID:          "PAY" + strconv.FormatInt(p.PayoutID, 10),
			ProviderID:        p.ProviderID.Hex(),
			ProviderName:      providerMap[p.ProviderID],
			ServiceIDs:        p.ServiceIDs,
			BaseAmount:        utils.RoundTo2(p.BaseAmount),
			CommissionPercent: p.CommissionPercent,
			CommissionAmount:  utils.RoundTo2(p.CommissionAmount),
			GSTPercent:        p.GSTPercent,
			GSTAmount:         utils.RoundTo2(p.GSTAmount),
			NetPayable:        utils.RoundTo2(p.NetPayable),
			Status:            string(p.Status),
			PeriodFrom:        p.PeriodFrom,
			PeriodTo:          p.PeriodTo,
			CreatedAt:         p.CreatedAt,
			UpdatedAt:         p.UpdatedAt,
		}
		responseData = append(responseData, resp)
	}

	totalPages := (total + int64(pagination.Limit) - 1) / int64(pagination.Limit)

	return &dto.PayoutListResponse{
		Data: responseData,
		Pagination: dto.PaginationMeta{
			Page:        pagination.Page,
			Limit:       pagination.Limit,
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     pagination.Page < totalPages,
			HasPrevious: pagination.Page > 1,
		},
	}, nil
}


func (s *PayoutService) GetPayoutServices(ctx context.Context, payoutID string) (*dto.PayoutServiceListResponse, error) {

	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "PAY") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)

	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filters := dto.PayoutFilters{PayoutIDInt : payoutIDInt}
	sort := dto.PayoutSort{SortBy: "createdAt", SortOrder: "-1"}
	pagination := dto.PaginationParams{Page: 1, Limit: 1000}

	payouts, _, err := s.payoutRepo.GetProviderPayouts(ctx, filters, sort, pagination)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}

	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	data := make([]dto.PayoutServiceResponse, 0, len(payout.ServiceIDs))

	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
		
		if err != nil {
			fmt.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		isInSettlement := service.SettlementStatus == domain.SettleStatusPending || service.SettlementStatus == domain.SettleStatusSettled

		if isInSettlement && !service.HasComplaintAdjustment {
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
		if service.HasComplaintAdjustment && isInSettlement && !service.IsSettledAfterComplaint {
			showComplaintAdjustment = true
		}

		resp := dto.PayoutServiceResponse{
			ID:                      service.ID.Hex(),
			BookingID:               fmt.Sprintf("BK%d", service.InternalID),
			AMCID:                   "amc",
			ProviderID:              payout.ProviderID.Hex(),
			ServiceAmount:           utils.RoundTo2(service.FinalPrice),
			CommissionPercent:       payout.CommissionPercent,
			CommissionAmount:        utils.RoundTo2(serviceCommission),
			GSTPercent:              payout.GSTPercent,
			GSTAmount:               utils.RoundTo2(serviceGST),
			NetAmount:               utils.RoundTo2(serviceNet),
			PartialAmount:           utils.RoundTo2(partialAmount),
			PayoutID:                fmt.Sprintf("SET%d", payout.PayoutID),
			SettlementStatus:        string(service.SettlementStatus),
			SettlementID:            service.SettlementID,
			SettledAt:               service.SettledAt,
			HasComplaintAdjustment:  service.HasComplaintAdjustment,
			PendingSettlement:       utils.RoundTo2(service.PendingDeductionAmount),
			ShowComplaintAdjustment: showComplaintAdjustment,
			PayoutStatus:            service.PayoutStatus,
			IsPayoutCancelled:       service.IsPayoutCancelled,
			IsSettledAfterComplaint: service.IsSettledAfterComplaint,
		}

		data = append(data, resp)
	}

	return &dto.PayoutServiceListResponse{
		Data: data,
	}, nil
}

func (s *PayoutService) GetProviderPayoutDetails( ctx context.Context, payoutID string ) (*dto.ProviderPayoutDetailsResponse, error) {

	payoutIDInt, err := parseProviderPayoutID(payoutID)
	if err != nil {
		return nil, err
	}

	payout, err := s.getPayoutByID(ctx, payoutIDInt)
	if err != nil {
		return nil, err
	}

	provider, err := s.providerRepo.FindByID(ctx, payout.ProviderID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provider: %w", err)
	}

	return &dto.ProviderPayoutDetailsResponse{
		ProviderDetails:   getProviderDetails(provider, payout),
		AccountDetails:    getAccountDetails(provider),
		EarningSummary:    s.calculateEarnings(ctx, payout.ProviderID.Hex()),
		SettlementHistory: s.getSettlementHistory(ctx, payout.ProviderID),
	}, nil
}

func (s *PayoutService) getPayoutByID(ctx context.Context, payoutID int64) (*domain.PaymentPayout, error) {

	filters := dto.PayoutFilters{
		Search: strconv.FormatInt(payoutID, 10),
	}

	sort := dto.PayoutSort{
		SortBy:    "createdAt",
		SortOrder: "desc",
	}

	pagination := dto.PaginationParams{Page: 1, Limit: 1}

	payouts, _, err := s.payoutRepo.GetProviderPayouts(ctx, filters, sort, pagination)
	if err != nil {
		return nil, err
	}
	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return &payouts[0], nil
}

func getProviderDetails(provider *domain.Provider, payout *domain.PaymentPayout) map[string]any {
	return map[string]any{
		"name":          provider.Name,
		"phone_number":  provider.Phone,
		"provider_id":   payout.ProviderID.Hex(),
		"email_id":      provider.Email,
		"mechanic_type": strings.Join(provider.VehicleType, ", "),
		"vehicle_brand": strings.Join(constants.GetBrandNamesByIDs(provider.ProviderBrands), ", "),
		"zone":          provider.City,
		"join_date":     provider.CreatedAt,
		"approval_date": provider.UpdatedAt,
		"service_type":  strings.Join(constants.GetServiceNamesByIDs(provider.ProviderServices), ", "),
	}
}

func getAccountDetails(provider *domain.Provider) map[string]any {
	details := map[string]any{
		"account_holder_name": "",
		"branch_name":         "",
		"ifsc_code":           "",
		"upi_id":              "",
		"gst_number":          provider.GSTNumber,
		"account_verified":    provider.Status,
	}

	if provider.BankDetails != nil {
		details["account_holder_name"] = provider.BankDetails.AccountHolderName
		details["branch_name"] = provider.BankDetails.BranchName
		details["ifsc_code"] = provider.BankDetails.IfscCode
		details["upi_id"] = provider.BankDetails.Upi
	}

	return details
}

func (s *PayoutService) calculateEarnings( ctx context.Context, providerID string ) map[string]any {

	var totalEarnings, totalSettled, pendingSettlement, totalGST float64

	filters := dto.PayoutFilters{ProviderID: providerID}
	sort := dto.PayoutSort{SortBy: "createdAt", SortOrder: "-1"}
	pagination := dto.PaginationParams{Page: 1, Limit: 1000}

	payouts, _, err := s.payoutRepo.GetProviderPayouts(ctx, filters, sort, pagination)
	if err != nil {
		return map[string]any{}
	}

	for _, p := range payouts {
		totalEarnings += p.BaseAmount
		totalGST += p.GSTAmount

		switch p.Status {
		case domain.PayoutStatusSettled:
			totalSettled += p.NetPayable
		case domain.PayoutStatusPending:
			pendingSettlement += p.NetPayable
		}
	}

	return map[string]any{
		"total_earnings":     utils.RoundTo2(totalEarnings),
		"total_settled":      utils.RoundTo2(totalSettled),
		"pending_settlement": utils.RoundTo2(pendingSettlement),
		"gst":                utils.RoundTo2(totalGST),
		"adjustments":        0.0,
	}
}

func (s *PayoutService) getSettlementHistory( ctx context.Context, providerID primitive.ObjectID ) []map[string]any {

	if s.settlementRepo == nil {
		return []map[string]any{}
	}

	settlements, _, err := s.settlementRepo.GetSettlements(ctx, bson.M{"providerId": providerID, "status": domain.SettleStatusSettled}, 0, 100,
		"settledAt", -1,
	)
	if err != nil {
		return []map[string]any{}
	}

	history := make([]map[string]any, 0, len(settlements))
	for _, s := range settlements {
		history = append(history, map[string]any{
			"date":          s.SettledAt,
			"settlement_id": fmt.Sprintf("SET%06d", s.SettlementID),
			"amount":        utils.RoundTo2(s.TotalAmount),
			"payment_mode":  s.PaymentMode,
			"method":        s.PaymentMethod,
		})
	}
	return history
}

func (s *PayoutService) ProcessPayout(ctx context.Context, req dto.PayoutRequest) error {

	providerID, err := primitive.ObjectIDFromHex(req.ProviderID)

	if err != nil {
		return err
	}

	serviceIDs := []primitive.ObjectID{}
	if req.BookingID != "" {
		if sid, err := primitive.ObjectIDFromHex(req.BookingID); err == nil {
			serviceIDs = append(serviceIDs, sid)
		}
	}

	commissionPercent := constants.DefaultCommissionPercent
	gstPercent := constants.DefaultGSTPercent

	existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerID)
	
	if err != nil {
		fmt.Printf("ERROR: FindPendingByProvider failed: %v", err)
		return err
	}

	if existing != nil {
		return s.updateExistingProviderPayout(ctx, existing, serviceIDs, req, commissionPercent, gstPercent)
	}

	if req.CreateDeduction {
		return s.createDeductionProviderPayout(ctx, providerID, serviceIDs, req, commissionPercent, gstPercent)
	}

	return s.createNewProcessPayout(ctx, providerID, serviceIDs, req, commissionPercent, gstPercent)

}

func (s *PayoutService) mergeIntoExistingPayoutWithoutComplaint( ctx context.Context, existing *domain.PaymentPayout, bucket *providerBucket ) {

	if existing.ServicePartialAmounts == nil {
		existing.ServicePartialAmounts = make(map[string]float64)
	}

	existing.ServiceIDs = append(existing.ServiceIDs, bucket.ServiceIDs...)
	existing.BaseAmount += bucket.Total

	effective := s.calculateEffectiveAmount(ctx, existing)

	calc := utils.CalculatePayout( effective, constants.DefaultCommissionPercent, constants.DefaultGSTPercent)

	existing.CommissionAmount = calc.Commission
	existing.GSTAmount = calc.GST
	existing.NetPayable = calc.NetPayable
	existing.UpdatedAt = time.Now()

	_ = s.payoutRepo.Update(ctx, existing)
}

func (s *PayoutService) createNewPayoutWithoutComplaint( ctx context.Context, providerID primitive.ObjectID, bucket *providerBucket, from, to time.Time ) error {

	calc := utils.CalculatePayout( bucket.Total, constants.DefaultCommissionPercent, constants.DefaultGSTPercent)

	payout := &domain.PaymentPayout{
		PayoutID:          time.Now().UnixMilli(),
		ProviderID:        providerID,
		ServiceIDs:        bucket.ServiceIDs,
		BaseAmount:        calc.BaseAmount,
		CommissionPercent: constants.DefaultCommissionPercent,
		CommissionAmount:  calc.Commission,
		GSTPercent:        constants.DefaultGSTPercent,
		GSTAmount:         calc.GST,
		NetPayable:        calc.NetPayable,
		Status:            domain.PayoutStatusPending,
		PayoutType:        domain.PayoutTypeRegular,
		PeriodFrom:        from,
		PeriodTo:          to,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return s.payoutRepo.Create(ctx, payout)
}

func (s *PayoutService) updateExistingProviderPayout( ctx context.Context, existing *domain.PaymentPayout, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

	if existing.ServicePartialAmounts == nil {
		existing.ServicePartialAmounts = make(map[string]float64)
	}

	for _, sid := range serviceIDs {
		key := sid.Hex()
		if _, exists := existing.ServicePartialAmounts[key]; !exists {
			existing.ServiceIDs = append(existing.ServiceIDs, sid)
			if req.PartialAmount > 0 {
				existing.ServicePartialAmounts[key] = req.PartialAmount
			}
			if req.Amount > 0 {
				existing.BaseAmount += req.Amount
			}
		}
	}

	for _, sid := range serviceIDs {
		if req.CancelPayout {
			if err := s.cancelServiceProviderPayout(ctx, sid.Hex()); err != nil {
				fmt.Printf("ERROR: Failed to cancel service: %v", err)
			}
			existing.ServicePartialAmounts[sid.Hex()] = 0
		}
	}

	effective := s.calculateEffectiveAmount(ctx, existing)
	calc := utils.CalculatePayout(effective, commissionPercent, gstPercent)

	existing.CommissionAmount = calc.Commission
	existing.GSTAmount = calc.GST
	existing.NetPayable = calc.NetPayable
	existing.UpdatedAt = time.Now()

	return s.payoutRepo.Update(ctx, existing)
}

func (s *PayoutService) cancelServiceProviderPayout(ctx context.Context, serviceID string) error {

	service, err := s.serviceRepo.FindByID(ctx, serviceID)
	if err != nil {
		return err
	}

	if service.SettlementStatus == domain.SettleStatusPending || service.SettlementStatus == domain.SettleStatusSettled {
		return nil
	}

	return s.serviceRepo.UpdatePayoutCancellation(ctx, serviceID, true)
}

func (s *PayoutService) createDeductionProviderPayout( ctx context.Context, providerID primitive.ObjectID, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

	effective := req.PartialAmount
	if effective == 0 {
		effective = req.Amount
	}

	calc := utils.CalculatePayout(effective, commissionPercent, gstPercent)
	complaintID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	servicePartialAmounts := make(map[string]float64)
	for _, sid := range serviceIDs {
		servicePartialAmounts[sid.Hex()] = req.Amount
	}

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintID,
		ComplaintInternalID:   &req.ComplaintInternalID,
		BaseAmount:            -req.Amount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      -calc.Commission,
		GSTPercent:            gstPercent,
		GSTAmount:             -calc.GST,
		NetPayable:            -calc.NetPayable,
		IsDeduction:           true,
		Status:                domain.PayoutStatusPending,
		PayoutType:            domain.PayoutTypeComplaint,
		PeriodFrom:            time.Now().Add(-24 * time.Hour),
		PeriodTo:              time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Remarks:               req.Reason,
	}

	return s.payoutRepo.Create(ctx, payout)
}

func (s *PayoutService) createNewProcessPayout( ctx context.Context, providerID primitive.ObjectID, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

	servicePartialAmounts := make(map[string]float64)
	effective := req.Amount

	if req.CancelPayout {
		for _, sid := range serviceIDs {
			if err := s.cancelServiceProviderPayout(ctx, sid.Hex()); err != nil {
				fmt.Printf("ERROR: cancel service failed: %v", err)
			}
			servicePartialAmounts[sid.Hex()] = 0
		}
		effective = 0
	} else if req.PartialAmount > 0 {
		for _, sid := range serviceIDs {
			servicePartialAmounts[sid.Hex()] = req.PartialAmount
		}
		effective = req.PartialAmount
	} else if len(serviceIDs) > 0 && effective == 0 {
		service, err := s.serviceRepo.FindByID(ctx, serviceIDs[0].Hex())
		if err == nil {
			effective = service.FinalPrice
		}
	}

	calc := utils.CalculatePayout(effective, commissionPercent, gstPercent)
	complaintID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintID,
		ComplaintInternalID:   &req.ComplaintInternalID,
		BaseAmount:            req.Amount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      calc.Commission,
		GSTPercent:            gstPercent,
		GSTAmount:             calc.GST,
		NetPayable:            calc.NetPayable,
		IsPayoutCancelled:     req.CancelPayout,
		Status:                domain.PayoutStatusPending,
		PayoutType:            domain.PayoutTypeComplaint,
		PeriodFrom:            time.Now().Add(-24 * time.Hour),
		PeriodTo:              time.Now(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Remarks:               req.Reason,
	}

	return s.payoutRepo.Create(ctx, payout)
}

func parseProviderPayoutID(payoutIDStr string) (int64, error) {
	numStr := strings.TrimPrefix(payoutIDStr, "PAY")
	return strconv.ParseInt(numStr, 10, 64)
}