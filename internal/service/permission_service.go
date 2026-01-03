package service

import (
	"context"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"time"
)

type PermissionService struct {
	permissionRepo *repository.PermissionRepo
}

func NewPermissionService(pr *repository.PermissionRepo) *PermissionService {
	return &PermissionService{
		permissionRepo: pr,
	}
}

type PermissionResponse struct {
	ID          string                   `json:"_id"`
	Category    string                   `json:"category"`
	Label       string                   `json:"label"`
	Permissions []PermissionItemResponse `json:"permissions"`
	Order       int                      `json:"order"`
	CreatedAt   string                   `json:"createdAt"`
	UpdatedAt   string                   `json:"updatedAt"`
}

type PermissionItemResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}


type UpdatePermissionRequest struct {
	Category    string                   `json:"category"`
	Label       string                   `json:"label"`
	Permissions []PermissionItemResponse `json:"permissions"`
	Order       int                      `json:"order"`
}

type CreatePermissionRequest struct {
	Category    string                   `json:"category" binding:"required"`
	Label       string                   `json:"label" binding:"required"`
	Permissions []PermissionItemResponse `json:"permissions" binding:"required"`
	Order       int                      `json:"order"`
}


func (s *PermissionService) ListPermissions(ctx context.Context) ([]PermissionResponse, error) {
	permissions, err := s.permissionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []PermissionResponse
	for _, p := range permissions {
		items := make([]PermissionItemResponse, len(p.Permissions))
		for i, item := range p.Permissions {
			items[i] = PermissionItemResponse{
				Key:   item.Key,
				Label: item.Label,
			}
		}

		result = append(result, PermissionResponse{
			ID:          p.ID.Hex(),
			Category:    p.Category,
			Label:       p.Label,
			Permissions: items,
			Order:       p.Order,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *PermissionService) GetPermission(ctx context.Context, id string) (*PermissionResponse, error) {
	p, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items := make([]PermissionItemResponse, len(p.Permissions))
	for i, item := range p.Permissions {
		items[i] = PermissionItemResponse{
			Key:   item.Key,
			Label: item.Label,
		}
	}

	return &PermissionResponse{
		ID:          p.ID.Hex(),
		Category:    p.Category,
		Label:       p.Label,
		Permissions: items,
		Order:       p.Order,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PermissionService) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error) {
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

	respItems := make([]PermissionItemResponse, len(items))
	for i, item := range items {
		respItems[i] = PermissionItemResponse{
			Key:   item.Key,
			Label: item.Label,
		}
	}

	return &PermissionResponse{
		ID:          permission.ID.Hex(),
		Category:    permission.Category,
		Label:       permission.Label,
		Permissions: respItems,
		Order:       permission.Order,
		CreatedAt:   permission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   permission.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PermissionService) UpdatePermission(ctx context.Context, id string, req UpdatePermissionRequest) (*PermissionResponse, error) {
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

	respItems := make([]PermissionItemResponse, len(existing.Permissions))
	for i, item := range existing.Permissions {
		respItems[i] = PermissionItemResponse{
			Key:   item.Key,
			Label: item.Label,
		}
	}

	return &PermissionResponse{
		ID:          existing.ID.Hex(),
		Category:    existing.Category,
		Label:       existing.Label,
		Permissions: respItems,
		Order:       existing.Order,
		CreatedAt:   existing.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   existing.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PermissionService) DeletePermission(ctx context.Context, id string) error {
	return s.permissionRepo.Delete(ctx, id)
}