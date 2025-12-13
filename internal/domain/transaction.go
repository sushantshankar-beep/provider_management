package domain

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type Transaction struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"userId"`
	ServiceID primitive.ObjectID `json:"service_id" bson:"serviceId"`
	InternalID int64   `json:"id" bson:"id"`
	TxnID      string  `json:"txnid" bson:"txnid"`
	Amount    float64 `json:"amount" bson:"amount"`
	Currency  string  `json:"currency" bson:"currency"`
	Status    string  `json:"status" bson:"status"`
	Method    string  `json:"method,omitempty" bson:"method,omitempty"`
	PaymentSource string `json:"payment_source" bson:"paymentSource"`
	MihPayID string `json:"mihpayid,omitempty" bson:"mihpayid,omitempty"`
	TxnResponse any `json:"txn_response,omitempty" bson:"txnResponse,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"createdAt"`
	UpdatedAt time.Time `json:"updated_at" bson:"updatedAt"`
}
