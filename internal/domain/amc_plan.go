package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AMCPlan struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	PlanName               string             `bson:"planName" json:"planName"`
	PlanSlug               string             `bson:"planSlug" json:"planSlug"`
	PlanVehicleType        string             `bson:"planVehicleType" json:"planVehicleType"`
	PlanCategory           string             `bson:"planCategory" json:"planCategory"`
	PlanDescription        string             `bson:"planDescription" json:"planDescription"`
	PlanDurationInMonth    int                `bson:"planDurationInMonth" json:"planDurationInMonth"`
	PlanStart              int                `bson:"planStart" json:"planStart"`
	PlanBasePrice          float64            `bson:"planBasePrice" json:"planBasePrice"`
	DiscountPercent        float64            `bson:"discountPercent" json:"discountPercent"`
	PlanDiscountAmount     float64            `bson:"planDiscountAmount" json:"planDiscountAmount"`
	PlanPriceAfterDiscount float64            `bson:"planPriceAfterDiscount" json:"planPriceAfterDiscount"`
	PlanGSTPercent         float64            `bson:"planGSTPercent" json:"planGSTPercent"`
	PlanGSTAmount          float64            `bson:"planGSTAmount" json:"planGSTAmount"`
	PlanTotalAmount        float64            `bson:"planTotalAmount" json:"planTotalAmount"`
	AllowedBrandModels     []BrandModel       `bson:"allowedBrandModels,omitempty" json:"allowedBrandModels,omitempty"`
	PlanCity               []string           `bson:"planCity,omitempty" json:"planCity,omitempty"`
	PlanServicesIncluded   []PlanService      `bson:"planServicesIncluded,omitempty" json:"planServicesIncluded,omitempty"`
	PlanTags               string             `bson:"planTags" json:"planTags"`
	PlanFeatures           []string           `bson:"planFeatures,omitempty" json:"planFeatures,omitempty"`
	Sorting                int                `bson:"sorting" json:"sorting"`
	IsActive               bool               `bson:"isActive" json:"isActive"`
	CreatedBy              primitive.ObjectID `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	LastModifiedBy         primitive.ObjectID `bson:"lastModifiedBy,omitempty" json:"lastModifiedBy,omitempty"`
	CreatedAt              time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type BrandModel struct {
	Brand          string   `bson:"brand" json:"brand"`
	Models         []string `bson:"models" json:"models"`
	AllowAllModels bool     `bson:"allowAllModels" json:"allowAllModels"`
}

type PlanService struct {
	ServiceName string      `bson:"serviceName" json:"serviceName"`
	ServiceType string      `bson:"serviceType" json:"serviceType"`
	Value       interface{} `bson:"value,omitempty" json:"value,omitempty"`
	Info        string      `bson:"info,omitempty" json:"info,omitempty"`
}
