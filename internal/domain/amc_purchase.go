package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AMCPurchase struct {
	ID                   string             `bson:"_id,omitempty" json:"_id"`
	UserID               primitive.ObjectID `bson:"user" json:"user"`
	PlanID               primitive.ObjectID `bson:"plan" json:"plan"`
	VehicleID            primitive.ObjectID `bson:"vehicle" json:"vehicle"`
	PlanName             string             `bson:"planName" json:"planName"`
	VehicleNumber        string             `bson:"vehicleNumber" json:"vehicleNumber"`
	PlanStatus           string             `bson:"planStatus" json:"planStatus"`
	PaymentStatus        string             `bson:"paymentStatus" json:"paymentStatus"`
	PlanStartDate        time.Time          `bson:"planStartDate" json:"planStartDate"`
	PlanEndDate          time.Time          `bson:"planEndDate" json:"planEndDate"`
	PlanServicesIncluded []string           `bson:"planServicesIncluded" json:"planServicesIncluded"`
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
}
