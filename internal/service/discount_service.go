package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type DiscountService struct {
	DiscountRepo *repository.DiscountRepo
}

func NewDiscountService(DiscountRepo *repository.DiscountRepo) *DiscountService {
	return &DiscountService{DiscountRepo: DiscountRepo}
}

func (s *DiscountService) CreateDiscount(ctx context.Context, req dto.CreateDiscountRequest) (*dto.DiscountResponse, error) {
	if !domain.IsValidDiscountType(domain.DiscountType(req.Type)) {
		return nil, fmt.Errorf("invalid discount type")
	}

	if req.Value <= 0 {
		return nil, fmt.Errorf("discount value must be greater than 0")
	}

	if domain.DiscountType(req.Type) == domain.DiscountPercent && req.Value > 100 {
		return nil, fmt.Errorf("percent discount cannot exceed 100")
	}

	if !domain.IsValidDiscountScope(domain.DiscountScope(req.Scope)) {
		return nil, fmt.Errorf("invalid scope")
	}

	if len(req.ApplicableOn) == 0 {
		return nil, fmt.Errorf("applicable_on is required")
	}

	for _, a := range req.ApplicableOn {
		if !domain.IsValidDiscountApplicableOn(domain.DiscountApplicableOn(a)) {
			return nil, fmt.Errorf("invalid applicable_on value: %s", a)
		}
	}

	userEligibility := domain.UserEligibility(req.UserEligibility)
	if userEligibility == "" {
		userEligibility = domain.UserEligibilityAll
	}
	if !domain.IsValidUserEligibility(userEligibility) {
		return nil, fmt.Errorf("invalid user eligibility")
	}

	status := domain.DiscountStatus(req.Status)
	if status == "" {
		status = domain.DiscountStatusDraft
	}
	if !domain.IsValidDiscountStatus(status) {
		return nil, fmt.Errorf("invalid status")
	}

	applicableOn := make([]domain.DiscountApplicableOn, len(req.ApplicableOn))
	for i, a := range req.ApplicableOn {
		applicableOn[i] = domain.DiscountApplicableOn(a)
	}

	d := &domain.Discount{
	    Code:                   req.Code,
		Name:                   req.Name,
		Description:            req.Description,
		Type:                   domain.DiscountType(req.Type),
		Value:                  req.Value,
		MaxDiscount:            req.MaxDiscount,
		Scope:                  domain.DiscountScope(req.Scope),
		ApplicableOn:           applicableOn,
		Zones:                  req.Zones,
		PaymentMethods:         req.PaymentMethods,
		UserEligibility:        userEligibility,
		AllowStackingWithPromo: req.AllowStackingWithPromo,
		Status:                 status,
		CreatedBy:              req.CreatedBy,
	}

	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_at format, use RFC3339")
		}
		d.StartAt = t
	} else {
		d.StartAt = time.Now()
	}

	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, fmt.Errorf("invalid end_at format, use RFC3339")
		}
		d.EndAt = &t
	}

	if err := s.DiscountRepo.Create(ctx, d); err != nil {
		return nil, err
	}

	resp := s.mapToDiscountResponse(*d)
	return &resp, nil
}

func (s *DiscountService) GetDiscounts(
	ctx context.Context,
	page, limit int64,
	search, status, scope, sortBy, sortOrder string,
) ([]dto.DiscountListResponse, int64, int64, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if status != "" {
		filter["status"] = status
	}
	if scope != "" {
		filter["scope"] = scope
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
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
		case "name":
			sortBy = "name"
		case "value":
			sortBy = "value"
		case "created_at":
			sortBy = "createdAt"
		case "updated_at":
			sortBy = "updatedAt"
		}
	}

	discounts, total, err := s.DiscountRepo.GetDiscounts(ctx, filter, skip, limit, sortBy, order)
	if err != nil {
		return nil, 0, 0, err
	}

	result := make([]dto.DiscountListResponse, len(discounts))
	for i, d := range discounts {
		result[i] = s.mapToDiscountListResponse(d)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return result, total, totalPages, nil
}

func (s *DiscountService) GetDiscountByID(ctx context.Context, id string) (*dto.DiscountResponse, error) {
	d, err := s.DiscountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.mapToDiscountResponse(*d)
	return &resp, nil
}

func (s *DiscountService) UpdateDiscount(ctx context.Context, id string, req dto.UpdateDiscountRequest, updatedBy string) (*dto.DiscountResponse, error) {
	update := bson.M{}

	if req.Code != "" {
		update["code"] = req.Code
	}
	
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Type != "" {
		if !domain.IsValidDiscountType(domain.DiscountType(req.Type)) {
			return nil, fmt.Errorf("invalid discount type")
		}
		update["type"] = req.Type
	}
	if req.Value > 0 {
		update["value"] = req.Value
	}
	if req.MaxDiscount > 0 {
		update["maxDiscount"] = req.MaxDiscount
	}
	if req.Scope != "" {
		if !domain.IsValidDiscountScope(domain.DiscountScope(req.Scope)) {
			return nil, fmt.Errorf("invalid scope")
		}
		update["scope"] = req.Scope
	}
	if len(req.ApplicableOn) > 0 {
		update["applicableOn"] = req.ApplicableOn
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
	if req.AllowStackingWithPromo != nil {
		update["allowStackingWithPromo"] = *req.AllowStackingWithPromo
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

	if err := s.DiscountRepo.Update(ctx, id, update); err != nil {
		return nil, err
	}

	d, err := s.DiscountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.mapToDiscountResponse(*d)
	return &resp, nil
}

func (s *DiscountService) UpdateDiscountStatus(ctx context.Context, id string, req dto.UpdateDiscountStatusRequest, updatedBy string) error {
	validStatuses := map[string]bool{
		"active": true, "inactive": true, "draft": true,
		"scheduled": true, "expired": true, "paused": true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: must be one of active, inactive, draft, scheduled, expired, paused")
	}

	if err := s.DiscountRepo.UpdateStatus(ctx, id, req.Status); err != nil {
		return err
	}

	return nil
}

func (s *DiscountService) DeleteDiscount(ctx context.Context, id string) error {
	return s.DiscountRepo.Delete(ctx, id)
}

func (s *DiscountService) GetDiscountStats(ctx context.Context) (*dto.DiscountStatsResponse, error) {
	stats, err := s.DiscountRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.DiscountStatsResponse{
		AllDiscounts:       stats["total"],
		ActiveDiscounts:    stats["active"],
		ScheduledDiscounts: stats["scheduled"],
		ExpiredDiscounts:   stats["expired"],
		Drafts:             stats["draft"],
		PausedDiscounts:    stats["paused"],
	}, nil
}

func (s *DiscountService) mapToDiscountResponse(d domain.Discount) dto.DiscountResponse {
	discount := fmt.Sprintf("%.0f%%", d.Value)
	if d.Type == domain.DiscountFlat {
		discount = fmt.Sprintf("₹%.0f", d.Value)
	}
	if d.MaxDiscount > 0 && d.Type == domain.DiscountPercent {
		discount = fmt.Sprintf("%.0f%% (max ₹%.0f)", d.Value, d.MaxDiscount)
	}

	applicableOn := make([]string, len(d.ApplicableOn))
	for i, a := range d.ApplicableOn {
		applicableOn[i] = string(a)
	}

	return dto.DiscountResponse{
		ID:                     d.ID.Hex(),
		Code: d.Code, 
		Name:                   d.Name,
		Description:            d.Description,
		Type:                   string(d.Type),
		Value:                  d.Value,
		Discount:               discount,
		MaxDiscount:            d.MaxDiscount,
		Scope:                  string(d.Scope),
		ApplicableOn:           applicableOn,
		Zones:                  d.Zones,
		PaymentMethods:         d.PaymentMethods,
		UserEligibility:        string(d.UserEligibility),
		AllowStackingWithPromo: d.AllowStackingWithPromo,
		Status:                 string(d.Status),
		TotalSavings:           d.TotalSavings,
		TotalOrders:            d.TotalOrders,
		CreatedBy:              d.CreatedBy,
		StartAt:                d.StartAt.Format(time.RFC3339),
		EndAt:                  dto.FormatEndAtPtr(d.EndAt),
		CreatedAt:              d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              d.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *DiscountService) mapToDiscountListResponse(d domain.Discount) dto.DiscountListResponse {
	discount := fmt.Sprintf("%.0f%%", d.Value)
	if d.Type == domain.DiscountFlat {
		discount = fmt.Sprintf("₹%.0f", d.Value)
	}
	if d.MaxDiscount > 0 && d.Type == domain.DiscountPercent {
		discount = fmt.Sprintf("%.0f%% (max ₹%.0f)", d.Value, d.MaxDiscount)
	}

	applicableOn := make([]string, len(d.ApplicableOn))
	for i, a := range d.ApplicableOn {
		applicableOn[i] = string(a)
	}

	return dto.DiscountListResponse{
		ID:            d.ID.Hex(),
		Name:          d.Name,
		Description:   d.Description,
		Type:          string(d.Type),
		Value:         d.Value,
		Discount:      discount,
		Scope:         string(d.Scope),
		ApplicableOn:  applicableOn,
		ValidityStart: d.StartAt.Format(time.RFC3339),
		ValidityEnd:   dto.FormatEndAtPtr(d.EndAt),
		Status:        string(d.Status),
		TotalSavings:  d.TotalSavings,
		TotalOrders:   d.TotalOrders,
		CreatedBy:     d.CreatedBy,
		CreatedAt:     d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     d.UpdatedAt.Format(time.RFC3339),
	}
}