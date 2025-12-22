package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefundService struct {
	refundRepo *repository.RefundRepository
}

func NewRefundService(refundRepo *repository.RefundRepository) *RefundService {
	return &RefundService{
		refundRepo: refundRepo,
	}
}

func (s *RefundService) ProcessRefund(ctx context.Context, req RefundRequest) error {
	log.Printf("ProcessRefund - Starting refund for user %s, amount: %.2f", req.UserID, req.Amount)

	var complaintID *primitive.ObjectID
	var complaintNo *int64
	if req.ComplaintID != "" {
		objID, err := primitive.ObjectIDFromHex(req.ComplaintID)
		if err != nil {
			log.Printf("Warning: Invalid complaint ID format: %v", err)
		} else {
			complaintID = &objID
		}
	}
	if req.ComplaintInternalID != 0 {
		complaintNo = &req.ComplaintInternalID
	}

	var bookingID *primitive.ObjectID
	var bookingNo *int64
	if req.BookingID != "" {
		objID, err := primitive.ObjectIDFromHex(req.BookingID)
		if err != nil {
			log.Printf("Warning: Invalid booking ID format: %v", err)
		} else {
			bookingID = &objID
		}
	}
	if req.BookingInternalID != 0 {
		bookingNo = &req.BookingInternalID
	}

	gstAmount := req.Amount * 0.18
    mode := "UPI"
	refund := &domain.Refund{
		UserID:        req.UserID,
		BookingID:     bookingID,
		BookingNo:     bookingNo,
		ComplaintID:   complaintID,
		ComplaintNo:   complaintNo,
		TransactionID: fmt.Sprintf("TXN-%d", primitive.NewObjectID().Timestamp().Unix()),
		Reason:        req.Reason,
		Amount:        req.Amount,
		GST:           gstAmount,
		Mode:          mode,
		Status:        domain.RefundStatusPending,
	}

	if err := s.refundRepo.Create(ctx, refund); err != nil {
		log.Printf("ProcessRefund - Failed to create refund: %v", err)
		return fmt.Errorf("failed to create refund: %w", err)
	}

	log.Printf("ProcessRefund - Refund created successfully with ID: %s", refund.RefundID)

	go func() {
		bgCtx := context.Background()
		if err := s.refundRepo.UpdateStatus(bgCtx, refund.RefundID, domain.RefundStatusSuccess, ""); err != nil {
			log.Printf("ProcessRefund - Failed to update refund status: %v", err)
		} else {
			log.Printf("ProcessRefund - Refund %s marked as success", refund.RefundID)
		}
	}()

	return nil
}