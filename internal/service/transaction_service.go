package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strconv"
	"provider_management/internal/repository"
)

type TransactionService struct {
	transactions repository.TransactionRepo
	services     repository.AcceptedServiceRepo
	users        repository.UserRepo
}

func NewTransactionService(t *repository.TransactionRepo, s *repository.AcceptedServiceRepo, u *repository.UserRepo) *TransactionService {
	return &TransactionService{
		transactions: *t,
		services:     *s,
		users:        *u,
	}
}

type TransactionResponse struct {
	ID            string  `json:"_id"`
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name,omitempty"`
	BookingID     string  `json:"booking_id,omitempty"`
	TxnID         string  `json:"txnid"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Method        string  `json:"method,omitempty"`
	PaymentSource string  `json:"payment_source,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	BookingNo     string  `json:"booking_no,omitempty"`
}

func (s *TransactionService) ListTransactions(
	ctx context.Context,
	pageStr, limitStr, search string,
) ([]TransactionResponse, int64, error) {

	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	txns, total, err := s.transactions.FindWithFilter(ctx, skip, limit, search)
	if err != nil {
		return nil, 0, err
	}

	var result []TransactionResponse

	for _, txn := range txns {
		resp := TransactionResponse{
			ID:            txn.ID.Hex(),
			TxnID:         txn.TxnID,
			Amount:        txn.Amount,
			Currency:      txn.Currency,
			Status:        txn.Status,
			Method:        txn.Method,
			PaymentSource: txn.PaymentSource,
			CreatedAt:     txn.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     txn.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if txn.UserID != primitive.NilObjectID {
			if user, err := s.users.FindByID(ctx, txn.UserID.Hex()); err == nil {
				resp.UserID = fmt.Sprintf("VW%d", user.InternalID)
				resp.UserName = user.Name
			}
		}
		
		if txn.ServiceID != primitive.NilObjectID {
			if as, err := s.services.FindByID(ctx, txn.ServiceID.Hex()); err == nil {
				resp.BookingID = as.ID.Hex()
				resp.BookingNo = fmt.Sprintf("BK%d", as.InternalID)
			}
		}

		result = append(result, resp)
	}

	return result, total, nil
}


func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*TransactionResponse, error) {
	txn, err := s.transactions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &TransactionResponse{
		ID:            txn.ID.Hex(),
		TxnID:         txn.TxnID,
		Amount:        txn.Amount,
		Currency:      txn.Currency,
		Status:        txn.Status,
		Method:        txn.Method,
		PaymentSource: txn.PaymentSource,
		CreatedAt:     txn.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     txn.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if txn.UserID != primitive.NilObjectID {
		user, err := s.users.FindByID(ctx, txn.UserID.Hex())
		if err == nil {
			resp.UserID = fmt.Sprintf("VW%d", user.InternalID)
			resp.UserName = user.Name
		} else {
			log.Printf("Failed to fetch user %s: %v", txn.UserID.Hex(), err)
		}
	}

	if txn.ServiceID != primitive.NilObjectID {
		acceptedService, err := s.services.FindByID(ctx, txn.ServiceID.Hex())
		if err == nil {
			resp.BookingID = acceptedService.ID.Hex()
			resp.BookingNo = fmt.Sprintf("BK%d", acceptedService.InternalID)
		} else {
			log.Printf("Failed to fetch accepted service %s: %v", txn.ServiceID.Hex(), err)
		}
	}
	

	return resp, nil
}
