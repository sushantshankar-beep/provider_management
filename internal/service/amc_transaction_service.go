package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AMCTransactionService struct {
	transactions repository.AMCTransactionRepo
	users        repository.UserRepo
	amcPurchases repository.AMCPurchaseRepo
}

func NewAMCTransactionService(t *repository.AMCTransactionRepo, u *repository.UserRepo, a *repository.AMCPurchaseRepo) *AMCTransactionService {
	return &AMCTransactionService{
		transactions: *t,
		users:        *u,
		amcPurchases: *a,
	}
}

type AMCTransactionListResponse struct {
	ID            string    `json:"id"`
	CreatedOn     time.Time `json:"createdOn"`
	Name          string    `json:"name"`
	Contact       string    `json:"contact"`
	Email         string    `json:"email"`
	Amount        string    `json:"amount"`
	PaymentMethod string    `json:"paymentMethod"`
	PaymentStatus string    `json:"paymentStatus"`
}

type AMCTransactionDetailResponse struct {
	AMCPurchaseID  string    `json:"amcPurchaseId"`
	OrderID        string    `json:"orderId"`
	User           string    `json:"user"`
	Email          string    `json:"email"`
	Contact        string    `json:"contact"`
	Amount         string    `json:"amount"`
	Currency       string    `json:"currency"`
	PaymentMethod  string    `json:"paymentMethod"`
	PaymentID      string    `json:"paymentId"`
	PaymentOrderID string    `json:"paymentOrderId"`
	Status         string    `json:"status"`
	TransactionType string   `json:"transactionType"`
	RefundStatus   string    `json:"refundStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	Message        string    `json:"message"`
}

func (s *AMCTransactionService) ListAMCTransactions(
	ctx context.Context,
	pageStr, limitStr, search, status, method string,
) ([]AMCTransactionListResponse, int64, error) {

	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	txns, total, err := s.transactions.FindAMCTransactions(ctx, skip, limit, status, method)
	if err != nil {
		return nil, 0, err
	}

	var result []AMCTransactionListResponse

	for _, txn := range txns {
		resp := AMCTransactionListResponse{
			ID:            strconv.FormatInt(txn.InternalID, 10),
			CreatedOn:     txn.CreatedAt,
			Name:          "N/A",
			Contact:       "N/A",
			Email:         "N/A",
			Amount:        fmt.Sprintf("₹%.0f", txn.Amount),
			PaymentMethod: "N/A",
			PaymentStatus: txn.Status,
		}

		if txnResp, ok := txn.TxnResponse.(map[string]interface{}); ok {
			if mode, exists := txnResp["mode"]; exists && mode != nil {
				resp.PaymentMethod = fmt.Sprintf("%v", mode)
			}
		}
		if resp.PaymentMethod == "N/A" && txn.Method != "" {
			resp.PaymentMethod = txn.Method
		}

		if txn.UserID != primitive.NilObjectID {
			if user, err := s.users.FindByID(ctx, txn.UserID.Hex()); err == nil {
				resp.Name = user.Name
				resp.Contact = user.Phone
				resp.Email = user.Email
			}
		}

		if search != "" {
			searchLower := strings.ToLower(search)
			nameLower := strings.ToLower(resp.Name)
			emailLower := strings.ToLower(resp.Email)
			contactLower := strings.ToLower(resp.Contact)

			if !strings.Contains(nameLower, searchLower) &&
				!strings.Contains(emailLower, searchLower) &&
				!strings.Contains(contactLower, searchLower) {
				continue
			}
		}

		result = append(result, resp)
	}

	return result, total, nil
}

func (s *AMCTransactionService) GetAMCTransaction(ctx context.Context, id string) (*AMCTransactionDetailResponse, error) {
	internalID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction id")
	}

	txn, err := s.transactions.FindByInternalID(ctx, internalID)
	if err != nil {
		return nil, err
	}

	resp := &AMCTransactionDetailResponse{
		OrderID:         strconv.FormatInt(txn.InternalID, 10),
		User:            "N/A",
		Email:           "N/A",
		Contact:         "N/A",
		Amount:          fmt.Sprintf("₹%.0f", txn.Amount),
		Currency:        txn.Currency,
		PaymentMethod:   "N/A",
		PaymentID:       "N/A",
		PaymentOrderID:  txn.TxnID,
		Status:          strings.Title(txn.Status),
		TransactionType: "Payment",
		RefundStatus:    "-",
		CreatedAt:       txn.CreatedAt,
		Message:         "Transaction processed successfully",
	}

	log.Println("kfjrcbjkc")
	if txnResp, ok := txn.TxnResponse.(map[string]interface{}); ok {
		log.Println("txnResp",txnResp)
		if mode, exists := txnResp["mode"]; exists && mode != nil {
			resp.PaymentMethod = fmt.Sprintf("%v", mode)
		}
		log.Println("txnResp",txnResp)
	}
	log.Println("txnResp",txn.Method)

	if resp.PaymentMethod == "N/A" && txn.Method != "" {
		resp.PaymentMethod = txn.Method 
	}

	if txn.MihPayID != "" {
		resp.PaymentID = txn.MihPayID
	}

	if txn.UserID != primitive.NilObjectID {
		if user, err := s.users.FindByID(ctx, txn.UserID.Hex()); err == nil {
			resp.User = user.Name
			resp.Email = user.Email
			resp.Contact = user.Phone
		}
	}

	if txn.AMCPurchaseID != primitive.NilObjectID {
		resp.AMCPurchaseID = txn.AMCPurchaseID.Hex()
		resp.TransactionType = "AMC Purchase"
	}

	return resp, nil
}