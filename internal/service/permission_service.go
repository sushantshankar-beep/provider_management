package service

import (
	"time"
	"context"
	"provider_management/internal/dto"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
)

type PermissionService struct {
	permissionRepo *repository.PermissionRepo
}

func NewPermissionService(pr *repository.PermissionRepo) *PermissionService {
	return &PermissionService{permissionRepo: pr}
}

func (s *PermissionService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	permissions, err := s.permissionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.PermissionResponse, len(permissions))
	for i, p := range permissions {
		result[i] = s.mapToPermissionResponse(p)
	}

	return result, nil
}

func (s *PermissionService) GetPermission(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	p, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.mapToPermissionResponse(*p)
	return &resp, nil
}

func (s *PermissionService) CreatePermission(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	items := make([]domain.PermissionItem, len(req.Permissions))
	for i, p := range req.Permissions {
		items[i] = domain.PermissionItem{
			Key:   p.Key,
			Label: p.Label,
		}
	}

	permission := &domain.Permission{
		Category:    req.Category,
		Label:       req.Label,
		Permissions: items,
		Order:       req.Order,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	resp := s.mapToPermissionResponse(*permission)
	return &resp, nil
}

func (s *PermissionService) UpdatePermission(ctx context.Context, id string, req dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	existing, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Label != "" {
		existing.Label = req.Label
	}
	if req.Permissions != nil {
		items := make([]domain.PermissionItem, len(req.Permissions))
		for i, p := range req.Permissions {
			items[i] = domain.PermissionItem{
				Key:   p.Key,
				Label: p.Label,
			}
		}
		existing.Permissions = items
	}
	if req.Order > 0 {
		existing.Order = req.Order
	}
	existing.UpdatedAt = time.Now()

	if err := s.permissionRepo.Update(ctx, id, existing); err != nil {
		return nil, err
	}

	resp := s.mapToPermissionResponse(*existing)
	return &resp, nil
}

func (s *PermissionService) DeletePermission(ctx context.Context, id string) error {
	return s.permissionRepo.Delete(ctx, id)
}

func (s *PermissionService) mapToPermissionResponse(p domain.Permission) dto.PermissionResponse {
	items := make([]dto.PermissionItemDTO, len(p.Permissions))
	for i, item := range p.Permissions {
		items[i] = dto.PermissionItemDTO{
			Key:   item.Key,
			Label: item.Label,
		}
	}

	return dto.PermissionResponse{
		ID:          p.ID.Hex(),
		Category:    p.Category,
		Label:       p.Label,
		Permissions: items,
		Order:       p.Order,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}