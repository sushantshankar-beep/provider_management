package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type PaymentPayoutType string

const (
	PayoutTypeRegular   PaymentPayoutType = "regular"
	PayoutTypeComplaint PaymentPayoutType = "complaint"
)

type PaymentPayoutStatus string

const (
	PayoutStatusPending          PaymentPayoutStatus = "pending"
	PayoutStatusPartiallySettled PaymentPayoutStatus = "partially settled"
	PayoutStatusSettled          PaymentPayoutStatus = "settled"
)

type PaymentPayout struct {
	ID                    primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	PayoutID              int64                 `bson:"payoutId,omitempty" json:"payout_id"`
	ProviderID            primitive.ObjectID    `bson:"providerId" json:"provider_id"`
	ServiceIDs            []primitive.ObjectID  `bson:"serviceIds" json:"service_ids"`
	ComplaintID           *primitive.ObjectID   `bson:"complaintId,omitempty" json:"complaint_id,omitempty"`
	ComplaintInternalID   string                `bson:"complaintInternalId,omitempty" json:"complaint_internal_id,omitempty"`
	ServicePartialAmounts map[string]float64    `bson:"servicePartialAmounts,omitempty"`
	TotalPayAmount        float64               `bson:"totalPayAmount" json:"total_pay_amount"`
	BaseAmount            float64               `bson:"baseAmount" json:"base_amount"`
	TDSPercent            float64                      `bson:"tdsPercent" json:"tds_percent"`
	TDSAmount             float64               `bson:"tdsAmount" json:"tds_amount"`
	PartialAmount         float64               `bson:"partialAmount" json:"partial_amount"`
	CommissionPercent     float64               `bson:"commissionPercent" json:"commission_percent"`
	CommissionAmount      float64               `bson:"commissionAmount" json:"commission_amount"`
	GSTPercent            float64               `bson:"gstPercent" json:"gst_percent"`
	GSTAmount             float64               `bson:"gstAmount" json:"gst_amount"`
	NetPayable            float64               `bson:"netPayable" json:"net_payable"`
	SettlementID          *primitive.ObjectID   `bson:"settlementId,omitempty" json:"settlement_id,omitempty"`
	ComplaintAdjustments  []ComplaintAdjustment `bson:"complaintAdjustments,omitempty" json:"complaint_adjustments,omitempty"`
	Status                PaymentPayoutStatus   `bson:"status" json:"status"`
	PeriodFrom            time.Time             `bson:"periodFrom" json:"period_from"`
	PeriodTo              time.Time             `bson:"periodTo" json:"period_to"`
	PayoutType            PaymentPayoutType     `bson:"payoutType" json:"payout_type"`
	IsDeduction           bool                  `bson:"isDeduction" json:"is_deduction"`
	IsPayoutCancelled     bool                  `bson:"isPayoutCancelled"`
	CreatedAt             time.Time             `bson:"createdAt" json:"created_at"`
	UpdatedAt             time.Time             `bson:"updatedAt" json:"updated_at"`
	Remarks               string                `bson:"remarks,omitempty" json:"remarks,omitempty"`
}

type ComplaintAdjustment struct {
	ServiceID           primitive.ObjectID `bson:"serviceId" json:"service_id"`
	ComplaintID         primitive.ObjectID `bson:"complaintId" json:"complaint_id"`
	ComplaintInternalID int64              `bson:"complaintInternalId" json:"complaint_internal_id"`
	Amount              float64            `bson:"amount" json:"amount"`
	CreatedAt           time.Time          `bson:"createdAt" json:"created_at"`
}
