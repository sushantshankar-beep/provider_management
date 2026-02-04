package service

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"provider_management/internal/domain"
	"strings"
	"math/rand"
	"time"
)

type PayUService struct {
	key     string
	salt    string
	baseURL string
}

func NewPayUService(key, salt, baseURL string) *PayUService {
	return &PayUService{
		key:     key,
		salt:    salt,
		baseURL: baseURL,
	}
}

type InitiateRefundResult struct {
	Success             bool                   `json:"success"`
	RequestID           string                 `json:"requestId"`
	RefundTransactionID string                 `json:"refundTransactionId"`
	PayUResponse        map[string]interface{} `json:"payuResponse"`
}

type CheckRefundStatusResult struct {
	RefundStatus string                 `json:"refundStatus"`
	Amount       float64                `json:"amount"`
	BankRefNum   string                 `json:"bankRefNum"`
	Mode         string                 `json:"mode"`
	SettlementID string                 `json:"settlementId"`
	BankArn      string                 `json:"bankArn"`
	ErrorMsg     string                 `json:"errorMsg"`
	RawResponse  map[string]interface{} `json:"rawResponse"`
}

func (s *PayUService) generateHash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *PayUService) InitiateRefund(ctx context.Context, refund *domain.AMCRefundRequest, purchase *domain.AMCPurchase) (*InitiateRefundResult, error) {
	if purchase.PayuTransactionID == "" {
		return nil, fmt.Errorf("Missing PayU Transaction ID")
	}

	command := "cancel_refund_transaction"
	var1 := purchase.PayuTransactionID
	var2 := fmt.Sprintf("REF%s%d", refund.ID.Hex()[len(refund.ID.Hex())-6:], time.Now().Unix()%1000000)
	var3 := fmt.Sprintf("%.2f", refund.RefundAmount)

	hash := s.generateHash(fmt.Sprintf("%s|%s|%s|%s", s.key, command, var1, s.salt))

	formData := url.Values{}
	formData.Set("key", s.key)
	formData.Set("command", command)
	formData.Set("var1", var1)
	formData.Set("var2", var2)
	formData.Set("var3", var3)
	formData.Set("hash", hash)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/merchant/postservice.php?form=2", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	status, _ := data["status"].(float64)
	if status != 1 {
		msg, _ := data["msg"].(string)
		if msg == "" {
			msg = "Refund failed"
		}
		return nil, fmt.Errorf(msg)
	}

	requestID, _ := data["request_id"].(string)
	if requestID == "" {
		requestID = var2
	}

	refundTransactionID, _ := data["mihpayid"].(string)
	if refundTransactionID == "" {
		refundTransactionID = var1
	}

	return &InitiateRefundResult{
		Success:             true,
		RequestID:           requestID,
		RefundTransactionID: refundTransactionID,
		PayUResponse:        data,
	}, nil
}

// func (s *PayUService) CheckRefundStatus(ctx context.Context, requestID string) (*CheckRefundStatusResult, error) {
// 	command := "check_action_status"

// 	hash := s.generateHash(fmt.Sprintf("%s|%s|%s|%s", s.key, command, requestID, s.salt))

// 	formData := url.Values{}
// 	formData.Set("key", s.key)
// 	formData.Set("command", command)
// 	formData.Set("var1", requestID)
// 	formData.Set("var2", "request_id")
// 	formData.Set("hash", hash)

// 	client := &http.Client{Timeout: 30 * time.Second}
// 	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/merchant/postservice.php?form=2", strings.NewReader(formData.Encode()))
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var data map[string]interface{}
// 	if err := json.Unmarshal(body, &data); err != nil {
// 		return nil, err
// 	}

// 	result := &CheckRefundStatusResult{
// 		RefundStatus: "pending",
// 		RawResponse:  data,
// 	}

// 	transactionDetails, ok := data["transaction_details"].(map[string]interface{})
// 	if !ok {
// 		return result, nil
// 	}

// 	requestIDData, ok := transactionDetails[requestID].(map[string]interface{})
// 	if !ok {
// 		return result, nil
// 	}

// 	txnDetails, ok := requestIDData[requestID].(map[string]interface{})
// 	if !ok {
// 		return result, nil
// 	}

// 	if status, ok := txnDetails["status"].(string); ok {
// 		result.RefundStatus = strings.ToLower(status)
// 	}

// 	if amt, ok := txnDetails["amt"].(float64); ok {
// 		result.Amount = amt
// 	}

// 	if bankRefNum, ok := txnDetails["bank_ref_num"].(string); ok {
// 		result.BankRefNum = bankRefNum
// 	}

// 	if mode, ok := txnDetails["refund_mode"].(string); ok {
// 		result.Mode = mode
// 	}

// 	if settlementID, ok := txnDetails["settlement_id"].(string); ok {
// 		result.SettlementID = settlementID
// 	}

// 	if bankArn, ok := txnDetails["bank_arn"].(string); ok {
// 		result.BankArn = bankArn
// 	}

// 	if errorMsg, ok := txnDetails["error"].(string); ok {
// 		result.ErrorMsg = errorMsg
// 	}

// 	return result, nil
// }

func (s *PayUService) InitiateRefundByParams(ctx context.Context, payuTransactionID, refundID string, amount float64) (*InitiateRefundResult, error) {
	if payuTransactionID == "" {
		return nil, fmt.Errorf("Missing PayU Transaction ID")
	}

	command := "cancel_refund_transaction"
	var1 := payuTransactionID
	var2 := refundID
	var3 := fmt.Sprintf("%.2f", amount)

	hash := s.generateHash(fmt.Sprintf("%s|%s|%s|%s", s.key, command, var1, s.salt))

	formData := url.Values{}
	formData.Set("key", s.key)
	formData.Set("command", command)
	formData.Set("var1", var1)
	formData.Set("var2", var2)
	formData.Set("var3", var3)
	formData.Set("hash", hash)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/merchant/postservice.php?form=2", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	status, _ := data["status"].(float64)
	if status != 1 {
		msg, _ := data["msg"].(string)
		if msg == "" {
			msg = "Refund failed"
		}
		return nil, fmt.Errorf(msg)
	}

	requestID, _ := data["request_id"].(string)
	if requestID == "" {
		requestID = var2
	}

	refundTransactionID, _ := data["mihpayid"].(string)
	if refundTransactionID == "" {
		refundTransactionID = var1
	}

	return &InitiateRefundResult{
		Success:             true,
		RequestID:           requestID,
		RefundTransactionID: refundTransactionID,
		PayUResponse:        data,
	}, nil
}

func (s *PayUService) CheckRefundStatus(ctx context.Context, requestID string) (*CheckRefundStatusResult, error) {

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(2) // 0 or 1

	// ❌ FAILED
	if random == 0 {
		return &CheckRefundStatusResult{
			RefundStatus: "failed",
			Amount:       250,
			ErrorMsg:     "Insufficient balance at bank",
			RawResponse: map[string]interface{}{
				"mock":   true,
				"status": "failed",
			},
		}, nil
	}

	// ✅ SUCCESS
	return &CheckRefundStatusResult{
		RefundStatus: "success",
		Amount:       250,
		BankRefNum:   fmt.Sprintf("TEST_REF_%d", rand.Intn(999999)),
		Mode:         "original",
		SettlementID: fmt.Sprintf("SETT_%d", rand.Intn(999999)),
		BankArn:      fmt.Sprintf("ARN_%d", rand.Intn(999999)),
		RawResponse: map[string]interface{}{
			"mock":   true,
			"status": "success",
		},
	}, nil
}