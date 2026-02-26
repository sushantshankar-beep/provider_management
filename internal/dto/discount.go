package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateDiscountRequest struct {
	Code                   string   `json:"code" binding:"required"`
	Name                   string   `json:"name" binding:"required"`
	Description            string   `json:"description"`
	Type                   string   `json:"type" binding:"required"`
	Value                  float64  `json:"value" binding:"required"`
	MaxDiscount            float64  `json:"max_discount"`
	Scope                  string   `json:"scope" binding:"required"`
	ApplicableOn           []string `json:"applicable_on" binding:"required"`
	Zones                  []string `json:"zones"`
	PaymentMethods         []string `json:"payment_methods"`
	UserEligibility        string   `json:"user_eligibility"`
	AllowStackingWithPromo bool     `json:"allow_stacking_with_promo"`
	Status                 string   `json:"status"`
	StartAt                string   `json:"start_at"`
	EndAt                  string   `json:"end_at"`
	CreatedBy              string   `json:"created_by"`
}

type UpdateDiscountRequest struct {
	Code                   string   `json:"code"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Type                   string   `json:"type"`
	Value                  float64  `json:"value"`
	MaxDiscount            float64  `json:"max_discount"`
	Scope                  string   `json:"scope"`
	ApplicableOn           []string `json:"applicable_on"`
	Zones                  []string `json:"zones"`
	PaymentMethods         []string `json:"payment_methods"`
	UserEligibility        string   `json:"user_eligibility"`
	AllowStackingWithPromo *bool    `json:"allow_stacking_with_promo"`
	StartAt                string   `json:"start_at"`
	EndAt                  string   `json:"end_at"`
}

type UpdateDiscountStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type DiscountResponse struct {
	ID                     string   `json:"id"`
	Code                   string   `json:"code" `
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Type                   string   `json:"type"`
	Value                  float64  `json:"value"`
	Discount               string   `json:"discount"`
	MaxDiscount            float64  `json:"max_discount"`
	Scope                  string   `json:"scope"`
	ApplicableOn           []string `json:"applicable_on"`
	Zones                  []string `json:"zones"`
	PaymentMethods         []string `json:"payment_methods"`
	UserEligibility        string   `json:"user_eligibility"`
	AllowStackingWithPromo bool     `json:"allow_stacking_with_promo"`
	Status                 string   `json:"status"`
	TotalSavings           float64  `json:"total_savings"`
	TotalOrders            int      `json:"total_orders"`
	CreatedBy              string   `json:"created_by"`
	StartAt                string   `json:"start_at"`
	EndAt                  *string  `json:"end_at"`
	CreatedAt              string   `json:"created_at"`
	UpdatedAt              string   `json:"updated_at"`
}

type DiscountListResponse struct {
	ID            string   `json:"id"`
	Code          string   `json:"code" `
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	Value         float64  `json:"value"`
	Discount      string   `json:"discount"`
	Scope         string   `json:"scope"`
	ApplicableOn  []string `json:"applicable_on"`
	ValidityStart string   `json:"validity_start"`
	ValidityEnd   *string  `json:"validity_end"`
	Status        string   `json:"status"`
	TotalSavings  float64  `json:"total_savings"`
	TotalOrders   int      `json:"total_orders"`
	CreatedBy     string   `json:"created_by"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type DiscountStatsResponse struct {
	AllDiscounts       int64 `json:"all_discounts"`
	ActiveDiscounts    int64 `json:"active_discounts"`
	ScheduledDiscounts int64 `json:"scheduled_discounts"`
	ExpiredDiscounts   int64 `json:"expired_discounts"`
	Drafts             int64 `json:"drafts"`
	PausedDiscounts    int64 `json:"paused_discounts"`
}

func FormatEndAtPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

type DiscountUsageListItem struct {
	DiscountID         string  `json:"discount_id"`
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	TotalRedemptions   int64     `json:"total_redemptions"`
	UniqueUsers        int64   `json:"unique_users"`
	TotalDiscountGiven float64 `json:"total_discount_given"`
	CreatedAt          string  `json:"created_at"`
}

type DiscountUsageSummary struct {
	TotalRedemptions int64   `json:"total_redemptions"`
	TotalDiscount    float64 `json:"total_discount"`
}

type DiscountUserUsageItem struct {
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	UserPhone     string  `json:"user_phone"`
	UsageCount    int64   `json:"usage_count"`
	TotalDiscount float64 `json:"total_discount"`
	LastUsedAt    string  `json:"last_used_at"`
}

type DiscountServiceUsageItem struct {
	ServiceID      string  `json:"service_id"`
	ServiceNumber  string  `json:"service_number"`
	DiscountCode   string  `json:"discount_code"`
	DiscountAmount float64 `json:"discount_amount"`
	TotalDiscount  float64 `json:"total_discount"`
	AmountPaid     float64 `json:"amount_paid"`
	ServiceType    string  `json:"service_type"`
	CreatedAt      string  `json:"created_at"`
}

type DiscountUserUsageRow struct {
	UserID        primitive.ObjectID `bson:"_id"`
	UsageCount    int64              `bson:"usage_count"`
	TotalDiscount float64            `bson:"total_discount"`
	LastUsedAt    time.Time          `bson:"last_used_at"`
}
