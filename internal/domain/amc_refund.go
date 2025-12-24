package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Timeline struct {
	Status      string    `bson:"status" json:"status"`
	Timestamp   time.Time `bson:"timestamp" json:"timestamp"`
	Description string    `bson:"description" json:"description"`
	Note        string    `bson:"note,omitempty" json:"note,omitempty"`
}

type AMCRefundRequest struct {
	ID                            primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	AMCPurchaseID                 primitive.ObjectID     `bson:"amcPurchase" json:"amcPurchase"`
	UserID                        primitive.ObjectID     `bson:"user" json:"user"`
	TotalAmount                   float64                `bson:"totalAmount" json:"totalAmount"`
	GSTAmount                     float64                `bson:"gstAmount" json:"gstAmount"`
	ProcessingFee                 float64                `bson:"processingFee" json:"processingFee"`
	RefundAmount                  float64                `bson:"refundAmount" json:"refundAmount"`
	Reason                        string                 `bson:"reason" json:"reason"`
	Status                        string                 `bson:"status" json:"status"`
	Timeline                      []Timeline             `bson:"timeline" json:"timeline"`
	ApprovedAt                    *time.Time             `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`
	CompletedAt                   *time.Time             `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	RejectedAt                    *time.Time             `bson:"rejectedAt,omitempty" json:"rejectedAt,omitempty"`
	CancelledAt                   *time.Time             `bson:"cancelledAt,omitempty" json:"cancelledAt,omitempty"`
	RejectionReason               string                 `bson:"rejectionReason,omitempty" json:"rejectionReason,omitempty"`
	RefundTransactionID           string                 `bson:"refundTransactionId,omitempty" json:"refundTransactionId,omitempty"`
	PayURequestID                 string                 `bson:"payuRequestId,omitempty" json:"payuRequestId,omitempty"`
	RefundToken                   string                 `bson:"refundToken,omitempty" json:"refundToken,omitempty"`
	PayURefundID                  string                 `bson:"payuRefundId,omitempty" json:"payuRefundId,omitempty"`
	PayURefundResponse            map[string]interface{} `bson:"payuRefundResponse" json:"payuRefundResponse"`
	PayURefundStatusCheckResponse map[string]interface{} `bson:"payuRefundStatusCheckResponse,omitempty" json:"payuRefundStatusCheckResponse,omitempty"`
	LastStatusCheckAt             *time.Time             `bson:"lastStatusCheckAt,omitempty" json:"lastStatusCheckAt,omitempty"`
	RefundMode                    string                 `bson:"refundMode" json:"refundMode"`
	EstimatedDays                 int                    `bson:"estimatedDays" json:"estimatedDays"`
	ProcessedBy                   primitive.ObjectID     `bson:"processedBy,omitempty" json:"processedBy,omitempty"`
	BankRefNum                    string                 `bson:"bankRefNum,omitempty" json:"bankRefNum,omitempty"`
	SettlementID                  string                 `bson:"settlementId,omitempty" json:"settlementId,omitempty"`
	BankArn                       string                 `bson:"bankArn,omitempty" json:"bankArn,omitempty"`
	CreatedAt                     time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt                     time.Time              `bson:"updatedAt" json:"updatedAt"`
}
