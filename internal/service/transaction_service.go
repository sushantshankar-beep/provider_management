package service

import (
	"context"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"
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
			ID:            txn.ID,
			TxnID:         txn.TxnID,
			Amount:        txn.Amount,
			Status:        txn.Status,
			Method:        txn.Method,
			PaymentSource: txn.PaymentSource,
			CreatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
			UpdatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
		}

		if txn.ServiceID != "" {
			service, err := s.services.FindByID(ctx, txn.ServiceID)
			if err == nil && service != nil {
				resp.BookingNo = service.ServiceNumber

				if service.User.Hex() != "" {
					user, err := s.users.FindByID(ctx, service.User.Hex())
					if err == nil && user != nil {
						resp.UserID = user.UserCode
						resp.UserName = user.Name
					}
				}
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
		ID:            txn.ID,
		TxnID:         txn.TxnID,
		Amount:        txn.Amount,
		Status:        txn.Status,
		Method:        txn.Method,
		PaymentSource: txn.PaymentSource,
		CreatedAt:     indianTime.Format("2006-01-02 15:04:05"), 
		UpdatedAt:     indianTime.Format("2006-01-02 15:04:05"), 

	}

	if txn.ServiceID != "" {
		service, err := s.services.FindByID(ctx, txn.ServiceID)
		if err == nil && service != nil {
			resp.BookingNo = service.ServiceNumber

			if service.User.Hex() != "" {
				user, err := s.users.FindByID(ctx, service.User.Hex())
				if err == nil && user != nil {
					resp.UserID = user.UserCode
					resp.UserName = user.Name
				}
			}
		}
	}
	
	return &resp, nil
}
