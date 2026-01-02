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
	TxnID             string `json:"-"`
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

	if primitive.IsValidObjectID(cleanID) {
		complaint, err := s.complaintRepo.GetByID(ctx, cleanID)
		if err == nil {
			log.Printf("Found complaint by MongoDB ObjectID: %s", cleanID)
			return complaint, nil
		}
	}

	if internalID, parseErr := strconv.ParseInt(cleanID, 10, 64); parseErr == nil {
		
		complaint, err := s.complaintRepo.GetByInternalID(ctx, internalID)
		if err == nil {
			log.Printf("Found complaint by internal ID: %d, MongoDB _id: %s", internalID, complaint.ID)
			return complaint, nil
		}
		
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

	var booking *domain.AcceptedService
	if complaint.AcceptedServiceID != "" {
		booking, err = s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedServiceID)
		if err != nil {
			log.Printf("Warning: Could not fetch booking details for ID %s: %v", complaint.AcceptedServiceID, err)
		} else {
			complaintWithDetails.BookingDetails = &domain.BookingDetails{
				ID:         booking.ID.Hex(),
				InternalID: booking.InternalID,
				BasePrice:  booking.BasePrice,
				FinalPrice: booking.FinalPrice,
			}
		}
	}

	userID := complaint.UserID
	if userID == "" && booking != nil {
		userID = booking.UserID
	}

	if userID != "" {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			log.Printf("Warning: Could not fetch user details for ID %s: %v", userID, err)
		} else {
			complaintWithDetails.UserDetails = &domain.UserDetails{
				ID:         user.ID,
				InternalID: fmt.Sprintf("VW%d", user.InternalID),
				Name:       user.Name,
				Email:      user.Email,
				Phone:      user.Phone,
			}
		}
	}

	providerID := complaint.ProviderID
	if providerID == "" && booking != nil {
		providerID = booking.ProviderID.Hex()
	}

	if providerID != "" {
		provider, err := s.providerRepo.FindByID(ctx, providerID)
		if err != nil {
			log.Printf("Warning: Could not fetch provider details for ID %s: %v", providerID, err)
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

func (s *ComplaintService) ListComplaints(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, *domain.ComplaintStats, error) {
	return s.complaintRepo.List(ctx, filter)
}
func (s *ComplaintService) AssessComplaint(ctx context.Context, complaintID string, req AssessComplaintRequest) error {
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != "in_review" {
		return fmt.Errorf("complaint must be in in_review status to save assessment, current status: %s", complaint.Status)
	}

	acceptedService, err := s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedServiceID)
	if err != nil {
		log.Printf("ERROR: Failed to get accepted service: %v", err)
		return fmt.Errorf("failed to get accepted service: %w", err)
	}

	originalAmount := acceptedService.BasePrice
	
	if originalAmount <= 0 {
		log.Printf("ERROR: Original amount is invalid: %.2f", originalAmount)
		return fmt.Errorf("invalid original booking amount: %.2f", originalAmount)
	}

	if req.RefundToUser == domain.RefundTypeFull {
		req.RefundAmount = originalAmount
		log.Printf("AssessComplaint - Full refund selected, setting refund amount to: %.2f", req.RefundAmount)
	}

	if req.PayoutToProvider == domain.PayoutTypeFull {
		req.PayoutAmount = originalAmount
		log.Printf("AssessComplaint - Full payout selected, setting payout amount to: %.2f", req.PayoutAmount)
	}

	if req.RefundToUser == domain.RefundTypePartial {
		if req.RefundAmount <= 0 {
			return fmt.Errorf("refund_amount must be greater than 0 when refund_to_user is 'partial'")
		}
		if req.RefundAmount > originalAmount {
			return fmt.Errorf("refund_amount cannot exceed original booking amount: %.2f", originalAmount)
		}
		log.Println("edciswhuiwhiu")
	}

	if req.PayoutToProvider == domain.PayoutTypePartial {
		if req.PayoutAmount <= 0 {
			return fmt.Errorf("payout_amount must be greater than 0 when payout_to_provider is 'partial'")
		}
		if req.PayoutAmount > originalAmount {
			return fmt.Errorf("payout_amount cannot exceed original booking amount: %.2f", originalAmount)
		}
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

	if err := s.complaintRepo.SaveAssessment(ctx, complaint.ID, assessment); err != nil {
		return fmt.Errorf("failed to save assessment: %w", err)
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, "resolved"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	paymentTracking := &domain.PaymentActionTracking{
		RefundStatus: domain.PaymentActionNA,
		PayoutStatus: domain.PaymentActionNA,
	}

	actions := []string{}
	var userObjID primitive.ObjectID
    if complaint.UserID != "" {
        userObjID, err = primitive.ObjectIDFromHex(complaint.UserID)
        if err != nil {
            log.Printf("Warning: Invalid UserID format: %s", complaint.UserID)
        }
    }
	if req.RefundToUser != domain.RefundTypeNone && req.RefundAmount > 0 {
		paymentTracking.RefundStatus = domain.PaymentActionPending
		refundReason := req.Remarks
		
		if req.RefundToUser == domain.RefundTypeFull {
			refundReason = "Full Refund - " + refundReason
		} else {
			refundReason = fmt.Sprintf("Partial Refund -" + refundReason)
		}
		
		if err := s.refundService.ProcessRefund(ctx, RefundRequest{
			UserID:              userObjID.Hex(),
			TxnID:               req.TxnID,
			BookingID:           complaint.AcceptedServiceID,
			BookingInternalID:   complaint.AcceptedServiceNo,
			ComplaintID:         complaint.ID,
			ComplaintInternalID: complaint.InternalID,
			Amount:              req.RefundAmount,
			Reason:              refundReason,
		}); err != nil {
			log.Printf("Warning: Failed to process refund: %v", err)
		} else {
			actionMsg := fmt.Sprintf("Refund of %.2f sent to Refund Management module", req.RefundAmount)
			if req.RefundToUser == domain.RefundTypeFull {
				actionMsg = "Full " + actionMsg
			} else {
				actionMsg = "Partial " + actionMsg
			}
			actions = append(actions, actionMsg)
		}
	}

	if req.PayoutToProvider != domain.PayoutTypeNone && req.PayoutAmount > 0 {
		paymentTracking.PayoutStatus = domain.PaymentActionPending
		acceptedService, err := s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedServiceID)
		if err != nil {
			log.Printf("Warning: Cannot fetch accepted service for payout: %v", err)
		} else if acceptedService.ProviderID.IsZero() {
			log.Printf("Warning: Cannot process payout - provider ID is empty in accepted service")
		} else {
			providerID := acceptedService.ProviderID.Hex()
			log.Printf("AssessComplaint - Processing payout of %.2f for provider %s", req.PayoutAmount, providerID)
			
			payoutReason := fmt.Sprintf("Complaint CMP%d - %s", complaint.InternalID, req.Remarks)
			if req.PayoutToProvider == domain.PayoutTypeFull {
				payoutReason = "Full Payout - " + payoutReason
			} else {
				payoutReason = fmt.Sprintf("Partial Payout" + payoutReason)
			}
			
			if err := s.payoutService.ProcessPayout(ctx, PayoutRequest{
				ProviderID:    providerID,
				BookingID:     complaint.AcceptedServiceID,
				Amount:        acceptedService.BasePrice,
				PartialAmount: req.PayoutAmount,
				Reason:        req.Remarks,
				ComplaintID:   complaint.ID,
			}); err != nil {
				log.Printf("Warning: Failed to process payout: %v", err)
			} else {
				actionMsg := fmt.Sprintf("Payout of %.2f sent to Settlement Management module", req.PayoutAmount)
				if req.PayoutToProvider == domain.PayoutTypeFull {
					actionMsg = "Full " + actionMsg
				} else {
					actionMsg = "Partial " + actionMsg
				}
				actions = append(actions, actionMsg)
			}
		}
	}

	if len(actions) == 0 {
		actions = append(actions, "No financial actions triggered")
	}

	log.Printf("AssessComplaint - Updating triggered actions: %v", actions)
	if err := s.complaintRepo.Update(ctx, complaint.ID, bson.M{
		"actionsTriggered": actions,
		"paymentTracking":  paymentTracking,
	}); err != nil {
		log.Printf("Warning: Failed to update actions: %v", err)
	}

	log.Printf("AssessComplaint - Assessment completed successfully")
	return nil
}

func (s *ComplaintService) StartAssessment(ctx context.Context, complaintID string) error {
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != "initiated" {
		return fmt.Errorf("complaint must be in initiated status to start assessment, current status: %s", complaint.Status)
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, "in_review"); err != nil {
		return fmt.Errorf("failed to update status to in_review: %w", err)
	}

	log.Printf("StartAssessment - Status changed to in_review for complaint %s", complaintID)
	return nil
}

func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID string, status string, adminID string) error {

	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}



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


	if err := s.complaintRepo.Update(ctx, complaint.ID, updateData); err != nil {
		
		return fmt.Errorf("failed to update admin info: %w", err)
	}


	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, status); err != nil {
		
		return fmt.Errorf("failed to update status: %w", err)
	}


	return nil
}

func (s *ComplaintService) AddNote(ctx context.Context, internalID int64, req AddNoteRequest) error {
	
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
	
	err := s.complaintRepo.AddNote(ctx, internalID, note)
	if err != nil {
		return fmt.Errorf("failed to add note to complaint: %w", err)
	}

	return nil
}

func (s *ComplaintService) GetComplaintStats(ctx context.Context) (*domain.ComplaintStats, error) {
	return s.complaintRepo.GetStats(ctx)
}

type RefundRequest struct {
	UserID string
	BookingID         string
	BookingInternalID int64
	ComplaintID         string
	ComplaintInternalID int64
	Amount float64
	Reason string
	TxnID  string     
}

type PayoutRequest struct {
	ProviderID  string
	Amount      float64
	Reason      string
	ComplaintID string
	BookingID     string
	PartialAmount float64
}

func (s *ComplaintService) UpdatePaymentStatus(ctx context.Context, complaintID string, isRefund bool, status domain.PaymentActionStatus, paymentID string) error {
	complaint, err := s.GetComplaint(ctx, complaintID)
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	tracking := complaint.PaymentTracking
	if tracking == nil {
		tracking = &domain.PaymentActionTracking{
			RefundStatus: domain.PaymentActionNA,
			PayoutStatus: domain.PaymentActionNA,
		}
	}

	if isRefund {
		tracking.RefundStatus = status
		tracking.RefundID = paymentID
	} else {
		tracking.PayoutStatus = status
		tracking.PayoutID = paymentID
	}

	return s.complaintRepo.Update(ctx, complaint.ID, bson.M{
		"paymentTracking": tracking,
	})
}