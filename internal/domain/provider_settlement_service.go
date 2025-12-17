package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProviderSettlement struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	SettlementID  int64                `bson:"settlementId" json:"settlement_id"`
	PayoutID      primitive.ObjectID   `bson:"payoutId" json:"payout_id"`
	ProviderID    primitive.ObjectID   `bson:"providerId" json:"provider_id"`
	ProviderName  string               `bson:"providerName" json:"provider_name"`
	AccountNo     string               `bson:"accountNo" json:"account_no"`
	IfscCode      string               `bson:"ifscCode" json:"ifsc_code"`
	TotalAmount   float64              `bson:"totalAmount" json:"total_amount"`
	PaymentMode   string               `bson:"paymentMode" json:"payment_mode"`
	PaymentMethod string               `bson:"paymentMethod" json:"payment_method"`
	Justification string               `bson:"justification,omitempty" json:"justification,omitempty"`
	Status        string               `bson:"status" json:"status"`
	SettledAt     *time.Time           `bson:"settledAt,omitempty" json:"settled_at,omitempty"`
	CreatedAt     time.Time            `bson:"createdAt" json:"created_at"`
}
