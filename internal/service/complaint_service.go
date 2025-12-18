package service

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"time"
)

type ComplaintService interface {
	CreateComplaint(ctx context.Context, req CreateComplaintRequest) (*domain.Complaint, error)
	GetComplaint(ctx context.Context, id string) (*domain.Complaint, error)
	ListComplaints(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error)
	AssessComplaint(ctx context.Context, complaintID string, req AssessComplaintRequest) error
	UpdateComplaintStatus(ctx context.Context, complaintID string, status string, adminID string) error
	AddNote(ctx context.Context, complaintID string, req AddNoteRequest) error
	GetComplaintStats(ctx context.Context) (*domain.ComplaintStats, error)
}

type complaintService struct {
	complaintRepo       repository.ComplaintRepository
	acceptedServiceRepo repository.AcceptedServiceRepo
	userRepo            repository.UserRepo
	providerRepo        repository.ProviderRepo
	refundService       RefundService
	payoutService       PayoutServices
}

type CreateComplaintRequest struct {
	AcceptedServiceID string   `json:"accepted_service_id"`
	RaisedBy          string   `json:"raised_by"` // "user" or "provider"
	Problem           string   `json:"problem"`
	Photos            []string `json:"photos,omitempty"`
	Category          string   `json:"category,omitempty"`
}

type AssessComplaintRequest struct {
	FaultParty       domain.FaultParty `json:"fault_party"`
	RefundToUser     domain.RefundType `json:"refund_to_user"`
	RefundAmount     float64           `json:"refund_amount,omitempty"`
	PayoutToProvider domain.PayoutType `json:"payout_to_provider"`
	PayoutAmount     float64           `json:"payout_amount,omitempty"`
	Remarks          string            `json:"remarks,omitempty"`
	AssessedBy       string            `json:"assessed_by"`
}

type AddNoteRequest struct {
	Content string `json:"content"`
	AddedBy string `json:"added_by"`
}

func NewComplaintService(
	complaintRepo repository.ComplaintRepository,
	acceptedServiceRepo repository.AcceptedServiceRepo,
	userRepo repository.UserRepo,
	providerRepo repository.ProviderRepo,
	refundService RefundService,
	payoutService PayoutServices,
) ComplaintService {
	return &complaintService{
		complaintRepo:       complaintRepo,
		acceptedServiceRepo: acceptedServiceRepo,
		userRepo:            userRepo,
		providerRepo:        providerRepo,
		refundService:       refundService,
		payoutService:       payoutService,
	}
}

func (s *complaintService) CreateComplaint(ctx context.Context, req CreateComplaintRequest) (*domain.Complaint, error) {
	// Get accepted service details
	// acceptedService, err := s.acceptedServiceRepo.GetByID(ctx, req.AcceptedServiceID)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get accepted service: %w", err)
	// }

	// complaint := &domain.Complaint{
	// 	AcceptedServiceID: req.AcceptedServiceID,
	// 	AcceptedServiceNo: acceptedService.InternalID,
	// 	RaisedBy:          req.RaisedBy,
	// 	Problem:           req.Problem,
	// 	Photos:            req.Photos,
	// 	Status:            "pending",
	// 	Category:          req.Category,
	// 	Timeline: domain.ComplaintTimeline{
	// 		Initiated: timePtr(time.Now()),
	// 	},
	// }

	// Set UserID and ProviderID from accepted service
	// complaint.UserID = acceptedService.UserID
	// complaint.ProviderID = acceptedService.ProviderID.Hex()

	// // Fetch and set user name
	// user, err := s.userRepo.GetByID(ctx, acceptedService.UserID)
	// if err == nil && user != nil {
	// 	complaint.UserName = user.Name
	// }

	// // Fetch and set provider name
	// provider, err := s.providerRepo.GetByID(ctx, acceptedService.ProviderID.Hex())
	// if err == nil && provider != nil {
	// 	complaint.ProviderName = provider.Name
	// }

	// // Set booking number (using AcceptedServiceNo)
	// complaint.BookingNumber = fmt.Sprintf("BKG-%d", acceptedService.InternalID)

	// // Create complaint
	// if err := s.complaintRepo.Create(ctx, complaint); err != nil {
	// 	return nil, fmt.Errorf("failed to create complaint: %w", err)
	// }

	// return complaint, nil
}

func (s *complaintService) GetComplaint(ctx context.Context, id string) (*domain.Complaint, error) {
	// Try to get by ID string
	complaint, err := s.complaintRepo.GetByID(ctx, id)
	if err == nil {
		return complaint, nil
	}

	// Try to parse as internal ID (integer)
	if internalID, parseErr := strconv.ParseInt(id, 10, 64); parseErr == nil {
		return s.complaintRepo.GetByInternalID(ctx, internalID)
	}

	return nil, fmt.Errorf("complaint not found")
}

func (s *complaintService) ListComplaints(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error) {
	return s.complaintRepo.List(ctx, filter)
}

func (s *complaintService) AssessComplaint(ctx context.Context, complaintID string, req AssessComplaintRequest) error {
	// Get complaint
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != "in_review" {
		return fmt.Errorf("complaint must be in review status to assess")
	}

	// Create assessment
	assessment := domain.ComplaintAssessment{
		FaultParty:       req.FaultParty,
		RefundToUser:     req.RefundToUser,
		RefundAmount:     req.RefundAmount,
		PayoutToProvider: req.PayoutToProvider,
		PayoutAmount:     req.PayoutAmount,
		Remarks:          req.Remarks,
		AssessedBy:       req.AssessedBy,
	}

	// Save assessment
	if err := s.complaintRepo.SaveAssessment(ctx, complaint.ID, assessment); err != nil {
		return fmt.Errorf("failed to save assessment: %w", err)
	}

	// Update status to resolved
	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, "resolved"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Process refund if applicable
	if req.RefundToUser != domain.RefundTypeNone && req.RefundAmount > 0 {
		if err := s.refundService.ProcessRefund(ctx, RefundRequest{
			UserID:      complaint.UserID,
			Amount:      req.RefundAmount,
			Reason:      fmt.Sprintf("Complaint ID %d - %s", complaint.InternalID, req.Remarks),
			ComplaintID: complaint.ID,
		}); err != nil {
			// Log error but don't fail the assessment
			fmt.Printf("failed to process refund: %v\n", err)
		}
	}

	// Process payout if applicable
	if req.PayoutToProvider != domain.PayoutTypeNone && req.PayoutAmount > 0 {
		if err := s.payoutService.ProcessPayout(ctx, PayoutRequest{
			ProviderID:  complaint.ProviderID,
			Amount:      req.PayoutAmount,
			Reason:      fmt.Sprintf("Complaint ID %d - %s", complaint.InternalID, req.Remarks),
			ComplaintID: complaint.ID,
		}); err != nil {
			// Log error but don't fail the assessment
			fmt.Printf("failed to process payout: %v\n", err)
		}
	}

	// Add actions triggered
	actions := []string{}
	if req.RefundToUser != domain.RefundTypeNone {
		actions = append(actions, "Refund sent to Refund Management module")
	}
	if req.PayoutToProvider != domain.PayoutTypeNone {
		actions = append(actions, "Payout sent to Settlement Management module")
	}

	if len(actions) > 0 {
		if err := s.complaintRepo.Update(ctx, complaint.ID, map[string]interface{}{
			"actionsTriggered": actions,
		}); err != nil {
			fmt.Printf("failed to update actions: %v\n", err)
		}
	}

	return nil
}

func (s *complaintService) UpdateComplaintStatus(ctx context.Context, complaintID string, status string, adminID string) error {
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	// Update admin info
	now := time.Now()
	updateData := map[string]interface{}{
		"updatedByAdmin": adminID,
		"adminUpdatedAt": now,
	}

	if err := s.complaintRepo.Update(ctx, complaint.ID, updateData); err != nil {
		return fmt.Errorf("failed to update admin info: %w", err)
	}

	return s.complaintRepo.UpdateStatus(ctx, complaint.ID, status)
}

func (s *complaintService) AddNote(ctx context.Context, complaintID string, req AddNoteRequest) error {
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	note := domain.ComplaintNote{
		Content: req.Content,
		AddedBy: req.AddedBy,
	}

	return s.complaintRepo.AddNote(ctx, complaint.ID, note)
}

func (s *complaintService) GetComplaintStats(ctx context.Context) (*domain.ComplaintStats, error) {
	return s.complaintRepo.GetStats(ctx)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Placeholder interfaces - you'll need to implement these based on your refund/payout logic
type RefundService interface {
	ProcessRefund(ctx context.Context, req RefundRequest) error
}

type RefundRequest struct {
	UserID      string
	Amount      float64
	Reason      string
	ComplaintID string
}

type PayoutServices interface {
	ProcessPayout(ctx context.Context, req PayoutRequest) error
}

type PayoutRequest struct {
	ProviderID  string
	Amount      float64
	Reason      string
	ComplaintID string
}
