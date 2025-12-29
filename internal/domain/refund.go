package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "pending"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusSuccess    RefundStatus = "success"
	RefundStatusFailed     RefundStatus = "failed"
)

// type RefundMode string

// const (
// 	RefundModePaytm     RefundMode = "Paytm"
// 	RefundModeGooglePay RefundMode = "Google Pay"
// 	RefundModePhonePe   RefundMode = "PhonePe"
// 	RefundModeRazorpay  RefundMode = "Razorpay"
// 	RefundModeUPI       RefundMode = "UPI"
// 	RefundModeBank      RefundMode = "Bank"
// )

type Refund struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	TransactionID string             `bson:"transactionId,omitempty" json:"transaction_id,omitempty"`
	RefundID      string              `bson:"refundId" json:"refund_id"`
	UserID        string              `bson:"userId" json:"user_id"`
	BookingID     *primitive.ObjectID `bson:"bookingId,omitempty" json:"booking_id,omitempty"`
	BookingNo     *int64              `bson:"bookingNo,omitempty" json:"booking_no,omitempty"`
	ComplaintID   *primitive.ObjectID `bson:"complaintId,omitempty" json:"complaint_id,omitempty"`
	ComplaintNo   *int64              `bson:"complaintNo,omitempty" json:"complaint_no,omitempty"`
	Reason        string              `bson:"reason" json:"reason"`
	Amount        float64             `bson:"amount" json:"amount"`
	GST           float64             `bson:"gst" json:"gst"`
	Mode          string              `bson:"mode" json:"mode"`
	Status        RefundStatus        `bson:"status" json:"status"`
	ProcessedAt   *time.Time          `bson:"processedAt,omitempty" json:"processed_at,omitempty"`
	FailureReason string              `bson:"failureReason,omitempty" json:"failure_reason,omitempty"`
	CreatedAt     time.Time           `bson:"createdAt" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updatedAt" json:"updated_at"`
}
type RefundFilter struct {
	Page   int
	Limit  int
	Status string
	UserID string
	Search string
}
