package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
     "provider_management/internal/utils"

	"strconv"
	"time"
)

type RefundService struct {
	refundRepo      *repository.RefundRepository
	transactionRepo *repository.TransactionRepo
	userRepo   *repository.UserRepo
}

func NewRefundService(refundRepo *repository.RefundRepository, transactionRepo *repository.TransactionRepo,userRepo   *repository.UserRepo) *RefundService {
	return &RefundService{
		refundRepo:      refundRepo,
		transactionRepo: transactionRepo,
		userRepo: userRepo,
	}
}

type RefundListItem struct {
	ID            primitive.ObjectID  `json:"id"`
	RefundID      string              `json:"refund_id"`
	UserID        string              `json:"user_id"`
	BookingNo     string              `json:"booking_no"`
	ComplaintNo   string              `json:"complaint_no"`
	TransactionID string              `json:"transaction_id`
	GST           float64             `json:"gst"`
	Mode          string              `json:"mode"`
	Amount        float64             `json:"amount"`
	Reason        string              `json:"reason"`
	Status        domain.RefundStatus `json:"status"`
	CreatedAt     time.Time           `json:"created_at"`
}

type RefundDetail struct {
	ID            primitive.ObjectID  `json:"id"`
	RefundID      string              `json:"refund_id"`
	UserID        string              `json:"user_id"`
	Amount        float64             `json:"amount"`
	BookingNo     string              `json:"booking_no"`
	ComplaintNo   string              `json:"complaint_no"`
	ComplaintID   *primitive.ObjectID `json:"complaint_id"`
	TransactionID string              `json:"transaction_id`
	GST           float64             `json:"gst"`
	Mode          string              `json:"mode"`
	Reason        string              `json:"reason"`
	Status        domain.RefundStatus `json:"status"`
	CreatedAt     time.Time           `json:"created_at"`
}

type RefundListResponse struct {
	Refunds    []RefundListItem `json:"refunds"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int64            `json:"total_pages"`
}

func (s *RefundService) ProcessRefund(ctx context.Context, req RefundRequest) error {
	log.Printf("ProcessRefund - Starting refund for user %s, amount: %.2f", req.UserID, req.Amount)
	log.Println("request", req)

	transaction, err := s.transactionRepo.FindByTxnID(ctx, req.TxnID)
	if err != nil {
		return fmt.Errorf("transaction not found with txnid %s: %w", req.TxnID, err)
	}

	log.Printf("Found transaction: ID=%s, TxnID=%s, Amount=%.2f, Method=%s",
		transaction.ID.Hex(), transaction.TxnID, transaction.Amount, transaction.Method)

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
	refund := &domain.Refund{
		UserID:        req.UserID,
		BookingID:     bookingID,
		BookingNo:     bookingNo,
		ComplaintID:   complaintID,
		ComplaintNo:   complaintNo,
		TransactionID: transaction.TxnID,
		Reason:        req.Reason,
		Amount:        utils.RoundTo2(req.Amount),
		GST:           utils.RoundTo2(gstAmount),
		Mode:          transaction.Method,
		Status:        domain.RefundStatusPending,
	}

	if err := s.refundRepo.Create(ctx, refund); err != nil {
		log.Printf("ProcessRefund - Failed to create refund: %v", err)
		return fmt.Errorf("failed to create refund: %w", err)
	}

	log.Printf("ProcessRefund - Refund created successfully with ID: %s", refund.RefundID)

	if err := s.transactionRepo.UpdateRefundID(ctx, transaction.ID, refund.ID); err != nil {
		log.Printf("ProcessRefund - Failed to update transaction with refund ID: %v", err)
	}

	return nil
}

func (s *RefundService) GetAllRefunds(
	ctx context.Context,
	filter domain.RefundFilter,
) (*RefundListResponse, error) {

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	skip := (filter.Page - 1) * filter.Limit

	refunds, total, err := s.refundRepo.FindAll(ctx, filter, skip, filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refunds: %w", err)
	}

	userIDs := make([]string, 0, len(refunds))
	for _, r := range refunds {
		userIDs = append(userIDs, r.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		log.Printf("GetAllRefunds - Failed to fetch users: %v", err)
	}

	userMap := make(map[string]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	items := make([]RefundListItem, 0, len(refunds))
	for _, r := range refunds {
		formattedUserID := r.UserID
		if user, ok := userMap[r.UserID]; ok && user.InternalID > 0 {
			formattedUserID = fmt.Sprintf("VW%d", user.InternalID)
		}

		items = append(items, RefundListItem{
			ID:            r.ID,
			RefundID:      r.RefundID,
			UserID:        formattedUserID,
			BookingNo:     formatWithPrefix("BK", r.BookingNo),
			ComplaintNo:   formatWithPrefix("CMP", r.ComplaintNo),
			TransactionID: r.TransactionID,
			GST:           utils.RoundTo2(r.GST),
			Mode:          r.Mode,
			Amount:        utils.RoundTo2(r.Amount),
			Status:        r.Status,
			Reason:        r.Reason,
			CreatedAt:     r.CreatedAt,
		})
	}

	return &RefundListResponse{
		Refunds: items,
		Total:   total,
		Page:    filter.Page,
		Limit:   filter.Limit,
		TotalPages: (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}
func (s *RefundService) GetRefundByID(
	ctx context.Context,
	refundID string,
) (*RefundDetail, error) {

	r, err := s.refundRepo.FindByRefundID(ctx, refundID)
	if err != nil {
		return nil, fmt.Errorf("refund not found: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, r.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found for refund: %w", err)
	}

	vwUserID := fmt.Sprintf("VW%d", user.InternalID)

	return &RefundDetail{
		ID:            r.ID,
		RefundID:      r.RefundID,
		UserID:        vwUserID,
		Amount:        utils.RoundTo2(r.Amount),
		BookingNo:     formatWithPrefix("BK", r.BookingNo),
		ComplaintNo:   formatWithPrefix("CMP", r.ComplaintNo),
		ComplaintID:   r.ComplaintID,
		TransactionID: r.TransactionID,
		GST:           utils.RoundTo2(r.GST),
		Mode:          r.Mode,
		Reason:        r.Reason,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt,
	}, nil
}

func formatWithPrefix(prefix string, v *int64) string {
	if v == nil {
		return ""
	}
	return prefix + strconv.FormatInt(*v, 10)
}
