package domain

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type Transaction struct {
	ID           string `json:"_id" bson:"_id,omitempty"`
	UserID        string `json:"user_id" bson:"userId"`
	ServiceID     string `json:"service_id" bson:"serviceId"`
	AMCPurchaseID primitive.ObjectID `json:"amc_purchase_id" bson:"AMCPurchaseId,omitempty"`
	InternalID    int64              `json:"id" bson:"id"`
	TxnID         string             `json:"txnid" bson:"txnid"`
	Amount        float64            `json:"amount" bson:"amount"`
	Currency      string             `json:"currency" bson:"currency"`
	Status        string             `json:"status" bson:"status"`
	Method        string             `json:"method,omitempty" bson:"method,omitempty"`
	PaymentSource string             `json:"payment_source" bson:"paymentSource"`
	ErrorMessage  string    `bson:"errorMessage,omitempty"`
	MihPayID      string             `json:"mihpayid,omitempty" bson:"mihpayid,omitempty"`
	TxnResponse   any                `json:"txn_response,omitempty" bson:"txnResponse,omitempty"`
	RefundID      *primitive.ObjectID `json:"refund_id,omitempty" bson:"refundId,omitempty"`
	CreatedAt     time.Time          `json:"created_at" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updatedAt"`
	InvoiceGenerated bool `bson:"invoiceGenerated" json:"invoiceGenerated"`
	FailReason string `bson:"failReason,omitempty" json:"failReason,omitempty"`
}


type PaymentStatus string

const (
	PaymentPending       PaymentStatus = "pending"
	PaymentPaid          PaymentStatus = "paid"
	PaymentFailed        PaymentStatus = "failed"
	PaymentFailedGrace   PaymentStatus = "failed_pending_release"
)