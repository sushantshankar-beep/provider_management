package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SettlementStatus string

const (
	SettleStatusPending    SettlementStatus = "pending"
	SettleStatusSettled    SettlementStatus = "settled"
)
type SettledPostData struct {
	TransactionID string  `bson:"transactionId" json:"transaction_id"`
	Amount        float64 `bson:"amount" json:"amount"`
	Status        string  `bson:"status" json:"status"`
	Note          string  `bson:"note" json:"note"`
}
type ProviderSettlement struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	SettlementID  int64                `bson:"settlementId" json:"settlement_id"`
	PayoutID      primitive.ObjectID   `bson:"payoutId" json:"payout_id"`
	ProviderID    primitive.ObjectID   `bson:"providerId" json:"provider_id"`
	PayoutIdNumber  int64              `json:"payoutIdNumber"` 
	ProviderName  string               `bson:"providerName" json:"provider_name"`
	AccountNo     string               `bson:"accountNo" json:"account_no"`
	IfscCode      string               `bson:"ifscCode" json:"ifsc_code"`
	TotalAmount   float64              `bson:"totalAmount" json:"total_amount"`
	PaymentMode   string               `bson:"paymentMode" json:"payment_mode"`
	DeductionAmount   float64             `bson:"deductionAmount" json:"deduction_amount"`   
	PaymentMethod string               `bson:"paymentMethod" json:"payment_method"`
	Justification string               `bson:"justification,omitempty" json:"justification,omitempty"`
	Status        SettlementStatus     `bson:"status" json:"status"`
	SettledPostData *SettledPostData   `bson:"settledPostData,omitempty" json:"settled_post_data,omitempty"`
	SettledAt     *time.Time           `bson:"settledAt,omitempty" json:"settled_at,omitempty"`
	CreatedAt     time.Time            `bson:"createdAt" json:"created_at"`
}

