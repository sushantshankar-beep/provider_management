package service

import (
	"fmt"
	"log"
	"time"
	"context"
	"provider_management/internal/dto"
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

func (s *TransactionService) GetTransactions( ctx context.Context, filters dto.TransactionFilters, pagination dto.PaginationParams ) (*dto.PaginatedResponse, error) {

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
	   pagination.Limit = 20
	}

	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	txns, total, err := s.transactions.FindWithFilter(ctx, filters, pagination)

	if err != nil {
		return nil, err
	}
	
	result := make([]dto.TransactionResponse, 0, len(txns))

	for _, txn := range txns {
		indianTime := txn.CreatedAt.Add(5*time.Hour + 30*time.Minute)
		resp := dto.TransactionResponse{
			ID:            txn.ID.Hex(),
			TxnID:         txn.TxnID,
			Amount:        txn.Amount,
			Currency:      txn.Currency,
			Status:        txn.Status,
			Method:        txn.Method,
			PaymentSource: txn.PaymentSource,
			CreatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
			UpdatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
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

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pagination.Limit - 1) / pagination.Limit
	}

	return &dto.PaginatedResponse{
		Data: result,
		Pagination: dto.PaginationMeta{
			HasNext:     pagination.Page < totalPages,
			HasPrevious: pagination.Page > 1,
			Limit:       pagination.Limit,
			Page:        pagination.Page,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}, nil
}

func (s *TransactionService) GetTransactionById(ctx context.Context, id string) (*dto.TransactionResponse, error) {
	txn, err := s.transactions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	indianTime := txn.CreatedAt.Add(5*time.Hour + 30*time.Minute)

	resp := dto.TransactionResponse{
		ID:            txn.ID.Hex(),
		TxnID:         txn.TxnID,
		Amount:        txn.Amount,
		Currency:      txn.Currency,
		Status:        txn.Status,
		Method:        txn.Method,
		PaymentSource: txn.PaymentSource,
		CreatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
		UpdatedAt:     indianTime.Format("2006-01-02 15:04:05"), 

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
	
	return &resp, nil
}
