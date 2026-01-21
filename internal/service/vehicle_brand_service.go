package service

import (
	"context"
	"errors"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
)

type VehicleBrandService struct {
	brands repository.VehicleBrandRepo
}

func NewVehicleBrandService(b *repository.VehicleBrandRepo) *VehicleBrandService {
	return &VehicleBrandService{
		brands: *b,
	}
}

func (s *VehicleBrandService) GetVehicleBrands(ctx context.Context, vehicleType, brandName string) ([]dto.VehicleBrandResponse, error) {
	brands, err := s.brands.FindWithFilter(ctx, vehicleType, brandName)
	if err != nil {
		return nil, err
	}

	result := make([]dto.VehicleBrandResponse, len(brands))
	for i, b := range brands {
		result[i] = dto.VehicleBrandResponse{
			BrandName:   b.BrandName,
			VehicleType: b.VehicleType,
			ModelName:   b.ModelName,
		}
	}

	return result, nil
}

func (s *VehicleBrandService) GetVehicleModels(ctx context.Context, vehicleType, brandName string) ([]string, error) {
	brand, err := s.brands.FindByTypeAndName(ctx, vehicleType, brandName)
	if err != nil {
		return nil, errors.New("brand not found")
	}

	return brand.ModelName, nil
}