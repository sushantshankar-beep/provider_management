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
	serviceRepo    *repository.AcceptedServiceRepo
	settlementRepo *repository.ProviderSettlementRepo
	payoutRepo     *repository.PaymentPayoutRepo
	providerRepo   *repository.ProviderRepo
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
}

type CreateSettlementRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Note          string  `json:"note"`
}

type ProviderSettlementResponse struct {
	ID             primitive.ObjectID      `json:"_id"`
	SettlementID   string                  `json:"settlementId"`
	ProviderID     primitive.ObjectID      `json:"providerId"`
	ProviderName   string                  `json:"providerName"`
	AccountNo      string                  `json:"accountNo"`
	IfscCode       string                  `json:"ifscCode"`
	TotalAmount    float64                 `json:"totalAmount"`
	IsDeduction    bool                    `json:"isDeduction"`
	PaymentMode    string                  `json:"paymentMode"`
	PaymentMethod  string                  `json:"paymentMethod"`
	Justification  string                  `json:"justification"`
	Status         domain.SettlementStatus `json:"status"`
	PayoutIdNumber string                  `json:"payoutId"`
	SettledAt      *time.Time              `json:"settledAt"`
	CreatedAt      time.Time               `json:"createdAt"`
}

type ProviderSettlementIDResponse struct {
	ID              primitive.ObjectID      `json:"id"`
	SettlementID    string                  `json:"settlement_id"`
	PayoutIdNumber  string                  `json:"payout_id"`
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

func NewSettlementService(
	serviceRepo *repository.AcceptedServiceRepo,
	settlementRepo *repository.ProviderSettlementRepo,
	payoutRepo *repository.PaymentPayoutRepo,
	providerRepo *repository.ProviderRepo,
) *SettlementService {
	return &SettlementService{
		serviceRepo:    serviceRepo,
		settlementRepo: settlementRepo,
		payoutRepo:     payoutRepo,
		providerRepo:   providerRepo,
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

	if provider.BankDetails == nil {
		return nil, fmt.Errorf("provider bank details missing")
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
		if service.IsSettled && !service.HasComplaintAdjustment {
			return nil, fmt.Errorf(
				"service %s already settled and has no complaint adjustment",
				service.ID.Hex(),
			)
		}
		if service.IsSettled && service.HasComplaintAdjustment && service.IsSettledAfterComplaint {
			return nil, fmt.Errorf(
				"service %s already settled after complaint resolution",
				service.ID.Hex(),
			)
		}
	}

	var settlementAmount float64

	for _, service := range services {
		serviceKey := service.ID.Hex()

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt == 0 {
				log.Printf("Service %s has cancelled payout (amount=0), skipping from settlement", serviceKey)
				continue
			}
		}

		if service.IsPayoutCancelled {
			log.Printf("Service %s marked as cancelled, skipping from settlement", serviceKey)
			continue
		}

		baseAmount := service.FinalPrice

		if payout.ServicePartialAmounts != nil {
			if partialAmt, exists := payout.ServicePartialAmounts[serviceKey]; exists && partialAmt > 0 {
				baseAmount = partialAmt
				log.Printf("Service %s using partial amount: %.2f (original: %.2f)",
					serviceKey, partialAmt, service.FinalPrice)
			}
		}

		if service.HasComplaintAdjustment {
			settlementAmount -= service.PendingDeductionAmount
			log.Printf("Service %s complaint adjustment: deducting %.2f",
				serviceKey, service.PendingDeductionAmount)
			continue
		}

		commission := baseAmount * (payout.CommissionPercent / 100)
		afterCommission := baseAmount - commission
		gst := afterCommission * (payout.GSTPercent / 100)
		net := afterCommission - gst

		log.Printf("Service %s: Base=%.2f Commission=%.2f GST=%.2f Net=%.2f",
			serviceKey, baseAmount, commission, gst, net)

		settlementAmount += net
	}

	if settlementAmount <= 0 {
		return nil, fmt.Errorf("settlement amount is %.2f, cannot process settlement", settlementAmount)
	}

	now := time.Now()

	settlement := &domain.ProviderSettlement{
		SettlementID:  time.Now().UnixMilli(),
		PayoutID:      payout.ID,
		ProviderID:    provider.ID,
		ProviderName:  provider.Name,
		AccountNo:     provider.BankDetails.AccountNumber,
		IfscCode:      provider.BankDetails.IfscCode,
		TotalAmount:   utils.RoundTo2(settlementAmount),
		PaymentMode:   req.PaymentMode,
		PaymentMethod: req.PaymentMethod,
		Justification: req.Justification,
		Status:        "pending",
		SettledAt:     &now,
		CreatedAt:     now,
	}

	if err := s.settlementRepo.Create(ctx, settlement); err != nil {
		return nil, fmt.Errorf("failed to create settlement")
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
		if service.HasComplaintAdjustment && service.IsSettled {
			if err := s.serviceRepo.MarkAsSettledAfterComplaint(ctx, service.ID, &now); err != nil {
				log.Printf("Failed to mark service %s as settled after complaint: %v", service.ID.Hex(), err)
			}
		}
	}

	unsettledCount, err := s.serviceRepo.CountUnsettledByIDs(ctx, payout.ServiceIDs)
	if err != nil {
		return nil, err
	}
	log.Println("Unsettled service count:", unsettledCount)

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
			return nil, 0, 0, fmt.Errorf("invalid tab value. Use 'pending' or 'settled'")
		}
	}

	if req.Status != "" {
		filter["status"] = req.Status
	}

	if req.ProviderID != "" {
		providerID, err := primitive.ObjectIDFromHex(req.ProviderID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid provider id")
		}
		filter["providerId"] = providerID
	}

	if req.PayoutID != "" {
		payoutID, err := primitive.ObjectIDFromHex(req.PayoutID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid payout id")
		}
		filter["payoutId"] = payoutID
	}

	if req.Status != "" {
		filter["status"] = req.Status
	}

	if req.PaymentMode != "" {
		filter["paymentMode"] = req.PaymentMode
	}

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
		}
		filter["createdAt"] = bson.M{"$gte": startDate}
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
		}
		endDate = endDate.Add(24 * time.Hour)
		if _, exists := filter["createdAt"]; exists {
			filter["createdAt"].(bson.M)["$lte"] = endDate
		} else {
			filter["createdAt"] = bson.M{"$lte": endDate}
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
		return nil, 0, 0, fmt.Errorf("failed to get settlements: %v", err)
	}

	var responses []ProviderSettlementResponse

	for _, settlement := range settlements {
		resp := ProviderSettlementResponse{
			ID:            settlement.ID,
			SettlementID:  "SET" + strconv.FormatInt(settlement.SettlementID, 10),
			ProviderID:    settlement.ProviderID,
			ProviderName:  settlement.ProviderName,
			AccountNo:     settlement.AccountNo,
			IfscCode:      settlement.IfscCode,
			TotalAmount:   utils.RoundTo2(settlement.TotalAmount),
			PaymentMode:   settlement.PaymentMode,
			PaymentMethod: settlement.PaymentMethod,
			Justification: settlement.Justification,

			Status:    settlement.Status,
			SettledAt: settlement.SettledAt,
			CreatedAt: settlement.CreatedAt,
		}

		payout, err := s.payoutRepo.FindByID(ctx, settlement.PayoutID)
		if err == nil && payout != nil {
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
	settlementIDStr string,
	req *CreateSettlementRequest,
) (*domain.ProviderSettlement, error) {

	settlementID, err := primitive.ObjectIDFromHex(settlementIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid settlement id")
	}

	settlement, err := s.settlementRepo.FindByID(ctx, settlementID)
	if err != nil {
		return nil, fmt.Errorf("payout not found")
	}

	if settlement.Status == domain.SettleStatusSettled {
		return nil, fmt.Errorf("payout already settled")
	}

	now := time.Now()
	settlement.SettledPostData = &domain.SettledPostData{
		TransactionID: req.TransactionID,
		Amount:        utils.RoundTo2(req.Amount),
		Status:        string(domain.SettleStatusSettled),
		Note:          req.Note,
	}

	settlement.Status = domain.SettleStatusSettled

	if err := s.settlementRepo.UpdateSettlementPostData(
		ctx,
		settlement.ID,
		settlement.SettledPostData,
		domain.SettleStatusSettled,
		now,
	); err != nil {
		return nil, fmt.Errorf("failed to update settlement: %w", err)
	}

	return settlement, nil
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

	return &resp, nil
}

func (s *SettlementService) GetSettlementsByIDs(
	ctx context.Context,
	ids []string,
) ([]domain.ProviderSettlement, error) {
	return s.settlementRepo.FindBySettlementIDs(ctx, ids)
}
