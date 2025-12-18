package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedVehicle struct {
	ID            string             `bson:"_id,omitempty" json:"_id"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`
	VehicleNumber string             `bson:"vehicleNumber" json:"vehicleNumber"`
	Brand         string             `bson:"brand" json:"brand"`
	Model         string             `bson:"model" json:"model"`
	Year          string             `bson:"year" json:"year"`
	FuelType      string             `bson:"fuelType" json:"fuelType"`
	VehicleType   string             `bson:"vehicleType" json:"vehicleType"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}
