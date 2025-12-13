package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID           string  `json:"_id"`
	UserID       string  `json:"user_id"`
	UserName     string  `json:"user_name,omitempty"`
	BookingID    string  `json:"booking_id,omitempty"`
	TxnID        string  `json:"txnid"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	Method       string  `json:"method,omitempty"`
	PaymentSource string `json:"payment_source,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	BookingNo    string  `json:"booking_no,omitempty"`
}

func (s *TransactionService) ListTransactions(ctx context.Context) ([]TransactionResponse, error) {
	txns, err := s.transactions.FindAll(ctx, 0, 100)
	if err != nil {
		return nil, err
	}

	var result []TransactionResponse
	for _, txn := range txns {
		resp := TransactionResponse{
			ID:           txn.ID.Hex(),
			TxnID:        txn.TxnID,
			Amount:       txn.Amount,
			Currency:     txn.Currency,
			Status:       txn.Status,
			Method:       txn.Method,
			PaymentSource: txn.PaymentSource,
			CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    txn.UpdatedAt.Format("2006-01-02 15:04:05"),
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
			service, err := s.services.FindByID(ctx, txn.ServiceID.Hex())
			if err == nil {
				resp.BookingID = service.ID

				resp.BookingNo = fmt.Sprintf("BK%d", service.ServiceRequestNo)
			} else {
				log.Printf("Failed to fetch service %s: %v", txn.ServiceID.Hex(), err)
			}
		}

		result = append(result, resp)
	}

	return result, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*TransactionResponse, error) {
	txn, err := s.transactions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &TransactionResponse{
		ID:           txn.ID.Hex(),
		TxnID:        txn.TxnID,
		Amount:       txn.Amount,
		Currency:     txn.Currency,
		Status:       txn.Status,
		Method:       txn.Method,
		PaymentSource: txn.PaymentSource,
		CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    txn.UpdatedAt.Format("2006-01-02 15:04:05"),
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
		service, err := s.services.FindByID(ctx, txn.ServiceID.Hex())
		if err == nil {
			resp.BookingID = service.ID
			resp.BookingNo = fmt.Sprintf("BK%d", service.ServiceRequestNo)
		} else {
			log.Printf("Failed to fetch service %s: %v", txn.ServiceID.Hex(), err)
		}
	}

	return resp, nil
}
