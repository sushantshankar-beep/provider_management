package service

import (
	"context"
	"provider_management/internal/repository"
)

type RoleService struct {
	roles repository.RoleRepository
}

func NewRoleService(r *repository.RoleRepository) *RoleService {
	return &RoleService{
		roles: *r,
	}
}

type RoleTypeResponse struct {
	RoleType string `json:"roleType"`
}

type RoleNameResponse struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	RoleType string `json:"roleType"`
}

func (s *RoleService) GetRoleTypes(ctx context.Context) ([]RoleTypeResponse, error) {
	roleTypes, err := s.roles.FindDistinctRoleTypes(ctx)
	if err != nil {
		return nil, err
	}

	var result []RoleTypeResponse
	for _, rt := range roleTypes {
		result = append(result, RoleTypeResponse{RoleType: rt})
	}
	return result, nil
}

func (s *RoleService) GetRoleNamesByType(ctx context.Context, roleType string) ([]RoleNameResponse, error) {
	roles, err := s.roles.FindByRoleType(ctx, roleType)
	if err != nil {
		return nil, err
	}

	var result []RoleNameResponse
	for _, role := range roles {
		result = append(result, RoleNameResponse{
			ID:       role.ID.Hex(),
			Name:     role.Name,
			RoleType: role.RoleType,
		})
	}
	return result, nil
}