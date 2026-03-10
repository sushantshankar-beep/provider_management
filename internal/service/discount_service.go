package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DiscountService struct {
	DiscountRepo        *repository.DiscountRepo
	AcceptedServiceRepo *repository.AcceptedServiceRepo
	UserRepo            *repository.UserRepo
}

func NewDiscountService(DiscountRepo *repository.DiscountRepo, AcceptedServiceRepo *repository.AcceptedServiceRepo, UserRepo *repository.UserRepo) *DiscountService {
	return &DiscountService{
		DiscountRepo:        DiscountRepo,
		AcceptedServiceRepo: AcceptedServiceRepo,
		UserRepo:            UserRepo,
	}
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

	now := time.Now().UTC()

	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_at format, use RFC3339")
		}
		if t.Before(now) {
			return nil, fmt.Errorf("start_at cannot be in the past")
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
		if t.Before(d.StartAt) {
			return nil, fmt.Errorf("end_at must be after start_at")
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
	search, status, discountType, scope, validityStart, validityEnd, sortBy, sortOrder string,
) ([]dto.DiscountListResponse, int64, int64, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if status != "" {
		filter["status"] = status
	}
	if scope != "" {
		filter["scope"] = scope
	}
	if discountType != "" {
		filter["type"] = discountType
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	if  validityStart != "" {
		if t, err := time.Parse("2006-01-02", validityStart); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			filter["startAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(t),
				"$lte": primitive.NewDateTimeFromTime(end),
			}
		}
	}

	if  validityEnd != "" {
		if t, err := time.Parse("2006-01-02", validityEnd); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			filter["endAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(t),
				"$lte": primitive.NewDateTimeFromTime(end),
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

	discounts, total, err := s.DiscountRepo.GetDiscounts(ctx, filter, skip, limit, sortBy,validityStart, order)
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

	now := time.Now().UTC()

	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_at format, use RFC3339")
		}
		if t.Before(now) {
			return nil, fmt.Errorf("start_at cannot be in the past")
		}
		update["startAt"] = t
	}
	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, fmt.Errorf("invalid end_at format, use RFC3339")
		}
		if t.Before(now) {
			return nil, fmt.Errorf("end_at cannot be in the past")
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
	d.Status = resolveDiscountStatus(d)
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
		Code:                   d.Code,
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
	d.Status = resolveDiscountStatus(d)
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
		Code:          d.Code,
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

func resolveDiscountStatus(d domain.Discount) domain.DiscountStatus {
	now := time.Now().UTC()

	if d.Status == domain.DiscountStatusDraft || d.Status == domain.DiscountStatusInActive || d.Status == domain.DiscountStatusPaused {
		return d.Status
	}

	if d.EndAt != nil && now.After(*d.EndAt) {
		return domain.DiscountStatusExpired
	}

	if now.Before(d.StartAt) {
		return domain.DiscountStatusScheduled
	}

	if d.Status == domain.DiscountStatusScheduled && !now.Before(d.StartAt) {
		return domain.DiscountStatusActive
	}

	return d.Status
}

func (s *DiscountService) SyncDiscountStatuses(ctx context.Context) error {
	now := time.Now().UTC()
	if err := s.DiscountRepo.BulkUpdateExpired(ctx, now); err != nil {
		return err
	}
	return s.DiscountRepo.BulkActivateScheduled(ctx, now)
}
func (s *DiscountService) ListDiscountsWithUsage(
	ctx context.Context,
	page, limit int64,
	search, status, createdAt string,
) ([]dto.DiscountUsageListItem, int64, int64, error) {

	skip := (page - 1) * limit

	discounts, total, err := s.DiscountRepo.GetDiscounts(ctx, bson.M{}, skip, limit,  "createdAt", createdAt, -1)
	if err != nil {
		return nil, 0, 0, err
	}

	ids := make([]string, 0, len(discounts))
	for _, d := range discounts {
		ids = append(ids, d.ID.Hex())
	}

	stats, err := s.AcceptedServiceRepo.GetDiscountUsageStats(ctx, ids)
	if err != nil {
		return nil, 0, 0, err
	}

	result := make([]dto.DiscountUsageListItem, 0, len(discounts))

	for _, d := range discounts {

		stat := stats[d.ID.Hex()]

		uniqueUsers, _ := s.AcceptedServiceRepo.GetDiscountUserCount(ctx, d.ID.Hex())

		item := dto.DiscountUsageListItem{
			DiscountID:         d.ID.Hex(),
			Code:               d.Code,
			Name:               d.Name,
			TotalRedemptions:   stat.TotalOrders,
			UniqueUsers:        uniqueUsers,
			TotalDiscountGiven: stat.TotalSavings,
			CreatedAt:          d.CreatedAt.Format(time.RFC3339),
		}

		result = append(result, item)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return result, total, totalPages, nil
}

func (s *DiscountService) GetDiscountUserUsage(
	ctx context.Context,
	discountID, userID, createdAt string,
	page, limit int,
) (interface{}, int64, error) {
	skip := int64((page - 1) * limit)
	lim := int64(limit)
  
	if userID != "" {
		services, total, err := s.AcceptedServiceRepo.GetServicesByDiscountAndUser(
			ctx, discountID, userID, createdAt, skip, lim,
		)
		if err != nil {
			return nil, 0, err
		}

		result := make([]dto.DiscountServiceUsageItem, 0, len(services))
		for _, svc := range services {
			item := dto.DiscountServiceUsageItem{
				ServiceID:     svc.ID.Hex(),
				ServiceNumber: svc.ServiceNumber,
				TotalDiscount: svc.TotalDiscount,
				AmountPaid:    svc.AmountPaidByUser,
				ServiceType:   svc.ServiceType,
				CreatedAt:     svc.CreatedAt.Format(time.RFC3339),
			}
			if svc.AppliedDiscount != nil {
				item.DiscountCode = svc.AppliedDiscount.Code
				item.DiscountAmount = svc.AppliedDiscount.DiscountAmt
			}
			result = append(result, item)
		}
		return result, total, nil
	}

	rows, total, err := s.AcceptedServiceRepo.GetDiscountUsersAggregated(
		ctx, discountID,createdAt, skip, lim,
	)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID.Hex())
	}

	userMap := map[string]*domain.User{}
	if s.UserRepo != nil && len(userIDs) > 0 {
		users, err := s.UserRepo.FindByIDs(ctx, userIDs)
		if err == nil {
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}

	result := make([]dto.DiscountUserUsageItem, 0, len(rows))
	for _, row := range rows {
		item := dto.DiscountUserUsageItem{
			UserID:        row.UserID.Hex(),
			UsageCount:    row.UsageCount,
			TotalDiscount: row.TotalDiscount,
			LastUsedAt:    row.LastUsedAt.Format(time.RFC3339),
		}
		if u, ok := userMap[row.UserID.Hex()]; ok {
			item.UserName = u.Name
			item.UserPhone = u.Phone
		}
		result = append(result, item)
	}

	return result, total, nil
}