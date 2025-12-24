package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AllowedBrandModel struct {
	Brand          string   `bson:"brand" json:"brand"`
	Models         []string `bson:"models" json:"models"`
	AllowAllModels bool     `bson:"allowAllModels" json:"allow_all_models"`
}

type PlanServiceIncluded struct {
	ServiceName string      `bson:"serviceName" json:"service_name"`
	ServiceType string      `bson:"serviceType" json:"service_type"`
	Value       interface{} `bson:"value" json:"value"`
	Info        string      `bson:"info" json:"info"`
}

type AMCPlan struct {
	ID                      primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	PlanName                string                `bson:"planName" json:"plan_name"`
	PlanSlug                string                `bson:"planSlug" json:"plan_slug"`
	PlanVehicleType         string                `bson:"planVehicleType" json:"plan_vehicle_type"` 
	PlanCategory            string                `bson:"planCategory" json:"plan_category"` 
	PlanDescription         string                `bson:"planDescription" json:"plan_description"`
	PlanDurationInMonth     int                   `bson:"planDurationInMonth" json:"plan_duration_in_month"`
	PlanStart               int                   `bson:"planStart" json:"plan_start"`
	PlanBasePrice           float64               `bson:"planBasePrice" json:"plan_base_price"`
	DiscountPercent         float64               `bson:"discountPercent" json:"discount_percent"`
	PlanDiscountAmount      float64               `bson:"planDiscountAmount" json:"plan_discount_amount"`
	PlanPriceAfterDiscount  float64               `bson:"planPriceAfterDiscount" json:"plan_price_after_discount"`
	PlanGSTPercent          float64               `bson:"planGSTPercent" json:"plan_gst_percent"`
	PlanGSTAmount           float64               `bson:"planGSTAmount" json:"plan_gst_amount"`
	PlanTotalAmount         float64               `bson:"planTotalAmount" json:"plan_total_amount"`
	AllowedBrandModels      []AllowedBrandModel   `bson:"allowedBrandModels" json:"allowed_brand_models"`
	PlanCity                []string              `bson:"planCity" json:"plan_city"`
	PlanServicesIncluded  []PlanServiceIncluded `bson:"planServicesIncluded" json:"plan_services_included"`
	PlanTags                string                `bson:"planTags" json:"plan_tags"` 
	PlanFeatures            []string              `bson:"planFeatures" json:"plan_features"`
	Sorting                 int                   `bson:"sorting" json:"sorting"`
	PlanStatus              string                `bson:"planStatus,omitempty" json:"plan_status,omitempty"`
	IsActive                bool                  `bson:"isActive" json:"is_active"`
	CreatedBy               primitive.ObjectID    `bson:"createdBy,omitempty" json:"created_by,omitempty"`
	LastModifiedBy          primitive.ObjectID    `bson:"lastModifiedBy,omitempty" json:"last_modified_by,omitempty"`
	CreatedAt               time.Time             `bson:"createdAt" json:"created_at"`
	UpdatedAt               time.Time             `bson:"updatedAt" json:"updated_at"`
}