package domain

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type ServiceMetadata struct {
	BroadcastedTo int       `bson:"broadcastedTo" json:"broadcasted_to"`
	LastBidAt     time.Time `bson:"lastBidAt,omitempty" json:"last_bid_at,omitempty"`
}

type ServiceRequest struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	InternalID     int64              `bson:"id" json:"id"`
	UserID         string             `bson:"user" json:"user_id"`
	VehicleNumber  string             `bson:"vehicleNumber" json:"vehicle_number"`
	VehicleType    string             `bson:"vehicleType" json:"vehicle_type"`
	Brand          string             `bson:"brand" json:"brand"`
	Model          string             `bson:"model" json:"model"`
	BasePrice      float64            `bson:"basePrice" json:"base_price"`
	FinalPrice     float64            `bson:"finalPrice,omitempty" json:"final_price,omitempty"`
	Year           int                `bson:"year,omitempty" json:"year,omitempty"`
	FuelType       string             `bson:"fuelType" json:"fuel_type"`
	ServiceType    string             `bson:"serviceType" json:"service_type"`
	ServiceBidType string             `bson:"serviceBidType" json:"service_bid_type"`
	Problems       []string           `bson:"problems,omitempty" json:"problems,omitempty"`
	Description    string             `bson:"description,omitempty" json:"description,omitempty"`
	Location       GeoPoint           `bson:"location,omitempty" json:"location,omitempty"`
	Address        string             `bson:"address,omitempty" json:"address,omitempty"`
	ScheduledDate  time.Time          `bson:"scheduledDate,omitempty" json:"scheduled_date,omitempty"`
	Status         string             `bson:"status" json:"status"`
	AcceptedBid    string             `bson:"acceptedBid,omitempty" json:"accepted_bid,omitempty"`
	TotalBids      int                `bson:"totalBids" json:"total_bids"`
	ExpiresAt      time.Time          `bson:"expiresAt" json:"expires_at"`
	Metadata       ServiceMetadata    `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updated_at"`
}
