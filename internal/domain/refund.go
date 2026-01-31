package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RefundStatus string

const (
	RefundStatusPending      RefundStatus = "pending"
	RefundStatusUnderProcess RefundStatus = "under_process"
	RefundStatusSuccess      RefundStatus = "success"
	RefundStatusFailed       RefundStatus = "failed"
)

type Refund struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	RefundID          string             `bson:"refundId" json:"refund_id"`
	TxnID             string             `bson:"txnid"`
	UserID            string             `bson:"userId" json:"userId"`
	ServiceID         string             `bson:"serviceId" json:"serviceId"`
	ComplaintID       string             `bson:"complaintId,omitempty" json:"complaint_id,omitempty"`
	Reason            string             `bson:"reason" json:"reason"`
	Amount            float64            `bson:"amount" json:"amount"`
	GST               float64            `bson:"gst" json:"gst"`
	NetRefund         float64            `bson:"netRefund" json:"net_refund"`
	Mode              string             `bson:"mode" json:"mode"`
	Status            RefundStatus       `bson:"status" json:"status"`
	Notes             []RefundNote     `bson:"notes,omitempty" json:"notes,omitempty"`
	PayURequestID     string             `bson:"payuRequestId,omitempty" json:"payu_request_id,omitempty"`
	PayUTransactionID string             `bson:"payuTransactionId,omitempty" json:"payu_transaction_id,omitempty"`
	BankRefNum        string             `bson:"bankRefNum,omitempty" json:"bank_ref_num,omitempty"`
	RefundMode        string             `bson:"refundMode,omitempty" json:"refund_mode,omitempty"`
	SettlementID      string             `bson:"settlementId,omitempty" json:"settlement_id,omitempty"`
	BankARN           string             `bson:"bankArn,omitempty" json:"bank_arn,omitempty"`
	PayUResponse      string             `bson:"payuResponse,omitempty" json:"payu_response,omitempty"`
	ProcessedAt       *time.Time         `bson:"processedAt,omitempty" json:"processed_at,omitempty"`
	FailureReason     string             `bson:"failureReason,omitempty" json:"failure_reason,omitempty"`
	InitiatedAt       *time.Time         `bson:"initiatedAt,omitempty" json:"initiated_at,omitempty"`
	StatusCheckedAt   *time.Time         `bson:"statusCheckedAt,omitempty" json:"status_checked_at,omitempty"`
	CompletedAt       *time.Time         `bson:"completedAt,omitempty" json:"completed_at,omitempty"`
	Timeline          RefundTimeline     `bson:"timeline" json:"timeline"`
	CreatedAt         time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updatedAt" json:"updated_at"`
}
type RefundFilter struct {
	Page   int
	Limit  int
	Status string
	Mode   string
	Reason string
	UserID string
	Search string
}

type RefundTimeline struct {
	Initiated     time.Time `bson:"initiated,omitempty" json:"initiated,omitempty"`
	UnderProcess  time.Time `bson:"underProcess,omitempty" json:"under_process,omitempty"`
	StatusChecked time.Time `bson:"statusChecked,omitempty" json:"status_checked,omitempty"`
	Completed     time.Time `bson:"completed,omitempty" json:"completed,omitempty"`
	Failed        time.Time `bson:"failed,omitempty" json:"failed,omitempty"`
}

type RefundNote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Content   string             `bson:"content" json:"content"`
	AddedBy   string             `bson:"addedBy" json:"addedBy"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
