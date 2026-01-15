package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"time"
    "log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AMCPlanService struct {
	AMCPlanRepo *repository.AMCPlanRepo
}

func NewAMCPlanService(amcPlanRepo *repository.AMCPlanRepo) *AMCPlanService {
	return &AMCPlanService{
		AMCPlanRepo: amcPlanRepo,
	}
}

func (s *AMCPlanService) CreateAMC(ctx context.Context, plan *domain.AMCPlan, adminID string) error {
	if plan.PlanName == "" {
		return fmt.Errorf("plan name is required")
	}
	if plan.PlanSlug == "" {
		return fmt.Errorf("plan slug is required")
	}
	if plan.PlanVehicleType == "" {
		return fmt.Errorf("plan vehicle type is required")
	}
	if plan.PlanCategory == "" {
		return fmt.Errorf("plan category is required")
	}

	// Check if slug already exists
	existingPlan, err := s.AMCPlanRepo.FindBySlug(ctx, plan.PlanSlug)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingPlan != nil {
		return fmt.Errorf("plan slug already exists")
	}

	// Set creator
	objID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID")
	}
	plan.CreatedBy = objID
	plan.LastModifiedBy = objID

	return s.AMCPlanRepo.Create(ctx, plan)
}

func (s *AMCPlanService) GetAllAMC(
	ctx context.Context,
	page, limit int64,
	isActive, planStatus, planVehicleType, planCategory, createdDate,validityPeriod, search string,
) ([]map[string]any, map[string]int64, map[string]any, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if isActive != "" {
		filter["isActive"] = isActive == "true"
	}

	if planStatus != "" {
		filter["planStatus"] = planStatus
	}

	if planCategory != "" {
		filter["planCategory"] = planCategory
	}

	if planVehicleType != "" {
		filter["planVehicleType"] = planVehicleType
	}

	if validityPeriod != "" {
        switch validityPeriod {
        case "6":
            filter["planDurationInMonth"] = 6
        case "12":
            filter["planDurationInMonth"] = 12
        case "24":
            filter["planDurationInMonth"] = 24
        case "36":
            filter["planDurationInMonth"] = 36
        default:
            if duration, err := strconv.Atoi(validityPeriod); err == nil {
                filter["planDurationInMonth"] = duration
            }
        }
    }

	if createdDate != "" {
		var parsedDate time.Time
		var err error
		
		parsedDate, err = time.Parse("2006-01-02", createdDate)
		if err != nil {
			parsedDate, err = time.Parse("02-01-2006", createdDate)
		}
		
		if err == nil {
			startOfDay := time.Date(
				parsedDate.Year(),
				parsedDate.Month(),
				parsedDate.Day(),
				0, 0, 0, 0,
				time.UTC,
			)
			
			endOfDay := time.Date(
				parsedDate.Year(),
				parsedDate.Month(),
				parsedDate.Day(),
				23, 59, 59, 999999999,
				time.UTC,
			)
			
			filter["createdAt"] = bson.M{
				"$gte": startOfDay,
				"$lte": endOfDay,
			}
			log.Printf("Searching for date range: %s to %s\n", startOfDay.Format(time.RFC3339), endOfDay.Format(time.RFC3339))
		} else {
			log.Printf("Error parsing date %s: %v\n", createdDate, err)
		}
	}

	if search != "" {
		filter["$or"] = []bson.M{
			{"planName": bson.M{"$regex": search, "$options": "i"}},
			{"planStatus": bson.M{"$regex": search, "$options": "i"}},
			{"planCategory": bson.M{"$regex": search, "$options": "i"}},
			{"planVehicleType": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	plans, _, err := s.AMCPlanRepo.GetPlans(ctx, filter, skip, limit, "createdAt", -1)
	if err != nil {
		return nil, nil, nil, err
	}

	filteredTotal, _ := s.AMCPlanRepo.CountDocuments(ctx, filter)

	totalPlans, _ := s.AMCPlanRepo.CountDocuments(ctx, bson.M{})
	activePlans, _ := s.AMCPlanRepo.CountDocuments(ctx, bson.M{"isActive": true})
	inactivePlans, _ := s.AMCPlanRepo.CountDocuments(ctx, bson.M{"isActive": false})

	responseData := make([]map[string]any, len(plans))
	for i, plan := range plans {
		responseData[i] = map[string]any{
			"_id":                    plan.ID.Hex(),
			"plan_name":              plan.PlanName,
			"plan_category":          plan.PlanCategory,
			"plan_vehicle_type":      plan.PlanVehicleType,
			"plan_total_amount":      plan.PlanTotalAmount,
			"plan_duration_in_month": plan.PlanDurationInMonth,
			"plan_base_price":        plan.PlanBasePrice,
			"created_at":             plan.CreatedAt,
			"is_active":              plan.IsActive,
		}
	}

	counts := map[string]int64{
		"total_plans":    totalPlans,
		"active_plans":   activePlans,
		"inactive_plans": inactivePlans,
	}

	totalPages := filteredTotal / limit
	if filteredTotal%limit > 0 {
		totalPages++
	}

	pagination := map[string]any{
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"total_items": filteredTotal,
	}

	return responseData, counts, pagination, nil
}

func (s *AMCPlanService) GetAMCByID(ctx context.Context, id string) (map[string]any, error) {
	plan, err := s.AMCPlanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"_id":                         plan.ID.Hex(),
		"plan_name":                   plan.PlanName,
		"plan_slug":                   plan.PlanSlug,
		"plan_vehicle_type":           plan.PlanVehicleType,
		"plan_category":               plan.PlanCategory,
		"plan_description":            plan.PlanDescription,
		"plan_duration_in_month":      plan.PlanDurationInMonth,
		"plan_start":                  plan.PlanStart,
		"plan_base_price":             plan.PlanBasePrice,
		"discount_percent":            plan.DiscountPercent,
		"plan_discount_amount":        plan.PlanDiscountAmount,
		"plan_price_after_discount":   plan.PlanPriceAfterDiscount,
		"plan_gst_percent":            plan.PlanGSTPercent,
		"plan_gst_amount":             plan.PlanGSTAmount,
		"plan_total_amount":           plan.PlanTotalAmount,
		"allowed_brand_models":        plan.AllowedBrandModels,
		"plan_city":                   plan.PlanCity,
		"plan_services_included":      plan.PlanServicesIncluded,
		"plan_tags":                   plan.PlanTags,
		"plan_features":               plan.PlanFeatures,
		"sorting":                     plan.Sorting,
		"is_active":                   plan.IsActive,
		"created_by":                  plan.CreatedBy.Hex(),
		"last_modified_by":            plan.LastModifiedBy.Hex(),
		"created_at":                  plan.CreatedAt,
		"updated_at":                  plan.UpdatedAt,
	}, nil
}

func (s *AMCPlanService) UpdateAMC(ctx context.Context, id string, updateData map[string]any, adminID string) error {
	if len(updateData) == 0 {
		return fmt.Errorf("no update data provided")
	}

	if slug, ok := updateData["plan_slug"]; ok {
		existingPlan, err := s.AMCPlanRepo.FindBySlug(ctx, slug.(string))
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		if existingPlan != nil && existingPlan.ID.Hex() != id {
			return fmt.Errorf("plan slug already exists")
		}
	}

	update := bson.M{}

	fieldMap := map[string]string{
		"plan_name":                   "planName",
		"plan_slug":                   "planSlug",
		"plan_vehicle_type":           "planVehicleType",
		"plan_category":               "planCategory",
		"plan_description":            "planDescription",
		"plan_duration_in_month":      "planDurationInMonth",
		"plan_start":                  "planStart",
		"plan_base_price":             "planBasePrice",
		"discount_percent":            "discountPercent",
		"plan_discount_amount":        "planDiscountAmount",
		"plan_price_after_discount":   "planPriceAfterDiscount",
		"plan_gst_percent":            "planGSTPercent",
		"plan_gst_amount":             "planGSTAmount",
		"plan_total_amount":           "planTotalAmount",
		"allowed_brand_models":        "allowedBrandModels",
		"plan_city":                   "planCity",
		"plan_services_included":      "planServicesIncluded",
		"plan_tags":                   "planTags",
		"plan_features":               "planFeatures",
		"sorting":                     "sorting",
		"plan_status":                 "planStatus",
		"is_active":                   "isActive",
	}

	for key, val := range updateData {
		if dbKey, ok := fieldMap[key]; ok {
			update[dbKey] = val
		}
	}

	objID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID")
	}
	update["lastModifiedBy"] = objID

	return s.AMCPlanRepo.Update(ctx, id, update)
}

func (s *AMCPlanService) DeleteAMC(ctx context.Context, id string) error {
	return s.AMCPlanRepo.Delete(ctx, id)
}

func (s *AMCPlanService) ToggleAMCStatus(ctx context.Context, id string, adminID string) (bool, error) {
	plan, err := s.AMCPlanRepo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}

	newStatus := !plan.IsActive

	update := bson.M{
		"isActive": newStatus,
	}

	objID, err := primitive.ObjectIDFromHex(adminID)
	if err == nil {
		update["lastModifiedBy"] = objID
	}

	err = s.AMCPlanRepo.Update(ctx, id, update)
	if err != nil {
		return false, err
	}

	return newStatus, nil
}

func (s *AMCPlanService) GetAMCPlansByCategory(
	ctx context.Context,
	vehicleType, category, cityName string,
) ([]map[string]any, error) {
	if cityName == "" {
		return nil, fmt.Errorf("city name is required")
	}

	plans, err := s.AMCPlanRepo.FindByCategory(ctx, vehicleType, category, cityName)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf("no AMC plans found for the given category and vehicle type in this city")
	}

	responseData := make([]map[string]any, len(plans))
	for i, plan := range plans {
		responseData[i] = map[string]any{
			"_id":                        plan.ID.Hex(),
			"plan_name":                  plan.PlanName,
			"plan_slug":                  plan.PlanSlug,
			"plan_vehicle_type":          plan.PlanVehicleType,
			"plan_category":              plan.PlanCategory,
			"plan_description":           plan.PlanDescription,
			"plan_duration_in_month":     plan.PlanDurationInMonth,
			"plan_start":                 plan.PlanStart,
			"plan_base_price":            plan.PlanBasePrice,
			"discount_percent":           plan.DiscountPercent,
			"plan_discount_amount":       plan.PlanDiscountAmount,
			"plan_price_after_discount":  plan.PlanPriceAfterDiscount,
			"plan_gst_percent":           plan.PlanGSTPercent,
			"plan_gst_amount":            plan.PlanGSTAmount,
			"plan_total_amount":          plan.PlanTotalAmount,
			"plan_services_included":     plan.PlanServicesIncluded,
			"plan_tags":                  plan.PlanTags,
			"plan_features":              plan.PlanFeatures,
			"sorting":                    plan.Sorting,
			"is_active":                  plan.IsActive,
		}
	}

	return responseData, nil
}