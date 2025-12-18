
package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceMaster struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ServiceName     string               `bson:"serviceName" json:"service_name"`
	Category        string               `bson:"category" json:"category"`
	VehicleType     []string             `bson:"vehicleType" json:"vehicle_type"`
	BrandScope      []string             `bson:"brandScope" json:"brand_scope"`
	SubBrandScope   []string             `bson:"subBrandScope" json:"sub_brand_scope"`
	FuelTypeScope   []string             `bson:"fuelTypeScope" json:"fuel_type_scope"`
	Location        string               `bson:"location" json:"location"`
	RequiresOTP     bool                 `bson:"requiresOtp" json:"requires_otp"`
	Status          string               `bson:"status" json:"status"`
	DisplayOrder    int                  `bson:"displayOrder" json:"display_order"`
	Tag             string               `bson:"tag" json:"tag"`
	MinCost         float64              `bson:"minCost" json:"min_cost"`
	ShortDesc       string               `bson:"shortDesc" json:"short_description"`
	CreatedAt       time.Time            `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updatedAt" json:"updated_at"`
}