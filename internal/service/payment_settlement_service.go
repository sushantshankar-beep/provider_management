package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SettlementRequest struct {
	PayoutID      string   `json:"payout_id" binding:"required"`
	ServiceIDs    []string `json:"service_ids" binding:"required"`
	Justification string   `json:"justification"`
	PaymentMode   string   `json:"payment_mode" binding:"required"`
	PaymentMethod string   `json:"payment_method" binding:"required"`
}

type SettlementService struct {
	serviceRepo           *repository.AcceptedServiceRepo
	settlementRepo        *repository.ProviderSettlementRepo
	payoutRepo            *repository.PaymentPayoutRepo
	providerRepo          *repository.ProviderRepo
	settlementHistoryRepo *repository.SettlementHistoryRepository
	kycRepo               *repository.ProviderKYCRepository
	transactionRepo       *repository.TransactionRepo
}

type GetSettlementsRequest struct {
	Tab         string `form:"tab"`
	ProviderID  string `form:"provider_id"`
	PayoutID    string `form:"payout_id"`
	Status      string `form:"status"`
	PaymentMode string `form:"payment_mode"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Page        int64  `form:"page,default=1"`
	Limit       int64  `form:"limit,default=10"`
	SortField   string `form:"sort_field,default=createdAt"`
	SortOrder   string `form:"sort_order,default=desc"`
	Search      string `form:"search"`
}

type CreateSettlementRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Note          string  `json:"note"`
}

type ProviderSettlementResponse struct {
	ID                 primitive.ObjectID      `json:"_id"`
	SettlementID       string                  `json:"settlementId"`
	ProviderID         primitive.ObjectID      `json:"providerId"`
	ProviderName       string                  `json:"providerName"`
	ProviderCode       string                  `json:"providerCode"`
	ProviderEmail      string                  `json:"providerEmail"`
	AccountNo          string                  `json:"accountNo"`
	PaymentType        string                  `json:"paymentType"`
	IfscCode           string                  `json:"ifscCode"`
	TotalAmount        float64                 `json:"totalAmount"`
	IsDeduction        bool                    `json:"isDeduction"`
	PaymentMode        string                  `json:"paymentMode"`
	PaymentMethod      string                  `json:"paymentMethod"`
	Justification      string                  `json:"justification"`
	Status             domain.SettlementStatus `json:"settelementStatus"`
	PayoutIdNumber     string                  `json:"payoutId"`
	DebitAccountNumber string                  `json:"debitAccountNumber"`
	SettledAt          *time.Time              `json:"settledAt"`
	CreatedAt          time.Time               `json:"createdAt"`
}

type ProviderSettlementIDResponse struct {
	ID              primitive.ObjectID      `json:"id"`
	SettlementID    string                  `json:"settlement_id"`
	PayoutIdNumber  string                  `json:"payout_id"`
	ProviderCode    string                  `json:"providerCode"`
	ProviderID      primitive.ObjectID      `json:"provider_id"`
	ProviderName    string                  `json:"provider_name"`
	AccountNo       string                  `json:"account_no"`
	IfscCode        string                  `json:"ifsc_code"`
	TotalAmount     float64                 `json:"total_amount"`
	PaymentMode     string                  `json:"payment_mode"`
	PaymentMethod   string                  `json:"payment_method"`
	Justification   string                  `json:"justification,omitempty"`
	Status          domain.SettlementStatus `json:"status"`
	SettledPostData *domain.SettledPostData `json:"settled_post_data,omitempty"`
	SettledAt       *time.Time              `json:"settled_at,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
}

type GetBookingsByFinancialTypeRequest struct {
	FinancialType string `form:"stats_type" binding:"required"`
	Status        string `form:"status"`
	StartDate     string `form:"start_date"`
	EndDate       string `form:"end_date"`
	Page          int64  `form:"page,default=1"`
	Limit         int64  `form:"limit,default=10"`
	SortField     string `form:"sort_field,default=createdAt"`
	SortOrder     string `form:"sort_order,default=desc"`
}

type BookingFinancialDetail struct {
	ServiceID          primitive.ObjectID `json:"service_id"`
	ServiceNumber      string             `json:"serviceNumber"`
	ProviderName       string             `json:"provider_name"`
	TotalPayAmount     float64            `json:"amount"`
	ProviderCode       string             `json:"provider_code"`
	ProviderID         primitive.ObjectID `json:"provider_id"`
	SettlementID       string             `json:"settlement_id"`
	PayoutID           string             `json:"payout_id"`
	BaseAmount         float64            `json:"base_amount"`
	CommissionAmount   float64            `json:"commission_amount,omitempty"`
	CommissionPercent  float64            `json:"commission_percent,omitempty"`
	GSTAmount          float64            `json:"gst_amount,omitempty"`
	GSTPercent         float64            `json:"gst_percent,omitempty"`
	TDSAmount          float64            `json:"tds_amount,omitempty"`
	TDSPercent         float64            `json:"tds_percent,omitempty"`
	VahanwireGSTAmount float64            `json:"vahanwire_gst_amount,omitempty"`
	NetAmount          float64            `json:"net_amount"`
	Status             string             `json:"status"`
	CreatedAt          time.Time          `json:"created_at"`
	SettledAt          *time.Time         `json:"settled_at,omitempty"`
}

type BookingFinancialResponse struct {
	Bookings   []BookingFinancialDetail `json:"bookings"`
	Total      int64                    `json:"total"`
	TotalPages int64                    `json:"total_pages"`
	Summary    FinancialSummary         `json:"summary"`
}

type FinancialSummary struct {
	TotalAmount float64 `json:"total_amount"`
	Count       int64   `json:"count"`
}

func NewSettlementService(
	serviceRepo *repository.AcceptedServiceRepo,
	settlementRepo *repository.ProviderSettlementRepo,
	payoutRepo *repository.PaymentPayoutRepo,
	providerRepo *repository.ProviderRepo,
	settlementHistoryRepo *repository.SettlementHistoryRepository,
	kycRepo *repository.ProviderKYCRepository,
	transactionRepo *repository.TransactionRepo,
) *SettlementService {
	return &SettlementService{
		serviceRepo:           serviceRepo,
		settlementRepo:        settlementRepo,
		payoutRepo:            payoutRepo,
		providerRepo:          providerRepo,
		settlementHistoryRepo: settlementHistoryRepo,
		kycRepo:               kycRepo,
		transactionRepo:       transactionRepo,
	}
}

func parsePayoutID(payoutIDStr string) (int64, error) {
	numStr := strings.TrimPrefix(payoutIDStr, "PAY")
	return strconv.ParseInt(numStr, 10, 64)
}

func (s *SettlementService) CreateSettlement(
	ctx context.Context,
	req *SettlementRequest,
) (*domain.ProviderSettlement, error) {

	payoutIDInt, err := parsePayoutID(req.PayoutID)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID format")
	}

	payout, err := s.payoutRepo.FindByPayoutID(ctx, payoutIDInt)
	if err != nil {
		return nil, fmt.Errorf("payout not found")
	}

	provider, err := s.providerRepo.FindByID(ctx, payout.ProviderID.Hex())
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	kyc, err := s.kycRepo.FindByProviderID(ctx, payout.ProviderID)
	if err != nil || kyc == nil {
		return nil, fmt.Errorf("provider KYC not found, cannot process settlement")
	}

	if kyc.Status != domain.KYC_APPROVED {
		return nil, fmt.Errorf("provider KYC is not approved, cannot process settlement")
	}

	if strings.TrimSpace(kyc.Bank.AccountNumber) == "" ||
		strings.TrimSpace(kyc.Bank.IFSC) == "" {
		return nil, fmt.Errorf("provider bank details incomplete, cannot process settlement")
	}

	providerCommissionPercent := payout.CommissionPercent
	if provider.CommissionPercentage > 0 {
		providerCommissionPercent = provider.CommissionPercentage
	}

	hasGSTNumber := false
	if kyc != nil && strings.TrimSpace(kyc.Bank.GSTNumber) != "" {
		hasGSTNumber = true
	}

	tdsPercent := 0.0
	gstPercent := 0.0
	if hasGSTNumber {
		tdsPercent = 10.0
		gstPercent = 18.0
	}

	serviceObjIDs := make([]primitive.ObjectID, 0, len(req.ServiceIDs))
	for _, sid := range req.ServiceIDs {
		id, err := primitive.ObjectIDFromHex(sid)
		if err != nil {
			return nil, fmt.Errorf("invalid service id %s", sid)
		}
		serviceObjIDs = append(serviceObjIDs, id)
	}

	for _, sid := range serviceObjIDs {
		found := false
		for _, ps := range payout.ServiceIDs {
			if ps == sid {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("service %s not part of payout", sid.Hex())
		}
	}

	services, err := s.serviceRepo.FindByIDs(ctx, serviceObjIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services")
	}

	for _, service := range services {
		isInSettlement := service.SettlementStatus == domain.SettleStatusPending ||
			service.SettlementStatus == domain.SettleStatusSettled

		if isInSettlement && !service.HasComplaintAdjustment {
			return nil, fmt.Errorf(
				"service %s already settled and has no complaint adjustment",
				service.ID.Hex(),
			)
		}

		if isInSettlement && service.HasComplaintAdjustment && service.IsSettledAfterComplaint {
			return nil, fmt.Errorf(
				"service %s already settled after complaint resolution",
				service.ID.Hex(),
			)
		}
	}

	transactions, err := s.transactionRepo.FindByServiceIDs(ctx, serviceObjIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions")
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found for services")
	}

	transactionMap := make(map[string]domain.Transaction)
	for _, txn := range transactions {
		transactionMap[txn.ServiceID] = txn
	}

	var settlementAmount float64

	for _, service := range services {
		serviceKey := service.ID.Hex()

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt == 0 {
				continue
			}
		}

		if service.IsPayoutCancelled {
			continue
		}

		transaction, exists := transactionMap[serviceKey]
		if !exists {
			return nil, fmt.Errorf("transaction not found for service %s", serviceKey)
		}

		baseAmount := service.FinalPrice

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt > 0 {
				baseAmount = partialAmt
				log.Printf("Service %s using partial amount: %.2f (original: %.2f)",
					serviceKey, partialAmt, transaction.Amount)
			}
		}

		if service.HasComplaintAdjustment {
			settlementAmount -= service.PendingDeductionAmount
			continue
		}

		var commission, tds, gst, net float64
		if hasGSTNumber {
			commission = baseAmount * (providerCommissionPercent / 100)
			afterCommission := baseAmount - commission
			gst = baseAmount * 0.18
			amountWithGST := afterCommission + gst
			tds = amountWithGST * 0.10
			net = amountWithGST - tds
		} else {
			commission = baseAmount * (providerCommissionPercent / 100)
			tds = 0
			gst = 0
			net = baseAmount - commission
		}

		settlementAmount += net
	}

	if settlementAmount <= 0 {
		return nil, fmt.Errorf("settlement amount is %.2f, cannot process settlement", settlementAmount)
	}

	now := time.Now()

	settledBookings := make([]domain.SettledBooking, 0)
	for _, service := range services {
		if service.IsPayoutCancelled {
			continue
		}

		serviceKey := service.ID.Hex()
		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt == 0 {
				continue
			}
		}

		transaction := transactionMap[serviceKey]
		var netAmount float64

		if service.HasComplaintAdjustment {
			netAmount = -service.PendingDeductionAmount
		} else {
			baseAmount := service.FinalPrice
			if payout.ServicePartialAmounts != nil {
				if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt > 0 {
					baseAmount = partialAmt
				}
			}

			if hasGSTNumber {
				commission := baseAmount * (providerCommissionPercent / 100)
				afterCommission := baseAmount - commission
				gst := baseAmount * 0.18
				amountWithGST := afterCommission + gst
				tds := amountWithGST * 0.10
				netAmount = amountWithGST - tds
			} else {
				commission := baseAmount * (providerCommissionPercent / 100)
				netAmount = baseAmount - commission
			}
		}

		settlementType := "regular"
		if service.HasComplaintAdjustment {
			settlementType = "complaint_deduction"
		} else if payout.PayoutType == domain.PayoutTypeComplaint {
			settlementType = "complaint"
		}

		settledBooking := domain.SettledBooking{
			ServiceID:        service.ID,
			ServiceRequestNo: service.ServiceRequestNo,
			OriginalAmount:   utils.RoundTo2(transaction.Amount),
			SettledAmount:    utils.RoundTo2(netAmount),
			SettlementType:   settlementType,
		}

		settledBookings = append(settledBookings, settledBooking)
	}

	settlement := &domain.ProviderSettlement{
		SettlementID:    time.Now().UnixMilli(),
		PayoutID:        payout.ID,
		ProviderID:      provider.ID,
		ProviderName:    provider.Name,
		AccountNo:       kyc.Bank.AccountNumber,
		IfscCode:        kyc.Bank.IFSC,
		TotalAmount:     utils.RoundTo2(settlementAmount),
		PaymentMode:     req.PaymentMode,
		PaymentMethod:   req.PaymentMethod,
		Justification:   req.Justification,
		Status:          domain.SettleStatusPending,
		SettledBookings: settledBookings,
		SettledAt:       &now,
		CreatedAt:       now,
	}

	if err := s.settlementRepo.Create(ctx, settlement); err != nil {
		return nil, fmt.Errorf("failed to create settlement")
	}

	for _, service := range services {
		serviceKey := service.ID.Hex()

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt == 0 {
				continue
			}
		}

		if service.IsPayoutCancelled {
			continue
		}

		existingRecord, _ := s.settlementHistoryRepo.FindByServiceID(ctx, service.ID)

		if service.HasComplaintAdjustment && existingRecord != nil {
			updateData := map[string]interface{}{
				"deductionAmount":       utils.RoundTo2(service.PendingDeductionAmount),
				"hasDeduction":          true,
				"deductionSettlementId": settlement.ID,
				"deductionPayoutId":     payout.ID,
				"deductionComplaintId":  payout.ComplaintID,
				"deductionRemarks":      req.Justification,
				"deductionProcessedAt":  now,
				"updatedAt":             now,
			}

			if err := s.settlementHistoryRepo.UpdateByID(ctx, existingRecord.ID, updateData); err != nil {
				return nil, fmt.Errorf("failed to update settlement record with deduction: %v", err)
			}

			continue
		}

		originalAmount := service.FinalPrice

		var partialAmount float64
		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt > 0 {
				partialAmount = partialAmt
			}
		}

		calculationAmount := originalAmount
		if partialAmount > 0 {
			calculationAmount = partialAmount
		}

		var commission, tds, gst, vahanwireGST, netAmount float64
		if hasGSTNumber {
			commission = calculationAmount * (providerCommissionPercent / 100)
			afterCommission := calculationAmount - commission
			gst = calculationAmount * 0.18
			amountWithGST := afterCommission + gst
			tds = amountWithGST * 0.10
			netAmount = amountWithGST - tds
			vahanwireGST = 0
		} else {
			commission = calculationAmount * (providerCommissionPercent / 100)
			vahanwireGST = calculationAmount * 0.18
			tds = 0
			gst = 0
			netAmount = calculationAmount - commission
		}

		settlementType := "regular"
		if payout.PayoutType == domain.PayoutTypeComplaint {
			settlementType = "complaint"
		}

		settlementRecord := &domain.SettlementRecord{
			ServiceID:             service.ID,
			PayoutID:              payout.ID,
			ProviderID:            payout.ProviderID,
			SettlementID:          settlement.ID,
			OriginalAmount:        utils.RoundTo2(originalAmount),
			PartialAmount:         utils.RoundTo2(partialAmount),
			SettlementAmount:      utils.RoundTo2(calculationAmount),
			CommissionPercent:     providerCommissionPercent,
			CommissionAmount:      utils.RoundTo2(commission),
			TDSPercent:            tdsPercent,
			TDSAmount:             utils.RoundTo2(tds),
			GSTPercent:            gstPercent,
			GSTAmount:             utils.RoundTo2(gst),
			NetAmount:             utils.RoundTo2(netAmount),
			DeductionAmount:       0,
			HasDeduction:          false,
			DeductionSettlementID: nil,
			DeductionPayoutID:     nil,
			DeductionComplaintID:  nil,
			DeductionRemarks:      "",
			DeductionProcessedAt:  nil,
			SettlementType:        settlementType,
			VahanwireGSTAmount:    utils.RoundTo2(vahanwireGST),
			ComplaintID:           payout.ComplaintID,
			SettlementStatus:      domain.SettleStatusPending,
			CreatedAt:             now,
			UpdatedAt:             now,
			Remarks:               req.Justification,
		}

		if _, err := s.settlementHistoryRepo.Create(ctx, settlementRecord); err != nil {
			return nil, fmt.Errorf("failed to create settlement record: %v", err)
		}
	}

	servicesToSettle := []primitive.ObjectID{}
	for _, service := range services {
		if service.IsPayoutCancelled {
			continue
		}

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[service.ID.Hex()]; exists && partialAmt == 0 {
				continue
			}
		}

		servicesToSettle = append(servicesToSettle, service.ID)
	}

	if len(servicesToSettle) > 0 {
		if err := s.serviceRepo.MarkAsSettled(ctx, servicesToSettle, settlement.ID); err != nil {
			return nil, err
		}
	}

	for _, service := range services {
		if service.HasComplaintAdjustment && service.SettlementStatus == domain.SettleStatusSettled {
			if err := s.serviceRepo.MarkAsSettledAfterComplaint(ctx, service.ID, &now); err != nil {
				log.Printf("Failed to mark service %s as settled after complaint: %v", service.ID.Hex(), err)
			}
		}
	}

	unsettledCount, err := s.serviceRepo.CountUnsettledByIDs(ctx, payout.ServiceIDs)
	if err != nil {
		return nil, err
	}

	if unsettledCount == 0 {
		err = s.payoutRepo.UpdateStatus(
			ctx,
			payout.ID,
			domain.PayoutStatusSettled,
			&settlement.ID,
		)
	} else {
		err = s.payoutRepo.UpdateStatus(
			ctx,
			payout.ID,
			domain.PayoutStatusPartiallySettled,
			nil,
		)
	}

	if err != nil {
		return nil, err
	}

	return settlement, nil
}

func (s *SettlementService) GetSettlements(
	ctx context.Context,
	req *GetSettlementsRequest,
) ([]ProviderSettlementResponse, int64, int64, error) {

	filter := bson.M{}

	if req.Tab != "" {
		switch strings.ToLower(req.Tab) {
		case "pending":
			filter["status"] = "pending"
		case "settled":
			filter["status"] = "settled"
		default:
			return nil, 0, 0, fmt.Errorf("invalid tab value")
		}
	}

	if req.Status != "" {
		filter["status"] = req.Status
	}

	if req.ProviderID != "" {
		providerObjID, err := primitive.ObjectIDFromHex(req.ProviderID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid provider id")
		}
		filter["providerId"] = providerObjID
	}

	if req.PayoutID != "" {
		payoutObjID, err := primitive.ObjectIDFromHex(req.PayoutID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid payout id")
		}
		filter["payoutId"] = payoutObjID
	}

	if req.PaymentMode != "" {
		filter["paymentMode"] = req.PaymentMode
	}

	if req.StartDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid start date")
		}
		filter["createdAt"] = bson.M{"$gte": start}
	}

	if req.EndDate != "" {
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid end date")
		}
		end = end.Add(24 * time.Hour)

		if filter["createdAt"] != nil {
			filter["createdAt"].(bson.M)["$lte"] = end
		} else {
			filter["createdAt"] = bson.M{"$lte": end}
		}
	}

	if strings.TrimSpace(req.Search) != "" {
		search := strings.TrimSpace(req.Search)
		or := bson.A{}

		if strings.HasPrefix(strings.ToUpper(search), "SET") {
			if num, err := strconv.ParseInt(strings.TrimPrefix(strings.ToUpper(search), "SET"), 10, 64); err == nil {
				or = append(or, bson.M{"settlementId": num})
			}
		}

		if strings.HasPrefix(strings.ToUpper(search), "PAY") {
			if num, err := strconv.ParseInt(strings.TrimPrefix(strings.ToUpper(search), "PAY"), 10, 64); err == nil {
				payout, _ := s.payoutRepo.FindByPayoutNumber(ctx, num)
				if payout != nil {
					or = append(or, bson.M{"payoutId": payout.ID})
				}
			}
		}

		if provider, _ := s.providerRepo.FindByProviderCode(ctx, search); provider != nil {
			or = append(or, bson.M{"providerId": provider.ID})
		}

		if len(or) > 0 {
			filter["$or"] = or
		}
	}

	skip := (req.Page - 1) * req.Limit
	sortOrder := -1
	if strings.ToLower(req.SortOrder) == "asc" {
		sortOrder = 1
	}

	settlements, total, err := s.settlementRepo.GetSettlements(
		ctx,
		filter,
		skip,
		req.Limit,
		req.SortField,
		sortOrder,
	)
	if err != nil {
		return nil, 0, 0, err
	}

	responses := make([]ProviderSettlementResponse, 0, len(settlements))

	for _, settlement := range settlements {

		resp := ProviderSettlementResponse{
			ID:                 settlement.ID,
			SettlementID:       "SET" + strconv.FormatInt(settlement.SettlementID, 10),
			ProviderID:         settlement.ProviderID,
			PaymentType:        "NEFT",
			AccountNo:          settlement.AccountNo,
			IfscCode:           settlement.IfscCode,
			TotalAmount:        utils.RoundTo2(settlement.TotalAmount),
			PaymentMode:        settlement.PaymentMode,
			PaymentMethod:      settlement.PaymentMethod,
			Justification:      settlement.Justification,
			Status:             settlement.Status,
			SettledAt:          settlement.SettledAt,
			DebitAccountNumber: "",
			CreatedAt:          settlement.CreatedAt,
		}

		provider, _ := s.providerRepo.FindByID(ctx, settlement.ProviderID.Hex())
		if provider != nil {
			resp.ProviderEmail = provider.Email
			resp.ProviderCode = provider.ProviderCode

			kyc, _ := s.kycRepo.FindByProviderID(ctx, provider.ID)
			if kyc != nil {
				resp.ProviderName = kyc.Bank.AccountHolderName
			}
		}

		payout, _ := s.payoutRepo.FindByID(ctx, settlement.PayoutID)
		if payout != nil {
			resp.PayoutIdNumber = "PAY" + strconv.FormatInt(payout.PayoutID, 10)
			resp.IsDeduction = payout.IsDeduction
		}

		responses = append(responses, resp)
	}

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return responses, total, totalPages, nil
}

func (s *SettlementService) ChangeProviderSettlementStatus(
	ctx context.Context,
	settlementID string,
	req *CreateSettlementRequest,
) (*domain.ProviderSettlement, error) {

	objID, err := primitive.ObjectIDFromHex(settlementID)
	if err != nil {
		return nil, fmt.Errorf("invalid settlement id: %w", err)
	}

	settlement, err := s.settlementRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("failed to find settlement: %w", err)
	}
	if settlement == nil {
		return nil, fmt.Errorf("settlement not found")
	}

	if settlement.Status == domain.SettleStatusSettled {
		return nil, fmt.Errorf("settlement already marked as settled")
	}

	serviceIDs := make([]primitive.ObjectID, len(settlement.SettledBookings))
	for i, booking := range settlement.SettledBookings {
		serviceIDs[i] = booking.ServiceID
	}

	if len(serviceIDs) > 0 {
		now := time.Now()

		if err := s.settlementHistoryRepo.UpdateStatusBySettlementID(
			ctx,
			settlement.ID,
			domain.SettleStatusSettled,
			&now,
		); err != nil {
			return nil, fmt.Errorf("failed to update settlement history: %w", err)
		}

		if err := s.serviceRepo.MarkServicesAsSettled(
			ctx,
			serviceIDs,
			settlement.ID,
			&now,
		); err != nil {
			return nil, fmt.Errorf("failed to mark services as settled: %w", err)
		}
	}

	updateData := map[string]interface{}{
		"status":    domain.SettleStatusSettled,
		"settledAt": time.Now(),
		"settledPostData": &domain.SettledPostData{
			TransactionID: req.TransactionID,
			Amount:        req.Amount,
			Status:        req.Status,
			Note:          req.Note,
		},
	}

	if err := s.settlementRepo.Update(ctx, settlement.ID, updateData); err != nil {
		return nil, fmt.Errorf("failed to update settlement: %w", err)
	}

	updatedSettlement, err := s.settlementRepo.FindByID(ctx, settlement.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated settlement: %w", err)
	}

	return updatedSettlement, nil
}

func (s *SettlementService) GetSettlementByID(
	ctx context.Context,
	settlementID string,
) (*ProviderSettlementIDResponse, error) {

	id, err := primitive.ObjectIDFromHex(settlementID)
	if err != nil {
		return nil, fmt.Errorf("invalid settlement id: %w", err)
	}

	settlement, err := s.settlementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find settlement: %w", err)
	}

	if settlement == nil {
		return nil, fmt.Errorf("settlement not found")
	}

	resp := ProviderSettlementIDResponse{
		ID:              settlement.ID,
		SettlementID:    "SET" + strconv.FormatInt(settlement.SettlementID, 10),
		ProviderID:      settlement.ProviderID,
		ProviderName:    settlement.ProviderName,
		AccountNo:       settlement.AccountNo,
		IfscCode:        settlement.IfscCode,
		TotalAmount:     utils.RoundTo2(settlement.TotalAmount),
		PaymentMode:     settlement.PaymentMode,
		PaymentMethod:   settlement.PaymentMethod,
		Justification:   settlement.Justification,
		Status:          settlement.Status,
		SettledPostData: settlement.SettledPostData,
		SettledAt:       settlement.SettledAt,
		CreatedAt:       settlement.CreatedAt,
	}

	if !settlement.PayoutID.IsZero() {
		payout, err := s.payoutRepo.FindByID(ctx, settlement.PayoutID)
		if err == nil && payout != nil {
			resp.PayoutIdNumber = "PAY" + strconv.FormatInt(payout.PayoutID, 10)
		}
	}

	if provider, _ := s.providerRepo.FindByID(ctx, settlement.ProviderID.Hex()); provider != nil {
		resp.ProviderCode = provider.ProviderCode
	}

	return &resp, nil
}

func (s *SettlementService) GetSettlementsByIDs(
	ctx context.Context,
	ids []string,
) ([]domain.ProviderSettlement, error) {
	return s.settlementRepo.FindBySettlementIDs(ctx, ids)
}

func (s *SettlementService) GetBookingsForPayoutStats(
	ctx context.Context,
	req *GetBookingsByFinancialTypeRequest,
) (*BookingFinancialResponse, error) {

	filter := bson.M{}

	if req.Status != "" {
		filter["settlementStatus"] = req.Status
	}

	if req.StartDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date")
		}
		filter["createdAt"] = bson.M{"$gte": start}
	}

	if req.EndDate != "" {
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date")
		}
		end = end.Add(24 * time.Hour)

		if filter["createdAt"] != nil {
			filter["createdAt"].(bson.M)["$lte"] = end
		} else {
			filter["createdAt"] = bson.M{"$lte": end}
		}
	}

	switch req.FinancialType {
	case "gst":
		filter["gstAmount"] = bson.M{"$gt": 0}
	case "commission":
		filter["commissionAmount"] = bson.M{"$gt": 0}
	case "tds":
		filter["tdsAmount"] = bson.M{"$gt": 0}
	case "vahanwire_gst":
		filter["vahanwireGstAmount"] = bson.M{"$gt": 0}
	default:
		return nil, fmt.Errorf("invalid financial type")
	}

	skip := (req.Page - 1) * req.Limit
	sortOrder := -1
	if strings.ToLower(req.SortOrder) == "asc" {
		sortOrder = 1
	}

	records, total, err := s.settlementHistoryRepo.GetSettlementRecordsPayout(
		ctx,
		filter,
		skip,
		req.Limit,
		req.SortField,
		sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch settlement records: %w", err)
	}

	bookings := make([]BookingFinancialDetail, 0, len(records))
	var totalAmount float64

	for _, record := range records {

		service, err := s.serviceRepo.FindByID(ctx, record.ServiceID.Hex())
		if err != nil || service == nil {
			log.Printf("service not found: %s", record.ServiceID.Hex())
			continue
		}

		provider, _ := s.providerRepo.FindByID(ctx, record.ProviderID.Hex())
		payout, _ := s.payoutRepo.FindByID(ctx, record.PayoutID)
		settlement, _ := s.settlementRepo.FindByID(ctx, record.SettlementID)

		transaction, err := s.transactionRepo.FindByServiceID(ctx, record.ServiceID.Hex())
		if err != nil {
			log.Printf("transaction fetch failed for service %s: %v", record.ServiceID.Hex(), err)
			continue
		}
		if transaction == nil {
			log.Printf("transaction not found for service %s", record.ServiceID.Hex())
			continue
		}

		booking := BookingFinancialDetail{
			ServiceID:      record.ServiceID,
			ServiceNumber:  service.ServiceNumber,
			TotalPayAmount: transaction.Amount,
			BaseAmount:     record.SettlementAmount,
			NetAmount:      record.NetAmount,
			Status:         string(record.SettlementStatus),
			CreatedAt:      record.CreatedAt,
			SettledAt:      record.SettledAt,
		}

		if provider != nil {
			booking.ProviderName = provider.Name
			booking.ProviderCode = provider.ProviderCode
			booking.ProviderID = provider.ID
		}

		if settlement != nil {
			booking.SettlementID = "SET" + strconv.FormatInt(settlement.SettlementID, 10)
		}

		if payout != nil {
			booking.PayoutID = "PAY" + strconv.FormatInt(payout.PayoutID, 10)
		}

		switch req.FinancialType {
		case "gst":
			booking.GSTAmount = record.GSTAmount
			booking.GSTPercent = record.GSTPercent
			totalAmount += record.GSTAmount

		case "commission":
			booking.CommissionAmount = record.CommissionAmount
			booking.CommissionPercent = record.CommissionPercent
			totalAmount += record.CommissionAmount

		case "tds":
			booking.TDSAmount = record.TDSAmount
			booking.TDSPercent = record.TDSPercent
			totalAmount += record.TDSAmount

		case "vahanwire_gst":
			booking.VahanwireGSTAmount = record.VahanwireGSTAmount
			totalAmount += record.VahanwireGSTAmount
		}

		bookings = append(bookings, booking)
	}

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return &BookingFinancialResponse{
		Bookings:   bookings,
		Total:      total,
		TotalPages: totalPages,
		Summary: FinancialSummary{
			TotalAmount: utils.RoundTo2(totalAmount),
			Count:       int64(len(bookings)),
		},
	}, nil
}
