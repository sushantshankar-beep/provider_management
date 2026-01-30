package service

import (
	"context"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
)

type ProviderBrandService struct {
	brandRepo   *repository.ProviderVehicleBrandRepo
	serviceRepo *repository.ServiceMasterRepo
}

func NewProviderBrandService(
	brandRepo *repository.ProviderVehicleBrandRepo,
	serviceRepo *repository.ServiceMasterRepo,
) *ProviderBrandService {
	return &ProviderBrandService{
		brandRepo:   brandRepo,
		serviceRepo: serviceRepo,
	}
}


func (s *ProviderBrandService) GetBrands( ctx context.Context, vehicle string) ([]domain.ProviderVehicleBrand, error) {

	brands, err := s.brandRepo.FindByVehicle(ctx, vehicle)
	if err != nil {
		return nil, err
	}
	return brands, nil
}


func (s *ProviderBrandService) GetServices( ctx context.Context, vehicle string ) ([]domain.ProviderService, error) {

	services, err := s.serviceRepo.FindByVehicle(ctx, vehicle)
	if err != nil {
		return nil, err
	}

	return services, nil
}
