package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VehicleBrand struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	VehicleType string             `bson:"vehicleType" json:"vehicleType"`
	BrandName   string             `bson:"brandName" json:"brandName"`
	ModelName   []string           `bson:"modelName" json:"modelName"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}