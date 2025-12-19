package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"
)

type ComplaintService struct {
	complaintRepo       *repository.ComplaintRepository
	acceptedServiceRepo *repository.AcceptedServiceRepo
	userRepo            *repository.UserRepo
	providerRepo        *repository.ProviderRepo
	refundService       *RefundService
	payoutService       *PayoutService
}

type CreateComplaintRequest struct {
	AcceptedServiceID string   `json:"accepted_service_id"`
	RaisedBy          string   `json:"raised_by"`
	Problem           string   `json:"problem"`
	Photos            []string `json:"photos,omitempty"`
	Category          string   `json:"category,omitempty"`
}

type AssessComplaintRequest struct {
	FaultParty       domain.FaultParty `json:"fault_party" binding:"required"`
	RefundToUser     domain.RefundType `json:"refund_to_user" binding:"required"`
	RefundAmount     float64           `json:"refund_amount,omitempty"`
	PayoutToProvider domain.PayoutType `json:"payout_to_provider" binding:"required"`
	PayoutAmount     float64           `json:"payout_amount,omitempty"`
	Remarks          string            `json:"remarks,omitempty"`
	AssessedBy       string            `json:"assessed_by"`
}

type AddNoteRequest struct {
	Content string `json:"content" binding:"required"`
	AddedBy string `json:"addedBy" binding:"required"`
}

func NewComplaintService(
	complaintRepo *repository.ComplaintRepository,
	acceptedServiceRepo *repository.AcceptedServiceRepo,
	userRepo *repository.UserRepo,
	providerRepo *repository.ProviderRepo,
	refundService *RefundService,
	payoutService *PayoutService,
) *ComplaintService {
	return &ComplaintService{
		complaintRepo:       complaintRepo,
		acceptedServiceRepo: acceptedServiceRepo,
		userRepo:            userRepo,
		providerRepo:        providerRepo,
		refundService:       refundService,
		payoutService:       payoutService,
	}
}

func (s *ComplaintService) GetComplaint(ctx context.Context, id string) (*domain.Complaint, error) {

	cleanID := strings.TrimPrefix(id, "CMP")

	log.Printf("GetComplaint - Original ID: %s, Clean ID: %s", id, cleanID)

	if primitive.IsValidObjectID(cleanID) {
		complaint, err := s.complaintRepo.GetByID(ctx, cleanID)
		if err == nil {
			log.Printf("Found complaint by MongoDB ObjectID: %s", cleanID)
			return complaint, nil
		}
		log.Printf("Not found by ObjectID, trying internal ID. Error: %v", err)
	}

	if internalID, parseErr := strconv.ParseInt(cleanID, 10, 64); parseErr == nil {
		log.Printf("Trying to find by internal ID: %d", internalID)
		complaint, err := s.complaintRepo.GetByInternalID(ctx, internalID)
		if err == nil {
			log.Printf("Found complaint by internal ID: %d, MongoDB _id: %s", internalID, complaint.ID)
			return complaint, nil
		}
		log.Printf("Not found by internal ID. Error: %v", err)
	}

	return nil, fmt.Errorf("complaint not found with ID: %s", id)
}

func (s *ComplaintService) GetComplaintWithDetails(ctx context.Context, id string) (*domain.ComplaintWithDetails, error) {
	complaint, err := s.GetComplaint(ctx, id)
	if err != nil {
		return nil, err
	}

	complaintWithDetails := &domain.ComplaintWithDetails{
		Complaint: *complaint,
	}

	if complaint.UserID != "" {
		user, err := s.userRepo.FindByID(ctx, complaint.UserID)
		if err != nil {
			log.Printf("Warning: Could not fetch user details for ID %s: %v", complaint.UserID, err)
		} else {
			complaintWithDetails.UserDetails = &domain.UserDetails{
				ID:    user.ID,
				InternalID: fmt.Sprintf("VW%d", user.InternalID),
				Name:  user.Name,
				Email: user.Email,
				Phone: user.Phone,
			}
		}
	}

	if complaint.ProviderID != "" {
		provider, err := s.providerRepo.FindByID(ctx, complaint.ProviderID)
		if err != nil {
			log.Printf("Warning: Could not fetch provider details for ID %s: %v", complaint.ProviderID, err)
		} else {
			complaintWithDetails.ProviderDetails = &domain.ProviderDetails{
				ID:          provider.ID.Hex(),
				InternalID:  fmt.Sprintf("PRO%d", provider.InternalID),
				Name:        provider.Name,
				Email:       provider.Email,
				Phone:       provider.Phone,
				CompanyName: provider.CompanyName,
			}
		}
	}

	return complaintWithDetails, nil
}

func (s *ComplaintService) ListComplaints(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error) {
	return s.complaintRepo.List(ctx, filter)
}

func (s *ComplaintService) AssessComplaint(ctx context.Context, complaintID string, req AssessComplaintRequest) error {

	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != "initiated" {
		return fmt.Errorf("complaint must be in initiated status to assess, current status: %s", complaint.Status)
	}

	if req.RefundToUser != domain.RefundTypeNone && req.RefundAmount <= 0 {
		return fmt.Errorf("refund_amount must be greater than 0 when refund_to_user is not 'none'")
	}

	if req.PayoutToProvider != domain.PayoutTypeNone && req.PayoutAmount <= 0 {
		return fmt.Errorf("payout_amount must be greater than 0 when payout_to_provider is not 'none'")
	}

	assessment := domain.ComplaintAssessment{
		FaultParty:       req.FaultParty,
		RefundToUser:     req.RefundToUser,
		RefundAmount:     req.RefundAmount,
		PayoutToProvider: req.PayoutToProvider,
		PayoutAmount:     req.PayoutAmount,
		Remarks:          req.Remarks,
		AssessedBy:       req.AssessedBy,
	}

	log.Printf("AssessComplaint - Saving assessment using MongoDB _id: %s", complaint.ID)
	if err := s.complaintRepo.SaveAssessment(ctx, complaint.ID, assessment); err != nil {
		log.Printf("AssessComplaint - Failed to save assessment: %v", err)
		return fmt.Errorf("failed to save assessment: %w", err)
	}

	log.Printf("AssessComplaint - Updating status to resolved")
	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, "resolved"); err != nil {
		log.Printf("AssessComplaint - Failed to update status: %v", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	if req.RefundToUser != domain.RefundTypeNone && req.RefundAmount > 0 {
		log.Printf("AssessComplaint - Processing refund of %.2f for user %s", req.RefundAmount, complaint.UserID)
		if err := s.refundService.ProcessRefund(ctx, RefundRequest{
			UserID:      complaint.UserID,
			Amount:      req.RefundAmount,
			Reason:      fmt.Sprintf("Complaint ID %d - %s", complaint.InternalID, req.Remarks),
			ComplaintID: complaint.ID,
		}); err != nil {
			log.Printf("Warning: Failed to process refund: %v", err)
		}
	}

	// if req.PayoutToProvider != domain.PayoutTypeNone && req.PayoutAmount > 0 {
	// 	log.Printf("AssessComplaint - Processing payout of %.2f for provider %s", req.PayoutAmount, complaint.ProviderID)
	// 	if err := s.payoutService.ProcessPayout(ctx, PayoutRequest{
	// 		ProviderID:  complaint.ProviderID,
	// 		Amount:      req.PayoutAmount,
	// 		Reason:      fmt.Sprintf("Complaint ID %d - %s", complaint.InternalID, req.Remarks),
	// 		ComplaintID: complaint.ID,
	// 	}); err != nil {
	// 		log.Printf("Warning: Failed to process payout: %v", err)
	// 	}
	// }

	actions := []string{}
	if req.RefundToUser != domain.RefundTypeNone {
		actions = append(actions, "Refund sent to Refund Management module")
	}
	if req.PayoutToProvider != domain.PayoutTypeNone {
		actions = append(actions, "Payout sent to Settlement Management module")
	}

	if len(actions) > 0 {
		log.Printf("AssessComplaint - Updating triggered actions: %v", actions)
		if err := s.complaintRepo.Update(ctx, complaint.ID, bson.M{
			"actionsTriggered": actions,
		}); err != nil {
			log.Printf("Warning: Failed to update actions: %v", err)
		}
	}

	log.Printf("AssessComplaint - Assessment completed successfully")
	return nil
}

func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID string, status string, adminID string) error {
	log.Printf("UpdateComplaintStatus - Starting for complaint ID: %s, new status: %s", complaintID, status)

	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		log.Printf("UpdateComplaintStatus - Failed to get complaint: %v", err)
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	log.Printf("UpdateComplaintStatus - Found complaint with MongoDB _id: %s", complaint.ID)

	validStatuses := []string{"initiated", "in_review", "resolved", "cancelled"}
	isValid := false
	for _, vs := range validStatuses {
		if status == vs {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid status: %s", status)
	}

	now := time.Now()
	updateData := bson.M{
		"updatedByAdmin": adminID,
		"adminUpdatedAt": now,
	}

	log.Printf("UpdateComplaintStatus - Updating admin info using MongoDB _id: %s", complaint.ID)
	if err := s.complaintRepo.Update(ctx, complaint.ID, updateData); err != nil {
		log.Printf("UpdateComplaintStatus - Failed to update admin info: %v", err)
		return fmt.Errorf("failed to update admin info: %w", err)
	}

	log.Printf("UpdateComplaintStatus - Updating status")
	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, status); err != nil {
		log.Printf("UpdateComplaintStatus - Failed to update status: %v", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	log.Printf("UpdateComplaintStatus - Status updated successfully")
	return nil
}

func (s *ComplaintService) AddNote(ctx context.Context, complaintID string, req AddNoteRequest) error {

	if _, err := primitive.ObjectIDFromHex(complaintID); err != nil {
		return fmt.Errorf("invalid complaint ID format: %w", err)
	}

	if req.Content == "" {
		return fmt.Errorf("note content cannot be empty")
	}

	if req.AddedBy == "" {
		return fmt.Errorf("addedBy field is required")
	}

	note := domain.ComplaintNote{
		ID:        primitive.NewObjectID().Hex(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	err := s.complaintRepo.AddNote(ctx, complaintID, note)
	if err != nil {
		log.Printf("Repository error: %v", err)
		return fmt.Errorf("failed to add note to complaint: %w", err)
	}

	return nil
}

func (s *ComplaintService) GetComplaintStats(ctx context.Context) (*domain.ComplaintStats, error) {
	return s.complaintRepo.GetStats(ctx)
}

type RefundRequest struct {
	UserID      string
	Amount      float64
	Reason      string
	ComplaintID string
}

type PayoutRequest struct {
	ProviderID  string
	Amount      float64
	Reason      string
	ComplaintID string
}
