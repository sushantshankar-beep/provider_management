package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bid struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InternalID       int64              `bson:"id" json:"bid_id"`
	ServiceRequestID primitive.ObjectID `bson:"serviceRequest" json:"service_request_id"`
	ProviderID       primitive.ObjectID `bson:"provider" json:"provider_id"`
	OfferedPrice     float64            `bson:"offeredPrice" json:"offered_price"`
	BasePrice        *float64           `bson:"basePrice,omitempty" json:"base_price,omitempty"`
	EstimatedTime    EstimatedTime      `bson:"estimatedTime" json:"estimated_time"`
	Distance         string             `bson:"distance" json:"distance"`
	Message          string             `bson:"message,omitempty" json:"message,omitempty"`
	Status           string             `bson:"status" json:"status"`
	ExpiresAt        time.Time          `bson:"expiresAt" json:"expires_at"`
	AcceptedAt       *time.Time         `bson:"acceptedAt,omitempty" json:"accepted_at,omitempty"`
	RejectedAt       *time.Time         `bson:"rejectedAt,omitempty" json:"rejected_at,omitempty"`
	ViewedByUser     bool               `bson:"viewedByUser" json:"viewed_by_user"`
	ViewedAt         *time.Time         `bson:"viewedAt,omitempty" json:"viewed_at,omitempty"`
	CreatedAt        time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updated_at"`
}

type EstimatedTime struct {
	Value int    `bson:"value" json:"value"`
	Unit  string `bson:"unit" json:"unit"`
}

const (
	BidStatusPending   = "pending"
	BidStatusAccepted  = "accepted"
	BidStatusRejected  = "rejected"
	BidStatusExpired   = "expired"
	BidStatusWithdrawn = "withdrawn"
)

const (
	TimeUnitMinutes = "minutes"
	TimeUnitHours   = "hours"
	TimeUnitDays    = "days"
)
