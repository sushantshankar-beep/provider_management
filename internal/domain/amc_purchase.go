package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AMCPurchase struct {
	ID                      primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	InternalID              int64                  `bson:"id" json:"id"`
	UserID                  primitive.ObjectID     `bson:"user" json:"user"`
	PlanID                  primitive.ObjectID     `bson:"plan" json:"plan"`
	SavedVehicleID          primitive.ObjectID     `bson:"savedVehicle,omitempty" json:"savedVehicle,omitempty"`
	Vehicle                 VehicleInfo            `bson:"vehicle" json:"vehicle"`
	PlanName                string                 `bson:"planName,omitempty" json:"planName,omitempty"`
	VehicleNumber           string                 `bson:"vehicleNumber" json:"vehicleNumber"`
	PlanPrice               float64                `bson:"planPrice" json:"planPrice"`
	PlanStartDate           time.Time              `bson:"planStartDate" json:"planStartDate"`
	PlanEndDate             time.Time              `bson:"planEndDate" json:"planEndDate"`
	PlanCityID              primitive.ObjectID     `bson:"planCity,omitempty" json:"planCity,omitempty"`
	PlanStatus              string                 `bson:"planStatus" json:"planStatus"`
	ServiceDetails          []ServiceDetail        `bson:"serviceDetails,omitempty" json:"serviceDetails,omitempty"`
	PaymentID               string                 `bson:"paymentId,omitempty" json:"paymentId,omitempty"`
	PaymentStatus           string                 `bson:"paymentStatus" json:"paymentStatus"`
	PaymentInitiationData   map[string]interface{} `bson:"paymentInitiationData,omitempty" json:"paymentInitiationData,omitempty"`
	PaymentVerificationData map[string]interface{} `bson:"paymentVerificationData,omitempty" json:"paymentVerificationData,omitempty"`
	PaymentCallbackData     map[string]interface{} `bson:"paymentCallbackData,omitempty" json:"paymentCallbackData,omitempty"`
	PayuTransactionID       string                 `bson:"payuTransactionId,omitempty" json:"payuTransactionId,omitempty"`
	PayuResponse            map[string]interface{} `bson:"payuResponse,omitempty" json:"payuResponse,omitempty"`
	PaymentSource           string                 `bson:"paymentSource" json:"paymentSource"`
	Currency                string                 `bson:"currency" json:"currency"`
	RefundStatus            string                 `bson:"refundStatus" json:"refundStatus"`
	RefundRequestID         primitive.ObjectID     `bson:"refundRequestId,omitempty" json:"refundRequestId,omitempty"`
	RefundCancelledByUser   bool                   `bson:"refundCancelledByUser" json:"refundCancelledByUser"`
	VehicleEditableUntil    time.Time              `bson:"vehicleEditableUntil" json:"vehicleEditableUntil"`
	PlanServicesIncluded    []string               `bson:"planServicesIncluded" json:"planServicesIncluded"`
	CreatedAt               time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt               time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type VehicleInfo struct {
	VehicleNumber string `bson:"vehicleNumber" json:"vehicleNumber"`
	Brand         string `bson:"brand" json:"brand"`
	Model         string `bson:"model" json:"model"`
	Year          string `bson:"year,omitempty" json:"year,omitempty"`
	FuelType      string `bson:"fuelType,omitempty" json:"fuelType,omitempty"`
	VehicleType   string `bson:"vehicleType,omitempty" json:"vehicleType,omitempty"`
}

type ServiceDetail struct {
	Service string `bson:"service" json:"service"`
	Count   string `bson:"count" json:"count"`
}
