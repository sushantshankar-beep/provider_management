package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type PaymentPayoutStatus string

const (
	PayoutStatusPending           PaymentPayoutStatus = "pending"
	PayoutStatusPartiallySettled  PaymentPayoutStatus = "partially settled"
	PayoutStatusSettled           PaymentPayoutStatus = "settled"
)

type PaymentPayout struct {
	ID                primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	PayoutID          int64                `bson:"payoutId,omitempty" json:"payout_id"`
	ProviderID        primitive.ObjectID   `bson:"providerId" json:"provider_id"`
	ServiceIDs        []primitive.ObjectID `bson:"serviceIds" json:"service_ids"`
	BaseAmount        float64              `bson:"baseAmount" json:"base_amount"`
	PartialAmount        float64              `bson:"partialAmount" json:"partial_amount"`
	CommissionPercent float64              `bson:"commissionPercent" json:"commission_percent"`
	CommissionAmount  float64              `bson:"commissionAmount" json:"commission_amount"`
	GSTPercent        float64              `bson:"gstPercent" json:"gst_percent"`
	GSTAmount         float64              `bson:"gstAmount" json:"gst_amount"`
	NetPayable        float64              `bson:"netPayable" json:"net_payable"`
	SettlementID      *primitive.ObjectID  `bson:"settlementId,omitempty" json:"settlement_id,omitempty"`
	Status            PaymentPayoutStatus  `bson:"status" json:"status"`
	PeriodFrom        time.Time            `bson:"periodFrom" json:"period_from"`
	PeriodTo          time.Time            `bson:"periodTo" json:"period_to"`
	CreatedAt         time.Time            `bson:"createdAt" json:"created_at"`
	UpdatedAt         time.Time            `bson:"updatedAt" json:"updated_at"`
}



