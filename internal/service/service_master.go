package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
)

type ServiceService struct {
	ServiceMasterRepo  *repository.ServiceMasterRepo
}

func NewServiceMaster(ServiceMasterRepo *repository.ServiceMasterRepo) *ServiceService {
	return &ServiceService{
		ServiceMasterRepo:  ServiceMasterRepo,
	}
}

func (s *ServiceService) CreateService(ctx context.Context, service *domain.ServiceMaster) error {
	if service.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if !domain.IsValidCategory(service.Category) {
		return fmt.Errorf("invalid category")
	}

	if len(service.VehicleType) == 0 {
		return fmt.Errorf("vehicle type is required")
	}

	if !domain.IsValidVehicleType(service.VehicleType) {
	    return fmt.Errorf("invalid vehicle type: %s", service.VehicleType)
	}



	if service.Location == "" {
		return fmt.Errorf("location is required")
	}

	if service.MinCost < 0 {
		return fmt.Errorf("min cost cannot be negative")
	}

	if service.Status == "" {
		service.Status = domain.StatusActive
	}

	if !domain.IsValidStatus(service.Status) {
		return fmt.Errorf("invalid status")
	}

	return s.ServiceMasterRepo.Create(ctx, service)
}

func (s *ServiceService) GetServices(
	ctx context.Context,
	page, limit int64,
	search, category, vehicleType, location, status, sortBy, sortOrder string,
) ([]map[string]any, int64, int64, error) {
	skip := (page - 1) * limit
	filter := bson.M{}

	if category != "" {
		filter["category"] = category
	}

	if vehicleType != "" {
		filter["vehicleType"] = vehicleType
	}

	if location != "" {
		filter["location"] = location
	}

	if status != "" {
		filter["status"] = status
	}

	if search != "" {
		filter["$or"] = []bson.M{
			{"serviceName": bson.M{"$regex": search, "$options": "i"}},
			{"tag": bson.M{"$regex": search, "$options": "i"}},
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
		case "service_name":
			sortBy = "serviceName"
		case "min_cost":
			sortBy = "minCost"
		case "display_order":
			sortBy = "displayOrder"
		case "created_at":
			sortBy = "createdAt"
		case "updated_at":
			sortBy = "updatedAt"
		}
	}

	services, total, err := s.ServiceMasterRepo.GetServices(ctx, filter, skip, limit, sortBy, order)
	if err != nil {
		return nil, 0, 0, err
	}

	responseData := make([]map[string]any, len(services))
	for i, svc := range services {
		responseData[i] = map[string]any{
			"id":                 svc.ID,
			"service_name":       svc.ServiceName,
			"category":           svc.Category,
			"vehicle_type":       svc.VehicleType,
			"location":           svc.Location,
			"requires_otp":       svc.RequiresOTP,
			"status":             svc.Status,
			"display_order":      svc.DisplayOrder,
			"tag":                svc.Tag,
			"min_cost":           svc.MinCost,
			"short_description":  svc.ShortDesc,
			"created_at":         svc.CreatedAt,
			"updated_at":         svc.UpdatedAt,
		}
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return responseData, total, totalPages, nil
}

func (s *ServiceService) GetServiceByID(ctx context.Context, id string) (map[string]any, error) {
	service, err := s.ServiceMasterRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                 service.ID.Hex(),
		"service_name":       service.ServiceName,
		"category":           service.Category,
		"vehicle_type":       service.VehicleType,
		"brand_scope":        service.BrandScope,
		"sub_brand_scope":    service.SubBrandScope,
		"fuel_type_scope":    service.FuelTypeScope,
		"location":           service.Location,
		"requires_otp":       service.RequiresOTP,
		"status":             service.Status,
		"display_order":      service.DisplayOrder,
		"tag":                service.Tag,
		"min_cost":           service.MinCost,
		"short_description":  service.ShortDesc,
		"created_at":         service.CreatedAt,
		"updated_at":         service.UpdatedAt,
	}, nil
}

func (s *ServiceService) UpdateService(
	ctx context.Context,
	id string,
	updateData map[string]any,
) error {

	if len(updateData) == 0 {
		return fmt.Errorf("no update data provided")
	}

	update := bson.M{}

	if val, ok := updateData["service_name"]; ok {
		name, ok := val.(string)
		if !ok || name == "" {
			return fmt.Errorf("invalid service name")
		}
		update["serviceName"] = name
	}

	if val, ok := updateData["category"]; ok {
		cat, ok := val.(string)
		if !ok || !domain.IsValidCategory(domain.ServiceCategory(cat)) {
			return fmt.Errorf("invalid category")
		}
		update["category"] = cat
	}

	if val, ok := updateData["vehicle_type"]; ok {
           vt, ok := val.(string)
			if !ok || !domain.IsValidVehicleType(domain.VehicleType(vt)) {
				return fmt.Errorf("invalid vehicle type: %v")
			}
		update["vehicleType"] = vt
	}

	if val, ok := updateData["brand_scope"]; ok {
		update["brandScope"] = val
	}

	if val, ok := updateData["sub_brand_scope"]; ok {
		update["subBrandScope"] = val
	}

	if val, ok := updateData["fuel_type_scope"]; ok {
		update["fuelTypeScope"] = val
	}

	if val, ok := updateData["location"]; ok {
		loc, ok := val.(string)
		if !ok || loc == "" {
			return fmt.Errorf("invalid location")
		}
		update["location"] = loc
	}

	if val, ok := updateData["requires_otp"]; ok {
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("invalid requires_otp")
		}
		update["requiresOtp"] = b
	}

	if val, ok := updateData["status"]; ok {
		status, ok := val.(string)
		if !ok || !domain.IsValidStatus(domain.ServiceMasterStatus(status)) {
			return fmt.Errorf("invalid status")
		}
		update["status"] = status
	}

	if val, ok := updateData["display_order"]; ok {
		update["displayOrder"] = val
	}

	if val, ok := updateData["tag"]; ok {
		update["tag"] = val
	}

	if val, ok := updateData["min_cost"]; ok {
		cost, ok := val.(float64)
		if !ok || cost < 0 {
			return fmt.Errorf("min cost cannot be negative")
		}
		update["minCost"] = cost
	}

	if val, ok := updateData["short_description"]; ok {
		update["shortDesc"] = val
	}

	return s.ServiceMasterRepo.Update(ctx, id, update)
}


func (s *ServiceService) UpdateServiceStatus(ctx context.Context, id string, status string) error {
	if status != "active" && status != "inactive" {
		return fmt.Errorf("invalid status: must be 'active' or 'inactive'")
	}

	return s.ServiceMasterRepo.UpdateStatus(ctx, id, status)
}

func (s *ServiceService) DeleteService(ctx context.Context, id string) error {
	return s.ServiceMasterRepo.Delete(ctx, id)
}

func (s *ServiceService) GetServiceStats(ctx context.Context) (map[string]any, error) {

	filter := bson.M{"status": "active"}
	activeServices, _, err := s.ServiceMasterRepo.GetServices(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"active_services":   len(activeServices),
	}, nil
}