package service

import (
	"context"
	"errors"
	"time"

	"provider_management/internal/domain"
	"provider_management/internal/repository"
)

var ErrComplaintNotFound = errors.New("complaint not found")

type ComplaintService struct {
	complaints repository.ComplaintRepo
	assess     repository.AssessmentRepo
}

func NewComplaintService(c *repository.ComplaintRepo, a *repository.AssessmentRepo) *ComplaintService {
	return &ComplaintService{complaints: *c, assess: *a}
}

func (s *ComplaintService) ListComplaints(ctx context.Context) ([]domain.Complaint, error) {
	return s.complaints.FindAll(ctx, 0, 100)
}

func (s *ComplaintService) GetComplaint(ctx context.Context, id string) (*domain.Complaint, error) {
	return s.complaints.FindByID(ctx, id)
}

func (s *ComplaintService) PostAssessment(ctx context.Context, id string, a *domain.Assessment) error {
	_, err := s.complaints.FindByID(ctx, id)
	if err != nil {
		return ErrComplaintNotFound
	}

	a.CreatedAt = time.Now().Unix()
	a.ComplaintID = id
	return s.assess.Save(ctx, a)
}
