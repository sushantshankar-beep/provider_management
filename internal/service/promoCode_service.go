package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PromoCodeService struct {
	PromoCodeRepo *repository.PromoCodeRepo
}

func NewPromoCodeService(PromoCodeRepo *repository.PromoCodeRepo) *PromoCodeService {
	return &PromoCodeService{PromoCodeRepo: PromoCodeRepo}
}

func (s *PromoCodeService) CreatePromoCode(ctx context.Context, req dto.CreatePromoCodeRequest) (*dto.PromoCodeResponse, error) {
	if !domain.IsValidDiscountType(domain.DiscountType(req.DiscountType)) {
		return nil, fmt.Errorf("invalid discount type")
	}

	if req.Value <= 0 {
		return nil, fmt.Errorf("discount value must be greater than 0")
	}

	if domain.DiscountType(req.DiscountType) == domain.DiscountPercent && req.Value > 100 {
		return nil, fmt.Errorf("percent discount cannot exceed 100")
	}

	if len(req.ServiceTypes) == 0 {
		return nil, fmt.Errorf("at least one service type is required")
	}

	for _, st := range req.ServiceTypes {
		if !domain.IsValidPromoServiceType(domain.PromoServiceType(st)) {
			return nil, fmt.Errorf("invalid service type: %s", st)
		}
	}

	userEligibility := domain.UserEligibility(req.UserEligibility)
	if userEligibility == "" {
		userEligibility = domain.UserEligibilityAll
	}
	if !domain.IsValidUserEligibility(userEligibility) {
		return nil, fmt.Errorf("invalid user eligibility")
	}

	status := domain.PromoStatus(req.Status)
	if status == "" {
		status = domain.PromoStatusDraft
	}
	if !domain.IsValidPromoStatus(status) {
		return nil, fmt.Errorf("invalid status")
	}

	existing, err := s.PromoCodeRepo.FindByCode(ctx, req.Code)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("promo code already exists")
	}

	serviceTypes := make([]domain.PromoServiceType, len(req.ServiceTypes))
	for i, st := range req.ServiceTypes {
		serviceTypes[i] = domain.PromoServiceType(st)
	}

	promo := &domain.PromoCode{
		Code:                req.Code,
		Title:               req.Title,
		Description:         req.Description,
		DiscountType:        domain.DiscountType(req.DiscountType),
		Value:               req.Value,
		MaxDiscount:         req.MaxDiscount,
		MinOrderValue:       req.MinOrderValue,
		PerUserLimit:        req.PerUserLimit,
		GlobalRedemptionCap: req.GlobalRedemptionCap,
		ServiceTypes:        serviceTypes,
		Zones:               req.Zones,
		PaymentMethods:      req.PaymentMethods,
		UserEligibility:     userEligibility,
		AllowStacking:       req.AllowStacking,
		Status:              status,
		CreatedBy:           req.CreatedBy,
	}

	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_at format, use RFC3339")
		}
		promo.StartAt = t
	} else {
		promo.StartAt = time.Now()
	}

	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, fmt.Errorf("invalid end_at format, use RFC3339")
		}
		promo.EndAt = &t
	}

	if err := s.PromoCodeRepo.Create(ctx, promo); err != nil {
		return nil, err
	}

	resp := s.mapToPromoCodeResponse(*promo)
	return &resp, nil
}

func (s *PromoCodeService) GetPromoCodes(
	ctx context.Context,
	page, limit int64,
	search, status, serviceType, sortBy, sortOrder string,
) ([]dto.PromoCodeListResponse, int64, int64, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if status != "" {
		filter["status"] = status
	}
	if serviceType != "" {
		filter["serviceTypes"] = serviceType
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"code": bson.M{"$regex": search, "$options": "i"}},
			{"title": bson.M{"$regex": search, "$options": "i"}},
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
		case "code":
			sortBy = "code"
		case "title":
			sortBy = "title"
		case "value":
			sortBy = "value"
		case "created_at":
			sortBy = "createdAt"
		case "updated_at":
			sortBy = "updatedAt"
		}
	}

	promos, total, err := s.PromoCodeRepo.GetPromoCodes(ctx, filter, skip, limit, sortBy, order)
	if err != nil {
		return nil, 0, 0, err
	}

	result := make([]dto.PromoCodeListResponse, len(promos))
	for i, p := range promos {
		result[i] = s.mapToPromoCodeListResponse(p)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return result, total, totalPages, nil
}

func (s *PromoCodeService) GetPromoCodeByID(ctx context.Context, id string) (*dto.PromoCodeResponse, error) {
	promo, err := s.PromoCodeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.mapToPromoCodeResponse(*promo)
	return &resp, nil
}

func (s *PromoCodeService) UpdatePromoCode(ctx context.Context, id string, req dto.UpdatePromoCodeRequest, updatedBy string) (*dto.PromoCodeResponse, error) {
	update := bson.M{}

	if req.Code != "" {
		update["code"] = req.Code
	}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.DiscountType != "" {
		if !domain.IsValidDiscountType(domain.DiscountType(req.DiscountType)) {
			return nil, fmt.Errorf("invalid discount type")
		}
		update["discountType"] = req.DiscountType
	}
	if req.Value > 0 {
		update["value"] = req.Value
	}
	if req.MaxDiscount > 0 {
		update["maxDiscount"] = req.MaxDiscount
	}
	if req.MinOrderValue > 0 {
		update["minOrderValue"] = req.MinOrderValue
	}
	if req.PerUserLimit > 0 {
		update["perUserLimit"] = req.PerUserLimit
	}
	if req.GlobalRedemptionCap > 0 {
		update["globalRedemptionCap"] = req.GlobalRedemptionCap
	}
	if len(req.ServiceTypes) > 0 {
		update["serviceTypes"] = req.ServiceTypes
	}
	if len(req.Zones) > 0 {
		update["zones"] = req.Zones
	}
	if len(req.PaymentMethods) > 0 {
		update["paymentMethods"] = req.PaymentMethods
	}
	if req.UserEligibility != "" {
		if !domain.IsValidUserEligibility(domain.UserEligibility(req.UserEligibility)) {
			return nil, fmt.Errorf("invalid user eligibility")
		}
		update["userEligibility"] = req.UserEligibility
	}
	if req.AllowStacking != nil {
		update["allowStacking"] = *req.AllowStacking
	}
	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_at format, use RFC3339")
		}
		update["startAt"] = t
	}
	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, fmt.Errorf("invalid end_at format, use RFC3339")
		}
		update["endAt"] = t
	}

	if len(update) == 0 {
		return nil, fmt.Errorf("no update data provided")
	}

	if err := s.PromoCodeRepo.Update(ctx, id, update); err != nil {
		return nil, err
	}

	promo, err := s.PromoCodeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.mapToPromoCodeResponse(*promo)
	return &resp, nil
}

func (s *PromoCodeService) UpdatePromoCodeStatus(ctx context.Context, id string, req dto.UpdatePromoCodeStatusRequest, updatedBy string) error {
	validStatuses := map[string]bool{
		"active": true, "inactive": true, "draft": true, "scheduled": true, "expired": true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: must be one of active, inactive, draft, scheduled, expired")
	}

	if err := s.PromoCodeRepo.UpdateStatus(ctx, id, req.Status); err != nil {
		return err
	}

	return nil
}

func (s *PromoCodeService) DeletePromoCode(ctx context.Context, id string) error {
	return s.PromoCodeRepo.Delete(ctx, id)
}

func (s *PromoCodeService) GetPromoCodeStats(ctx context.Context) (*dto.PromoCodeStatsResponse, error) {
	stats, err := s.PromoCodeRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.PromoCodeStatsResponse{
		TotalPromoCodes: stats["total"],
		ActivePromos:    stats["active"],
		ScheduledPromos: stats["scheduled"],
		ExpiredPromos:   stats["expired"],
		Drafts:          stats["draft"],
	}, nil
}


func (s *PromoCodeService) mapToPromoCodeResponse(p domain.PromoCode) dto.PromoCodeResponse {
	discount := fmt.Sprintf("%.0f%%", p.Value)
	if p.DiscountType == domain.DiscountFlat {
		discount = fmt.Sprintf("₹%.0f", p.Value)
	}
	if p.MaxDiscount > 0 && p.DiscountType == domain.DiscountPercent {
		discount = fmt.Sprintf("%.0f%% (max ₹%.0f)", p.Value, p.MaxDiscount)
	}

	serviceTypes := make([]string, len(p.ServiceTypes))
	for i, st := range p.ServiceTypes {
		serviceTypes[i] = string(st)
	}
	serviceType := "All Services"
	if len(serviceTypes) == 1 {
		serviceType = serviceTypes[0]
	}

	return dto.PromoCodeResponse{
		ID:                  p.ID.Hex(),
		Code:                p.Code,
		Title:               p.Title,
		Description:         p.Description,
		DiscountType:        string(p.DiscountType),
		Value:               p.Value,
		Discount:            discount,
		MaxDiscount:         p.MaxDiscount,
		MinOrderValue:       p.MinOrderValue,
		PerUserLimit:        p.PerUserLimit,
		GlobalRedemptionCap: p.GlobalRedemptionCap,
		ServiceTypes:        serviceTypes,
		ServiceType:         serviceType,
		Zones:               p.Zones,
		PaymentMethods:      p.PaymentMethods,
		UserEligibility:     string(p.UserEligibility),
		AllowStacking:       p.AllowStacking,
		Status:              string(p.Status),
		UsageCount:          p.UsageCount,
		Usage:               fmt.Sprintf("%d/%d", p.UsageCount, p.GlobalRedemptionCap),
		TotalDiscount:       p.TotalDiscount,
		CreatedBy:           p.CreatedBy,
		StartAt:             p.StartAt.Format(time.RFC3339),
		EndAt:               dto.FormatEndAt(p.EndAt),
		CreatedAt:           p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           p.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *PromoCodeService) mapToPromoCodeListResponse(p domain.PromoCode) dto.PromoCodeListResponse {
	discount := fmt.Sprintf("%.0f%%", p.Value)
	if p.DiscountType == domain.DiscountFlat {
		discount = fmt.Sprintf("₹%.0f", p.Value)
	}
	if p.MaxDiscount > 0 && p.DiscountType == domain.DiscountPercent {
		discount = fmt.Sprintf("%.0f%% (max ₹%.0f)", p.Value, p.MaxDiscount)
	}

	serviceTypes := make([]string, len(p.ServiceTypes))
	for i, st := range p.ServiceTypes {
		serviceTypes[i] = string(st)
	}
	serviceType := "All Services"
	if len(serviceTypes) == 1 {
		serviceType = serviceTypes[0]
	}

	return dto.PromoCodeListResponse{
		ID:            p.ID.Hex(),
		Code:          p.Code,
		Title:         p.Title,
		Discount:      discount,
		ServiceType:   serviceType,
		ServiceTypes:  serviceTypes,
		ValidityStart: p.StartAt.Format(time.RFC3339),
		ValidityEnd:   dto.FormatEndAt(p.EndAt),
		Status:        string(p.Status),
		Usage:         fmt.Sprintf("%d/%d", p.UsageCount, p.GlobalRedemptionCap),
		TotalDiscount: fmt.Sprintf("₹%.3f", p.TotalDiscount),
		CreatedBy:     p.CreatedBy,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}