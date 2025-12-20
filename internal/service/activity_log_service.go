package service

import (
	"context"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
)

type ActivityLogService struct {
	repo *repository.ActivityLogRepository
}

func NewActivityLogService(repo *repository.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{repo: repo}
}

func (s *ActivityLogService) GetAll(ctx context.Context, page, limit int, adminID, entityType, action string) ([]domain.ActivityLogResponse, int64, error) {
	return s.repo.GetAll(ctx, page, limit, adminID, entityType, action)
}

func (s *ActivityLogService) GetByEntityID(ctx context.Context, entityType, entityID string, page, limit int) ([]domain.ActivityLogResponse, int64, error) {
	return s.repo.GetByEntityID(ctx, entityType, entityID, page, limit)
}
