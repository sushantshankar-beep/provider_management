package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/constants"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PayoutService struct {
	serviceRepo     *repository.AcceptedServiceRepo
	payoutRepo      *repository.PaymentPayoutRepo
	providerRepo    *repository.ProviderRepo
	settlementRepo  *repository.ProviderSettlementRepo
	kycRepo         *repository.ProviderKYCRepository
	transactionRepo *repository.TransactionRepo
}

func NewPayoutService(serviceRepo *repository.AcceptedServiceRepo, payoutRepo *repository.PaymentPayoutRepo, providerRepo *repository.ProviderRepo, settlementRepo *repository.ProviderSettlementRepo, kycRepo *repository.ProviderKYCRepository, transactionRepo *repository.TransactionRepo) *PayoutService {
	return &PayoutService{
		serviceRepo:     serviceRepo,
		payoutRepo:      payoutRepo,
		providerRepo:    providerRepo,
		settlementRepo:  settlementRepo,
		kycRepo:         kycRepo,
		transactionRepo: transactionRepo,
	}
}

type providerBucket struct {
	ServiceIDs       []primitive.ObjectID
	TotalTransAmount float64
	Total            float64
}

func (s *PayoutService) CreatePayoutLast6Hours(ctx context.Context, hours int) error {

	to := time.Now()
	from := to.Add(-time.Duration(hours) * time.Hour)

	services, err := s.serviceRepo.FindCompletedPaidBetween(ctx, from, to)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return nil
	}

	group := make(map[primitive.ObjectID]*providerBucket)

	for _, svc := range services {
		if svc.IsPayoutCancelled {
			continue
		}

		transaction, err := s.transactionRepo.FindByServiceID(ctx, svc.ID.Hex())
		if err != nil {
			log.Printf("Error fetching transaction for service %v: %v", svc.ID, err)
			continue
		}

		bucket, ok := group[svc.Provider]
		if !ok {
			bucket = &providerBucket{}
			group[svc.Provider] = bucket
		}

		bucket.ServiceIDs = append(bucket.ServiceIDs, svc.ID)
		bucket.Total += svc.FinalPrice
		bucket.TotalTransAmount += transaction.Amount
	}

	for providerID, bucket := range group {
		if err := s.serviceRepo.MarkPayoutCreated(ctx, bucket.ServiceIDs); err != nil {
			log.Printf("Error marking services for provider %v: %v", providerID, err)
			return err
		}

		existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerID)
		if err != nil {
			return err
		}

		if existing != nil {
			if err := s.mergeIntoExistingPayoutWithoutComplaint(ctx, existing, bucket, providerID); err != nil {
				return err
			}
			continue
		}

		if err := s.createNewPayoutWithoutComplaint(ctx, providerID, bucket, from, to); err != nil {
			return err
		}
	}

	return nil
}

func (s *PayoutService) calculateEffectiveAmount(
	ctx context.Context,
	payout *domain.PaymentPayout,
) float64 {

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

func (s *PayoutService) GetProviderPayouts(ctx context.Context, filters dto.PayoutFilters, sort dto.PayoutSort, pagination dto.PaginationParams) (*dto.PayoutListResponse, error) {
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

	responseData := make([]dto.PayoutResponse, 0, len(payouts))

	for _, p := range payouts {
		resp := dto.PayoutResponse{
			ID:                 p.ID.Hex(),
			PayoutID:           "PAY" + strconv.FormatInt(p.PayoutID, 10),
			ProviderID:         p.ProviderID.Hex(),
			ProviderName:       providerMap[p.ProviderID],
			ServiceIDs:         p.ServiceIDs,
			TotalPayAmount:     utils.RoundTo2(p.TotalPayAmount),
			BaseAmount:         utils.RoundTo2(p.BaseAmount),
			CommissionPercent:  p.CommissionPercent,
			CommissionAmount:   utils.RoundTo2(p.CommissionAmount),
			GSTPercent:         p.GSTPercent,
			GSTAmount:          utils.RoundTo2(p.GSTAmount),
			TDSPercent:         p.TDSPercent,
			TDSAmount:          utils.RoundTo2(p.TDSAmount),
			NetPayable:         utils.RoundTo2(p.NetPayable),
			VahanwireGSTAmount: utils.RoundTo2(p.VahanwireGSTAmount),
			Status:             string(p.Status),
			PeriodFrom:         p.PeriodFrom,
			PeriodTo:           p.PeriodTo,
			CreatedAt:          p.CreatedAt,
			UpdatedAt:          p.UpdatedAt,
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

	filters := dto.PayoutFilters{PayoutIDInt: payoutIDInt}
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

	provider, err := s.providerRepo.FindByID(ctx, payout.ProviderID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provider: %v", err)
	}

	providerCommissionPercent := payout.CommissionPercent
	if provider.CommissionPercentage > 0 {
		providerCommissionPercent = provider.CommissionPercentage
	}

	kyc, _ := s.kycRepo.FindByProviderID(ctx, payout.ProviderID)
	hasGSTNumber := false
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		hasGSTNumber = true
	}

	data := make([]dto.PayoutServiceResponse, 0, len(payout.ServiceIDs))
	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())

		if err != nil {
			fmt.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		transaction, err := s.transactionRepo.FindByServiceID(ctx, serviceID.Hex())
		if err != nil {
			fmt.Printf("Error fetching transaction for service %s: %v", serviceID.Hex(), err)
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

		var baseAmount float64

		if isDeduction {
			baseAmount = service.FinalPrice
			partialAmount = 0
		} else if hasPartialAmount {
			baseAmount = partialAmount
		} else {
			baseAmount = service.FinalPrice
			partialAmount = 0
		}

		var serviceCommission, serviceTDS, serviceGST, vahanwireGSTAmount, serviceNet float64
		var tdsPercent, gstPercent float64

		isComplaintAfterSettlement := service.PayoutStatus == domain.PayoutStatusComplaintAfterSettlement && isInSettlement

		if isComplaintAfterSettlement {
			serviceCommission = 0
			serviceTDS = 0
			serviceGST = 0
			vahanwireGSTAmount = 0
			serviceNet = partialAmount
			tdsPercent = 0
			gstPercent = 0
		} else if hasGSTNumber {
			tdsPercent = constants.DefaultTDSPercent
			gstPercent = constants.DefaultGSTPercent
			serviceCommission = baseAmount * (providerCommissionPercent / 100)
			afterCommission := baseAmount - serviceCommission
			serviceGST = baseAmount * 0.18
			amountWithGST := afterCommission + serviceGST
			serviceTDS = amountWithGST * 0.10
			serviceNet = amountWithGST - serviceTDS
			vahanwireGSTAmount = 0.0
		} else {
			tdsPercent = 0.0
			gstPercent = 0.0
			serviceCommission = baseAmount * (providerCommissionPercent / 100)
			vahanwireGSTAmount = baseAmount * 0.18
			serviceTDS = 0
			serviceGST = 0
			serviceNet = baseAmount - serviceCommission
		}

		showComplaintAdjustment := false
		if service.HasComplaintAdjustment && isInSettlement && !service.IsSettledAfterComplaint {
			showComplaintAdjustment = true
		}

		complaints := make([]dto.ComplaintBookingInfo, 0, 2)

		if strings.TrimSpace(service.ComplaintUserID) != "" {
			complaints = append(complaints, dto.ComplaintBookingInfo{
				ID:       service.ComplaintUserID,
				RaisedBy: "user",
			})
		}

		if strings.TrimSpace(service.ComplaintProviderID) != "" {
			complaints = append(complaints, dto.ComplaintBookingInfo{
				ID:       service.ComplaintProviderID,
				RaisedBy: "provider",
			})
		}

		resp := dto.PayoutServiceResponse{
			ID:                      service.ID.Hex(),
			BookingID:               service.ServiceNumber,
			AMCID:                   "amc",
			ProviderID:              payout.ProviderID.Hex(),
			ProviderCode:            provider.ProviderCode,
			ServiceAmount:           utils.RoundTo2(service.FinalPrice),
			TotalPaidAmount:         utils.RoundTo2(transaction.Amount),
			CommissionPercent:       providerCommissionPercent,
			CommissionAmount:        utils.RoundTo2(serviceCommission),
			TDSPercent:              tdsPercent,
			TDSAmount:               utils.RoundTo2(serviceTDS),
			GSTPercent:              gstPercent,
			GSTAmount:               utils.RoundTo2(serviceGST),
			VahanwireGSTAmount:      utils.RoundTo2(vahanwireGSTAmount),
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
			Complaints:              complaints,
		}

		data = append(data, resp)
	}

	return &dto.PayoutServiceListResponse{
		Data: data,
	}, nil
}

func (s *PayoutService) GetProviderPayoutDetails(ctx context.Context, payoutID string) (*dto.ProviderPayoutDetailsResponse, error) {

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

	kycDeatils, err := s.kycRepo.FindByProviderID(ctx, payout.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kyc: %w", err)
	}

	return &dto.ProviderPayoutDetailsResponse{
		ProviderDetails:   getProviderDetails(provider, payout),
		AccountDetails:    getAccountDetails(kycDeatils),
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
		"provider_code": provider.ProviderCode,
		"email_id":      provider.Email,
		"mechanic_type": strings.Join(provider.VehicleType, ", "),
		"vehicle_brand": provider.ProviderBrands,
		"zone":          provider.City,
		"join_date":     provider.CreatedAt,
		"approval_date": provider.UpdatedAt,
		"service_type":  provider.ProviderServices,
	}
}

func getAccountDetails(provider *domain.ProviderKYC) map[string]any {
	details := map[string]any{
		"account_holder_name": "",
		"branch_name":         "",
		"ifsc_code":           "",
		"upi_id":              "",
		"gst_number":          provider.Bank.GSTNumber,
		"account_verified":    provider.Status,
	}

	if (provider.Bank != domain.ProviderBankDetails{}) {
		details["account_holder_name"] = provider.Bank.AccountHolderName
		details["branch_name"] = provider.Bank.BranchName
		details["ifsc_code"] = provider.Bank.IFSC
		details["upi_id"] = provider.Bank.UPIID
	}

	return details
}

func (s *PayoutService) calculateEarnings(ctx context.Context, providerID string) map[string]any {

	var totalEarnings, totalSettled, pendingSettlement, totalGST float64

	filters := dto.PayoutFilters{ProviderID: providerID}
	sort := dto.PayoutSort{SortBy: "createdAt", SortOrder: "-1"}
	pagination := dto.PaginationParams{Page: 1, Limit: 1000}

	payouts, _, err := s.payoutRepo.GetProviderPayouts(ctx, filters, sort, pagination)
	if err != nil {
		return map[string]any{}
	}

	for _, p := range payouts {
		totalEarnings += p.TotalPayAmount
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

func (s *PayoutService) getSettlementHistory(ctx context.Context, providerID primitive.ObjectID) []map[string]any {

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

	provider, err := s.providerRepo.FindByID(ctx, providerID.Hex())
	if err != nil {
		fmt.Printf("ERROR: Failed to fetch provider: %v", err)
		return err
	}

	commissionPercent := constants.DefaultCommissionPercent
	if provider.CommissionPercentage > 0 {
		commissionPercent = provider.CommissionPercentage
	}

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

func (s *PayoutService) mergeIntoExistingPayoutWithoutComplaint(ctx context.Context, existing *domain.PaymentPayout, bucket *providerBucket, providerID primitive.ObjectID) error {

	if existing.ServicePartialAmounts == nil {
		existing.ServicePartialAmounts = make(map[string]float64)
	}

	existing.ServiceIDs = append(existing.ServiceIDs, bucket.ServiceIDs...)
	existing.TotalPayAmount += bucket.TotalTransAmount
	existing.BaseAmount += bucket.Total

	effective := s.calculateEffectiveAmount(ctx, existing)

	provider, err := s.providerRepo.FindByID(ctx, providerID.Hex())
	if err != nil {
		log.Printf("Error fetching provider: %v", err)
		return err
	}

	commissionPercent := constants.DefaultCommissionPercent
	if provider.CommissionPercentage > 0 {
		commissionPercent = provider.CommissionPercentage
	}

	tdsPercent := 0.0
	finalGSTPercent := 0.0
	vahanwireGSTAmount := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	} else {
		vahanwireGSTAmount = effective * 0.18
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)

	existing.CommissionPercent = commissionPercent
	existing.CommissionAmount = calc.Commission
	existing.GSTPercent = finalGSTPercent
	existing.GSTAmount = calc.GST
	existing.VahanwireGSTAmount = vahanwireGSTAmount
	existing.TDSPercent = tdsPercent
	existing.TDSAmount = calc.TDS
	existing.NetPayable = calc.NetPayable
	existing.UpdatedAt = time.Now()

	return s.payoutRepo.Update(ctx, existing)
}

func (s *PayoutService) createNewPayoutWithoutComplaint(ctx context.Context, providerID primitive.ObjectID, bucket *providerBucket, from, to time.Time) error {

	provider, err := s.providerRepo.FindByID(ctx, providerID.Hex())
	if err != nil {
		log.Printf("Error fetching provider: %v", err)
		return err
	}

	commissionPercent := constants.DefaultCommissionPercent
	if provider.CommissionPercentage > 0 {
		commissionPercent = provider.CommissionPercentage
	}

	tdsPercent := 0.0
	finalGSTPercent := 0.0
	vahanwireGSTAmount := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	} else {
		vahanwireGSTAmount = bucket.Total * 0.18
	}

	calc := utils.CalculatePayout(bucket.Total, commissionPercent, finalGSTPercent,
		tdsPercent)

	payout := &domain.PaymentPayout{
		PayoutID:           time.Now().UnixMilli(),
		ProviderID:         providerID,
		ServiceIDs:         bucket.ServiceIDs,
		TotalPayAmount:     bucket.TotalTransAmount,
		BaseAmount:         bucket.Total,
		CommissionPercent:  commissionPercent,
		VahanwireGSTAmount: vahanwireGSTAmount,
		CommissionAmount:   calc.Commission,
		GSTPercent:         constants.DefaultGSTPercent,
		GSTAmount:          calc.GST,
		NetPayable:         calc.NetPayable,
		TDSPercent:         constants.DefaultTDSPercent,
		TDSAmount:          calc.TDS,
		Status:             domain.PayoutStatusPending,
		PayoutType:         domain.PayoutTypeRegular,
		PeriodFrom:         from,
		PeriodTo:           to,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	return s.payoutRepo.Create(ctx, payout)
}

func (s *PayoutService) updateExistingProviderPayout(ctx context.Context, existing *domain.PaymentPayout, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64) error {

	if existing.ServicePartialAmounts == nil {
		existing.ServicePartialAmounts = make(map[string]float64)
	}

	serviceExists := make(map[string]bool)
	for _, sid := range existing.ServiceIDs {
		serviceExists[sid.Hex()] = true
	}

	for _, sid := range serviceIDs {
		key := sid.Hex()

		if !serviceExists[key] {

			service, err := s.serviceRepo.FindByID(ctx, key)
			if err != nil {
				fmt.Printf("ERROR: Failed to fetch service: %v", err)
				continue
			}

			transaction, err := s.transactionRepo.FindByServiceID(ctx, key)
			if err != nil {
				fmt.Printf("ERROR: Failed to fetch transaction: %v", err)
				continue
			}

			existing.ServiceIDs = append(existing.ServiceIDs, sid)
			serviceExists[key] = true

			if req.PartialAmount > 0 {
				existing.ServicePartialAmounts[key] = req.PartialAmount
				existing.BaseAmount += req.PartialAmount
			} else {
				existing.BaseAmount += service.FinalPrice
			}

			existing.TotalPayAmount += transaction.Amount
			continue
		}

		if req.PartialAmount > 0 {

			service, err := s.serviceRepo.FindByID(ctx, key)
			if err != nil {
				fmt.Printf("ERROR: Failed to fetch service: %v", err)
				continue
			}

			delta := calculateBaseAmountDelta(
				existing,
				key,
				service.FinalPrice,
				req.PartialAmount,
			)

			existing.BaseAmount += delta
			existing.ServicePartialAmounts[key] = req.PartialAmount
		}
	}

	for _, sid := range serviceIDs {
		if req.CancelPayout {
			key := sid.Hex()

			if err := s.cancelServiceProviderPayout(ctx, key); err != nil {
				fmt.Printf("ERROR: Failed to cancel service: %v", err)
			}

			if old, ok := existing.ServicePartialAmounts[key]; ok {
				existing.BaseAmount -= old
			}

			delete(existing.ServicePartialAmounts, key)
		}
	}

	effective := s.calculateEffectiveAmount(ctx, existing)
	tdsPercent := 0.0
	finalGSTPercent := 0.0
	vahanwireGSTAmount := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
		vahanwireGSTAmount = 0.0
	} else {
		tdsPercent = 0.0
		finalGSTPercent = 0.0
		vahanwireGSTAmount = effective * 0.18
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)

	existing.CommissionPercent = commissionPercent
	existing.CommissionAmount = calc.Commission
	existing.GSTPercent = finalGSTPercent
	existing.GSTAmount = calc.GST
	existing.VahanwireGSTAmount = vahanwireGSTAmount
	existing.TDSPercent = tdsPercent
	existing.TDSAmount = calc.TDS
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

func (s *PayoutService) createDeductionProviderPayout(ctx context.Context, providerID primitive.ObjectID, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64) error {

	effective := req.PartialAmount
	if effective == 0 {
		effective = req.Amount
	}

	tdsPercent := 0.0
	finalGSTPercent := 0.0
	vahanwireGSTAmount := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	} else {
		vahanwireGSTAmount = effective * 0.18
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)
	complaintID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	servicePartialAmounts := make(map[string]float64)
	for _, sid := range serviceIDs {
		servicePartialAmounts[sid.Hex()] = req.Amount
	}

	transaction, _ := s.transactionRepo.FindByServiceID(ctx, serviceIDs[0].Hex())

	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintID,
		ComplaintInternalID:   req.ComplaintInternalID,
		TotalPayAmount:        -transaction.Amount,
		BaseAmount:            -req.Amount,
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      -calc.Commission,
		VahanwireGSTAmount:    -vahanwireGSTAmount,
		GSTPercent:            gstPercent,
		GSTAmount:             -calc.GST,
		TDSPercent:            tdsPercent,
		TDSAmount:             -calc.TDS,
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

func (s *PayoutService) createNewProcessPayout(
	ctx context.Context,
	providerID primitive.ObjectID,
	serviceIDs []primitive.ObjectID,
	req dto.PayoutRequest,
	commissionPercent, gstPercent float64,
) error {
	var totalPaidAmount float64
	var baseAmount float64

	if len(serviceIDs) > 0 {
		service, err := s.serviceRepo.FindByID(ctx, serviceIDs[0].Hex())
		if err != nil {
			return err
		}

		transaction, err := s.transactionRepo.FindByServiceID(ctx, service.ID.Hex())
		if err != nil {
			return err
		}

		// ✅ User paid amount (includes GST)
		totalPaidAmount = transaction.Amount

		// ✅ Actual service price (without GST)
		baseAmount = service.FinalPrice
	}

	// 🔥 FIX: effective amount should be based on PARTIAL or FULL decision
	effective := baseAmount
	if req.PartialAmount > 0 {
		effective = req.PartialAmount
	}

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)

	tdsPercent := 0.0
	finalGSTPercent := 0.0
	vahanwireGSTAmount := 0.0

	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	} else {
		vahanwireGSTAmount = effective * 0.18
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)

	complaintID, _ := primitive.ObjectIDFromHex(req.ComplaintID)

	payout := &domain.PaymentPayout{
		PayoutID:   time.Now().UnixMilli(),
		ProviderID: providerID,
		ServiceIDs: serviceIDs,

		// ✅ FIX: do NOT overwrite what user paid
		TotalPayAmount: utils.RoundTo2(totalPaidAmount),

		// ✅ FIX: base amount must be service price
		BaseAmount: utils.RoundTo2(baseAmount),

		ServicePartialAmounts: map[string]float64{
			serviceIDs[0].Hex(): effective,
		},

		CommissionPercent:  commissionPercent,
		CommissionAmount:   calc.Commission,
		GSTPercent:         gstPercent,
		GSTAmount:          calc.GST,
		VahanwireGSTAmount: vahanwireGSTAmount,
		TDSPercent:         tdsPercent,
		TDSAmount:          calc.TDS,
		NetPayable:         calc.NetPayable,

		IsPayoutCancelled: req.CancelPayout,
		Status:            domain.PayoutStatusPending,
		PayoutType:        domain.PayoutTypeComplaint,

		ComplaintID:         &complaintID,
		ComplaintInternalID: req.ComplaintInternalID,

		PeriodFrom: time.Now().Add(-24 * time.Hour),
		PeriodTo:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Remarks:    req.Reason,
	}

	return s.payoutRepo.Create(ctx, payout)
}

func parseProviderPayoutID(payoutIDStr string) (int64, error) {
	numStr := strings.TrimPrefix(payoutIDStr, "PAY")
	return strconv.ParseInt(numStr, 10, 64)
}

type GetSettlementStatsRequest struct {
	Tab        string `form:"tab"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	ProviderID string `form:"provider_id"`
}

type SettlementStatsResponse struct {
	TotalPayout          float64 `json:"total_payout"`
	TotalProviderRevenue float64 `json:"total_provider_revenue"`
	VahanwireCommission  float64 `json:"vahanwire_commission"`
	VahanwireGST         float64 `json:"vahanwire_gst"`
	PartnerGST           float64 `json:"partner_gst"`
	TotalTDS             float64 `json:"total_tds"`
}

func (s *PayoutService) GetPayoutStats(
	ctx context.Context,
	req *GetSettlementStatsRequest,
) (*SettlementStatsResponse, error) {

	filter := bson.M{
		"status": bson.M{
			"$in": []string{
				string(domain.PayoutStatusSettled),
				string(domain.PayoutStatusPartiallySettled),
			},
		},
	}

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
		}
		filter["createdAt"] = bson.M{"$gte": startDate}
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
		}
		endDate = endDate.Add(24 * time.Hour)

		if _, ok := filter["createdAt"]; ok {
			filter["createdAt"].(bson.M)["$lte"] = endDate
		} else {
			filter["createdAt"] = bson.M{"$lte": endDate}
		}
	}

	if req.ProviderID != "" {
		providerID, err := primitive.ObjectIDFromHex(req.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("invalid provider id")
		}
		filter["providerId"] = providerID
	}

	stats, err := s.payoutRepo.GetAggregatedStats(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get settlement stats: %v", err)
	}

	response := &SettlementStatsResponse{
		TotalPayout:          utils.RoundTo2(stats.TotalPayAmount),
		TotalProviderRevenue: utils.RoundTo2(stats.NetPayable),
		VahanwireCommission:  utils.RoundTo2(stats.CommissionAmount),
		VahanwireGST:         utils.RoundTo2(stats.VahanwireGSTAmount),
		PartnerGST:           utils.RoundTo2(stats.GSTAmount),
		TotalTDS:             utils.RoundTo2(stats.TDSAmount),
	}

	return response, nil
}

func calculateBaseAmountDelta(
	existing *domain.PaymentPayout,
	serviceID string,
	servicePrice float64,
	newPartial float64,
) float64 {

	oldPartial, hasPartial := existing.ServicePartialAmounts[serviceID]
	if hasPartial {
		return newPartial - oldPartial
	}
	return newPartial - servicePrice
}
