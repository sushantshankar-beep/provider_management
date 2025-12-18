package service

import (
	"context"
	"fmt"
	"provider_management/internal/repository"
	"time"
)

type RefundService struct {
	userRepo        *repository.UserRepo
	transactionRepo *repository.TransactionRepo
}

func NewRefundService(
	userRepo *repository.UserRepo,
	transactionRepo *repository.TransactionRepo,
) *RefundService {
	return &RefundService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *RefundService) ProcessRefund(ctx context.Context, req RefundRequest) error {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	newBalance := user.WalletBalance + req.Amount
	if err := s.userRepo.UpdateWallet(ctx, req.UserID, newBalance); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	transaction := map[string]interface{}{
		"userId":      req.UserID,
		"type":        "refund",
		"amount":      req.Amount,
		"description": req.Reason,
		"complaintId": req.ComplaintID,
		"status":      "completed",
		"createdAt":   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}
