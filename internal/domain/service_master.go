
package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceCategory string

const (
	BatteryServices    ServiceCategory = "Battery Services"
	TireServices ServiceCategory = "Tire Services"
	MechanicServices   ServiceCategory = "Mechanic Services"
)

type VehicleType string

const (
	VehicleCar   VehicleType = "Car"
	VehicleBike  VehicleType = "Bike"
	VehicleBoth VehicleType = "Both Car & Bike"
)

type ServiceMasterStatus string

const (
	ServiceStatusActive   ServiceMasterStatus = "active"
	ServiceStatusInactive ServiceMasterStatus = "inactive"
)
  
type ServiceMaster struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceName     string             `bson:"serviceName" json:"service_name"`
	Category        ServiceCategory          `bson:"category" json:"category"`
	VehicleType     VehicleType        `bson:"vehicleType,omitempty" json:"vehicle_type,omitempty"`
	BrandScope      []string           `bson:"brandScope,omitempty" json:"brand_scope,omitempty"`
	SubBrandScope   []string           `bson:"subBrandScope,omitempty" json:"sub_brand_scope,omitempty"`
	FuelTypeScope   []string           `bson:"fuelTypeScope,omitempty" json:"fuel_type_scope,omitempty"`
	Location        string             `bson:"location" json:"location"`
	RequiresOTP     bool               `bson:"requiresOtp" json:"requires_otp"`
	Status          ServiceMasterStatus         `bson:"status" json:"status"`
	DisplayOrder    int                `bson:"displayOrder" json:"display_order"`
	Tag             string             `bson:"tag" json:"tag"`
	MinCost         float64            `bson:"minCost" json:"min_cost"`
	ShortDesc       string             `bson:"shortDesc" json:"short_description"`
	CreatedAt       time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updated_at"`
}

func IsValidCategory(c ServiceCategory) bool {
	switch c {
	case BatteryServices  , TireServices, MechanicServices:
		return true
	default:
		return false
	}
}

func IsValidVehicleType(v VehicleType) bool {
	switch v {
	case VehicleCar, VehicleBike, VehicleBoth:
		return true
	default:
		return false
	}
}

func IsValidStatus(s ServiceMasterStatus) bool {
	return s == ServiceStatusActive  || s == ServiceStatusInactive
}