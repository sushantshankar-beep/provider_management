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
	serviceRepo    *repository.AcceptedServiceRepo
	payoutRepo     *repository.PaymentPayoutRepo
	providerRepo   *repository.ProviderRepo
	settlementRepo *repository.ProviderSettlementRepo
	kycRepo *repository.ProviderKYCRepository
	transactionRepo *repository.TransactionRepo
}

func NewPayoutService(serviceRepo *repository.AcceptedServiceRepo, payoutRepo *repository.PaymentPayoutRepo, providerRepo *repository.ProviderRepo, settlementRepo *repository.ProviderSettlementRepo,kycRepo *repository.ProviderKYCRepository,transactionRepo *repository.TransactionRepo) *PayoutService {
	return &PayoutService{
		serviceRepo:    serviceRepo,
		payoutRepo:     payoutRepo,
		providerRepo:   providerRepo,
		settlementRepo: settlementRepo,
		kycRepo: kycRepo,
		transactionRepo: transactionRepo,
	}
}

type providerBucket struct {
	ServiceIDs []primitive.ObjectID
	TotalTransAmount  float64 
	Total      float64
}

func (s *PayoutService) CreatePayoutLast6Hours(ctx context.Context) error {

	to := time.Now()
	from := to.Add(-6 * time.Hour)

	services, err := s.serviceRepo.FindCompletedPaidBetween(ctx, from, to)
	log.Println("Found services:", len(services))
	if err != nil {
		return err
	}

	if len(services) == 0 {
		log.Println("No services to process for payout")
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

		log.Println("Provider:", svc.Provider, "Service:", svc.ID, "Amount:", transaction.Amount, "Total:", bucket.Total)
	}

	for providerID, bucket := range group {
		log.Println("Processing provider:", providerID, "Services count:", len(bucket.ServiceIDs), "Total:", bucket.Total)

		if err := s.serviceRepo.MarkPayoutCreated(ctx, bucket.ServiceIDs); err != nil {
			log.Printf("Error marking services for provider %v: %v", providerID, err)
			return err
		}

		existing, err := s.payoutRepo.FindPendingByProvider(ctx, providerID)
		if err != nil {
			return err
		}

		if existing != nil {
			log.Println("Merging into existing payout:", existing.PayoutID)
			if err := s.mergeIntoExistingPayoutWithoutComplaint(ctx, existing, bucket,providerID); err != nil {
				return err
			}
			continue
		}

		log.Println("Creating new payout for provider:", providerID)
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
			total+=svc.FinalPrice
		}
	}

	for _, amt := range payout.ServicePartialAmounts {
		total += amt
	}

	log.Println("Calculated effective amount:", total)
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
			TotalPayAmount:    utils.RoundTo2(p.TotalPayAmount),
			BaseAmount: utils.RoundTo2(p.BaseAmount),
			CommissionPercent: p.CommissionPercent,
			CommissionAmount:  utils.RoundTo2(p.CommissionAmount),
			GSTPercent:        p.GSTPercent,
			GSTAmount:         utils.RoundTo2(p.GSTAmount),
			TDSPercent: p.TDSPercent,
			TDSAmount: utils.RoundTo2(p.TDSAmount),
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
    log.Println("dkjcnkjanjkcds",payoutID)
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
    log.Println("donee hereee",payouts)
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
    log.Println("datajsxannbjdcs",data)
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

		var serviceCommission, serviceTDS, serviceGST, serviceNet float64
		var tdsPercent, gstPercent float64
		

		if hasGSTNumber {
			tdsPercent = constants.DefaultTDSPercent
			gstPercent = constants.DefaultGSTPercent
			serviceCommission = baseAmount * (providerCommissionPercent / 100)
			afterCommission := baseAmount - serviceCommission
			serviceGST = afterCommission * 0.18
			amountWithGST := afterCommission + serviceGST
			serviceTDS = amountWithGST * 0.10
			serviceNet = amountWithGST - serviceTDS
		} else {
			tdsPercent = 0.0
			gstPercent = 0.0
			
			serviceCommission = baseAmount * (providerCommissionPercent / 100)
			serviceTDS = 0
			serviceGST = 0
			serviceNet = baseAmount - serviceCommission
		}

		showComplaintAdjustment := false
		if service.HasComplaintAdjustment && isInSettlement && !service.IsSettledAfterComplaint {
			showComplaintAdjustment = true
		}

		resp := dto.PayoutServiceResponse{
			ID:                      service.ID.Hex(),
			BookingID:               service.ServiceNumber,
			AMCID:                   "amc",
			ProviderID:              payout.ProviderID.Hex(),
			ServiceAmount:           utils.RoundTo2(service.FinalPrice),  // ← Display service.FinalPrice
			TotalPaidAmount:         utils.RoundTo2(transaction.Amount),  // ← NEW: Display transaction.Amount
			CommissionPercent:       providerCommissionPercent,
			CommissionAmount:        utils.RoundTo2(serviceCommission),    // ← Calculated from transaction.Amount
			TDSPercent:              tdsPercent,
			TDSAmount:               utils.RoundTo2(serviceTDS),           // ← Calculated from transaction.Amount
			GSTPercent:              gstPercent,
			GSTAmount:               utils.RoundTo2(serviceGST),           // ← Calculated from transaction.Amount
			NetAmount:               utils.RoundTo2(serviceNet),           // ← Calculated from transaction.Amount
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
		"email_id":      provider.Email,
		"mechanic_type": strings.Join(provider.VehicleType, ", "),
		"vehicle_brand": provider.ProviderBrands,
		"zone":          provider.City,
		"join_date":     provider.CreatedAt,
		"approval_date": provider.UpdatedAt,
		"service_type":   provider.ProviderServices,
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
    log.Println("jkdcsbjsknbkjsdd",req)
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
    log.Println("jdcjbvjbsjhbvhjbsdjh")
	return s.createNewProcessPayout(ctx, providerID, serviceIDs, req, commissionPercent, gstPercent)

}

func (s *PayoutService) mergeIntoExistingPayoutWithoutComplaint( ctx context.Context, existing *domain.PaymentPayout, bucket *providerBucket,providerID primitive.ObjectID, ) error {

	log.Println("kjcbsdjbdjsccsd",bucket)
	if existing.ServicePartialAmounts == nil {
		existing.ServicePartialAmounts = make(map[string]float64)
	}
    log.Println("BucketTotal",bucket.Total)
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

	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	}

	calc := utils.CalculatePayout( effective, commissionPercent, finalGSTPercent,tdsPercent)

	existing.CommissionPercent = commissionPercent
	existing.CommissionAmount = calc.Commission
	existing.GSTPercent = finalGSTPercent
	existing.GSTAmount = calc.GST
	existing.TDSPercent = tdsPercent
	existing.TDSAmount = calc.TDS
	existing.NetPayable = calc.NetPayable
	existing.UpdatedAt = time.Now()

	return  s.payoutRepo.Update(ctx, existing)
}

func (s *PayoutService) createNewPayoutWithoutComplaint( ctx context.Context, providerID primitive.ObjectID, bucket *providerBucket, from, to time.Time ) error {

	provider, err := s.providerRepo.FindByID(ctx, providerID.Hex())
	if err != nil {
		log.Printf("Error fetching provider: %v", err)
		return err
	}

	commissionPercent := constants.DefaultCommissionPercent
	if provider.CommissionPercentage > 0 {
		commissionPercent = provider.CommissionPercentage
	}
    
    log.Println("lcldsmklmskcdnlnslds",bucket)
	tdsPercent := 0.0
	finalGSTPercent := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	}

	calc := utils.CalculatePayout( bucket.Total, commissionPercent, finalGSTPercent,
		tdsPercent)

	payout := &domain.PaymentPayout{
		PayoutID:          time.Now().UnixMilli(),
		ProviderID:        providerID,
		ServiceIDs:        bucket.ServiceIDs,
		TotalPayAmount:    bucket.TotalTransAmount,
		BaseAmount:        bucket.Total,  
		CommissionPercent: commissionPercent,
		CommissionAmount:  calc.Commission,
		GSTPercent:        constants.DefaultGSTPercent,
		GSTAmount:         calc.GST,
		NetPayable:        calc.NetPayable,
		TDSPercent: constants.DefaultTDSPercent,
		TDSAmount:  calc.TDS,
		Status:            domain.PayoutStatusPending,
		PayoutType:        domain.PayoutTypeRegular,
		PeriodFrom:        from,
		PeriodTo:          to,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return s.payoutRepo.Create(ctx, payout)
}

// func (s *PayoutService) updateExistingProviderPayout( ctx context.Context, existing *domain.PaymentPayout, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

// 	if existing.ServicePartialAmounts == nil {
// 		existing.ServicePartialAmounts = make(map[string]float64)
// 	}

// 	serviceExists := make(map[string]bool)
// 	for _, sid := range existing.ServiceIDs {
// 		serviceExists[sid.Hex()] = true
// 	}

// 	for _, sid := range serviceIDs {
// 		key := sid.Hex()
// 		if !serviceExists[key] {
// 			existing.ServiceIDs = append(existing.ServiceIDs, sid)
// 			serviceExists[key] = true
// 			if req.PartialAmount > 0 {
// 				existing.ServicePartialAmounts[key] = req.PartialAmount
// 			}
// 			if req.Amount > 0 {
// 				existing.BaseAmount += req.Amount
// 				existing.TotalPayAmount += req.Amount 
// 			}
// 		} else if req.PartialAmount > 0 {
// 			existing.ServicePartialAmounts[key] = req.PartialAmount
// 		}
// 	}

// 	for _, sid := range serviceIDs {
// 		if req.CancelPayout {
// 			if err := s.cancelServiceProviderPayout(ctx, sid.Hex()); err != nil {
// 				fmt.Printf("ERROR: Failed to cancel service: %v", err)
// 			}
// 			existing.ServicePartialAmounts[sid.Hex()] = 0
// 		}
// 	}

// 	effective := s.calculateEffectiveAmount(ctx, existing)
// 	tdsPercent := 0.0
// 	finalGSTPercent := 0.0

// 	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
// 	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
// 		tdsPercent = constants.DefaultTDSPercent
// 		finalGSTPercent = 18.0
// 	}

// 	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent,tdsPercent)

// 	existing.CommissionPercent = commissionPercent
// 	existing.CommissionAmount = calc.Commission
// 	existing.GSTPercent = finalGSTPercent
// 	existing.GSTAmount = calc.GST
// 	existing.TDSPercent = tdsPercent
// 	existing.TDSAmount = calc.TDS
// 	existing.NetPayable = calc.NetPayable
// 	existing.UpdatedAt = time.Now()

// 	return s.payoutRepo.Update(ctx, existing)
// }

// func (s *PayoutService) updateExistingProviderPayout( ctx context.Context, existing *domain.PaymentPayout, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

// 	if existing.ServicePartialAmounts == nil {
// 		existing.ServicePartialAmounts = make(map[string]float64)
// 	}

// 	serviceExists := make(map[string]bool)
// 	for _, sid := range existing.ServiceIDs {
// 		serviceExists[sid.Hex()] = true
// 	}

// 	for _, sid := range serviceIDs {
// 		key := sid.Hex()
// 		if !serviceExists[key] {
// 			service, err := s.serviceRepo.FindByID(ctx, sid.Hex())
// 			if err != nil {
// 				fmt.Printf("ERROR: Failed to fetch service: %v", err)
// 				continue
// 			}

// 			transaction, err := s.transactionRepo.FindByServiceID(ctx, sid.Hex())
// 			if err != nil {
// 				fmt.Printf("ERROR: Failed to fetch transaction: %v", err)
// 				continue
// 			}

// 			existing.ServiceIDs = append(existing.ServiceIDs, sid)
// 			serviceExists[key] = true
			
// 			if req.PartialAmount > 0 {
// 				existing.ServicePartialAmounts[key] = req.PartialAmount
// 				existing.BaseAmount += req.PartialAmount 
// 			} else {
// 				existing.BaseAmount += service.FinalPrice
// 			}
			
// 			existing.TotalPayAmount += transaction.Amount
// 		} else if req.PartialAmount > 0 {
// 			oldPartialAmount := existing.ServicePartialAmounts[key]
// 			existing.ServicePartialAmounts[key] = req.PartialAmount
// 			existing.BaseAmount = existing.BaseAmount - oldPartialAmount + req.PartialAmount
// 		}
// 	}

// 	for _, sid := range serviceIDs {
// 		if req.CancelPayout {
// 			if err := s.cancelServiceProviderPayout(ctx, sid.Hex()); err != nil {
// 				fmt.Printf("ERROR: Failed to cancel service: %v", err)
// 			}
// 			existing.ServicePartialAmounts[sid.Hex()] = 0
// 		}
// 	}

// 	effective := s.calculateEffectiveAmount(ctx, existing)
// 	tdsPercent := 0.0
// 	finalGSTPercent := 0.0

// 	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
// 	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
// 		tdsPercent = constants.DefaultTDSPercent
// 		finalGSTPercent = 18.0
// 	}

// 	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent,tdsPercent)

// 	existing.CommissionPercent = commissionPercent
// 	existing.CommissionAmount = calc.Commission
// 	existing.GSTPercent = finalGSTPercent
// 	existing.GSTAmount = calc.GST
// 	existing.TDSPercent = tdsPercent
// 	existing.TDSAmount = calc.TDS
// 	existing.NetPayable = calc.NetPayable
// 	existing.UpdatedAt = time.Now()

// 	return s.payoutRepo.Update(ctx, existing)
// }

func (s *PayoutService) updateExistingProviderPayout(
	ctx context.Context,
	existing *domain.PaymentPayout,
	serviceIDs []primitive.ObjectID,
	req dto.PayoutRequest,
	commissionPercent, gstPercent float64,
) error {
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
			transaction, err := s.transactionRepo.FindByServiceID(ctx, sid.Hex())
			if err != nil {
				fmt.Printf("ERROR: Failed to fetch transaction: %v", err)
				continue
			}
			existing.ServiceIDs = append(existing.ServiceIDs, sid)
			serviceExists[key] = true
			existing.TotalPayAmount += transaction.Amount
		}
		
		if req.PartialAmount > 0 {
			existing.ServicePartialAmounts[key] = req.PartialAmount
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
	existing.BaseAmount = effective
	
	tdsPercent := 0.0
	finalGSTPercent := 0.0
	kyc, _ := s.kycRepo.FindByProviderID(ctx, existing.ProviderID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	}
	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)
	existing.CommissionPercent = commissionPercent
	existing.CommissionAmount = calc.Commission
	existing.GSTPercent = finalGSTPercent
	existing.GSTAmount = calc.GST
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

func (s *PayoutService) createDeductionProviderPayout( ctx context.Context, providerID primitive.ObjectID, serviceIDs []primitive.ObjectID, req dto.PayoutRequest, commissionPercent, gstPercent float64 ) error {

	effective := req.PartialAmount
	if effective == 0 {
		effective = req.Amount
	}

	tdsPercent := 0.0
	finalGSTPercent := 0.0

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)
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
		ComplaintInternalID:   req.ComplaintInternalID,
		TotalPayAmount:            -req.Amount,
		BaseAmount:            -req.Amount,  
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      -calc.Commission,
		GSTPercent:            gstPercent,
		GSTAmount:             -calc.GST,
		TDSPercent:            tdsPercent,
		TDSAmount:  -calc.TDS,
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
    log.Println("jnfwejnkjewnjffe")
	log.Println("doneee hereee req",req)
	servicePartialAmounts := make(map[string]float64)
	effective := req.Amount
    log.Println("kjdksjbsdjkbdscddcs",effective)
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

	kyc, _ := s.kycRepo.FindByProviderID(ctx, providerID)

	tdsPercent := 0.0
    finalGSTPercent := 0.0

	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		tdsPercent = constants.DefaultTDSPercent
		finalGSTPercent = 18.0
	}

	calc := utils.CalculatePayout(effective, commissionPercent, finalGSTPercent, tdsPercent)
	complaintID, _ := primitive.ObjectIDFromHex(req.ComplaintID)
    
	payout := &domain.PaymentPayout{
		PayoutID:              time.Now().UnixMilli(),
		ProviderID:            providerID,
		ServiceIDs:            serviceIDs,
		ComplaintID:           &complaintID,
		ComplaintInternalID:   req.ComplaintInternalID,
		TotalPayAmount:        req.Amount,
		BaseAmount:            req.Amount, 
		ServicePartialAmounts: servicePartialAmounts,
		CommissionPercent:     commissionPercent,
		CommissionAmount:      calc.Commission,
		GSTPercent:            gstPercent,
		GSTAmount:             calc.GST,
		TDSPercent:            tdsPercent,
		TDSAmount:             calc.TDS,
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

type GetSettlementStatsRequest struct {
	Tab        string `form:"tab"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	ProviderID string `form:"provider_id"`
}

type SettlementStatsResponse struct {
	TotalPayout           float64 `json:"total_payout"`          
	TotalProviderRevenue  float64 `json:"total_provider_revenue"` 
	VahanwireCommission   float64 `json:"vahanwire_commission"`  
	VahanwireGST         float64 `json:"vahanwire_gst"`         
	PartnerGST           float64 `json:"partner_gst"`            // GST on service amount
	TotalTDS             float64 `json:"total_tds"`             
}

func (s *PayoutService) GetPayoutStats(
	ctx context.Context,
	req *GetSettlementStatsRequest,
) (*SettlementStatsResponse, error) {
	
	if req.Tab == "" {
		return nil, fmt.Errorf("tab parameter is required. Use 'pending' or 'settled'")
	}

	filter := bson.M{}
	
	switch strings.ToLower(req.Tab) {
	case "pending":
		filter["status"] = "pending"
	case "settled":
		filter["status"] = "settled"
	default:
		return nil, fmt.Errorf("invalid tab value. Use 'pending' or 'settled'")
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
		if _, exists := filter["createdAt"]; exists {
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

	// partnerGST := 0.0
	// if stats.ServiceAmount > 0 {
	// 	partnerGST = stats.ServiceAmount * 0.18
	// }

	response := &SettlementStatsResponse{
		TotalPayout:           utils.RoundTo2(stats.TotalPayAmount),
		TotalProviderRevenue:  utils.RoundTo2(stats.NetPayable),
		VahanwireCommission:   utils.RoundTo2(stats.CommissionAmount),
		VahanwireGST:         utils.RoundTo2(stats.GSTAmount),
		PartnerGST:           0,
		TotalTDS:             utils.RoundTo2(stats.TDSAmount),
	}

	return response, nil
}

