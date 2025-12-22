package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedVehicle struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`
	VehicleNumber string             `bson:"vehicleNumber" json:"vehicleNumber"`
	Brand         string             `bson:"brand" json:"brand"`
	Model         string             `bson:"model" json:"model"`
	Year          string             `bson:"year,omitempty" json:"year,omitempty"`
	FuelType      string             `bson:"fuelType,omitempty" json:"fuelType,omitempty"`
	VehicleType   string             `bson:"vehicleType,omitempty" json:"vehicleType,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
