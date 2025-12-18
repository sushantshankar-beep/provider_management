package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"
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
	numStr := strings.TrimPrefix(payoutIDStr, "SET")
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

	now := time.Now()

	settlement := &domain.ProviderSettlement{
		SettlementID:  time.Now().UnixMilli(),
		PayoutID:      payout.ID,
		ProviderID:    provider.ID,
		ProviderName:  provider.Name,
		AccountNo:     provider.BankDetails.AccountNumber,
		IfscCode:      provider.BankDetails.IfscCode,
		TotalAmount:   payout.NetPayable,
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
	if err := s.serviceRepo.MarkAsSettled(ctx, serviceObjIDs, settlement.ID); err != nil {
		return nil, err
	}
	_ = s.payoutRepo.MarkSettled(ctx, payout.ID, settlement.ID)

	return settlement, nil
}

func (s *SettlementService) GetSettlements(
	ctx context.Context,
	req *GetSettlementsRequest,
) ([]domain.ProviderSettlement, int64, int64, error) {

	filter := bson.M{}

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

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return settlements, total, totalPages, nil
}
