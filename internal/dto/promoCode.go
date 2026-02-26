package dto

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type CreatePromoCodeRequest struct {
	Code                string   `json:"code" binding:"required"`
	Title               string   `json:"title" binding:"required"`
	Description         string   `json:"description"`
	DiscountType        string   `json:"discount_type" binding:"required"`
	Value               float64  `json:"value" binding:"required"`
	MaxDiscount         float64  `json:"max_discount"`
	MinOrderValue       float64  `json:"min_order_value"`
	PerUserLimit        int      `json:"per_user_limit"`
	GlobalRedemptionCap int      `json:"global_redemption_cap"`
	ServiceTypes        []string `json:"service_types" binding:"required"`
	Zones               []string `json:"zones"`
	PaymentMethods      []string `json:"payment_methods"`
	UserEligibility     string   `json:"user_eligibility"`
	AllowStacking       bool     `json:"allow_stacking"`
	Status              string   `json:"status"`
	StartAt             string   `json:"start_at"`
	EndAt               string   `json:"end_at"`
	CreatedBy           string   `json:"created_by"`
}

type UpdatePromoCodeRequest struct {
	Code                string   `json:"code"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	DiscountType        string   `json:"discount_type"`
	Value               float64  `json:"value"`
	MaxDiscount         float64  `json:"max_discount"`
	MinOrderValue       float64  `json:"min_order_value"`
	PerUserLimit        int      `json:"per_user_limit"`
	GlobalRedemptionCap int      `json:"global_redemption_cap"`
	ServiceTypes        []string `json:"service_types"`
	Zones               []string `json:"zones"`
	PaymentMethods      []string `json:"payment_methods"`
	UserEligibility     string   `json:"user_eligibility"`
	AllowStacking       *bool    `json:"allow_stacking"`
	StartAt             string   `json:"start_at"`
	EndAt               string   `json:"end_at"`
}

type UpdatePromoCodeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type PromoCodeResponse struct {
	ID                  string   `json:"id"`
	Code                string   `json:"code"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	DiscountType        string   `json:"discount_type"`
	Value               float64  `json:"value"`
	Discount            string   `json:"discount"`
	MaxDiscount         float64  `json:"max_discount"`
	MinOrderValue       float64  `json:"min_order_value"`
	PerUserLimit        int      `json:"per_user_limit"`
	GlobalRedemptionCap int      `json:"global_redemption_cap"`
	ServiceTypes        []string `json:"service_types"`
	ServiceType         string   `json:"service_type"`
	Zones               []string `json:"zones"`
	PaymentMethods      []string `json:"payment_methods"`
	UserEligibility     string   `json:"user_eligibility"`
	AllowStacking       bool     `json:"allow_stacking"`
	Status              string   `json:"status"`
	UsageCount          int      `json:"usage_count"`
	Usage               string   `json:"usage"`
	TotalDiscount       float64  `json:"total_discount"`
	CreatedBy           string   `json:"created_by"`
	StartAt             string   `json:"start_at"`
	EndAt               *string  `json:"end_at"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type PromoCodeListResponse struct {
	ID            string   `json:"id"`
	Code          string   `json:"code"`
	Title         string   `json:"title"`
	Discount      string   `json:"discount"`
	ServiceType   string   `json:"service_type"`
	ServiceTypes  []string `json:"service_types"`
	ValidityStart string   `json:"validity_start"`
	ValidityEnd   *string  `json:"validity_end"`
	Status        string   `json:"status"`
	Usage         string   `json:"usage"`
	TotalDiscount string   `json:"total_discount"`
	CreatedBy     string   `json:"created_by"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type PromoCodeStatsResponse struct {
	TotalPromoCodes  int64 `json:"total_promo_codes"`
	ActivePromos     int64 `json:"active_promos"`
	ScheduledPromos  int64 `json:"scheduled_promos"`
	ExpiredPromos    int64 `json:"expired_promos"`
	Drafts           int64 `json:"drafts"`
	InactivePromos   int64 `json:"inactive_promos"`
}

func FormatEndAt(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
type PromoUsageSummary struct {
	ID               string `bson:"_id" json:"id"`
	Code                string             `bson:"code" json:"code"`
	UsageCount      int64   `json:"usageCount" bson:"usageCount"`
	TotalDiscount   float64 `json:"totalDiscount" bson:"totalDiscount"`
	TotalOrderValue float64 `json:"totalOrderValue" bson:"totalOrderValue"`
}

type PromoUsageListItem struct {
	PromoID             string  `json:"promo_id"`
	Code                string  `json:"code"`
	TotalRedemptions    int     `json:"total_redemptions"`   
	UniqueUsers         int64   `json:"unique_users"`         
	TotalDiscountGiven  float64 `json:"total_discount_given"`
	GlobalRedemptionCap int     `json:"global_redemption_cap"`
	CreatedAt           string  `json:"created_at"`
}

type PromoUserUsageRow struct {
	UserID        primitive.ObjectID `bson:"_id"`
	UsageCount    int64              `bson:"usage_count"`
	TotalDiscount float64            `bson:"total_discount"`
	LastUsedAt    time.Time          `bson:"last_used_at"`
}

type PromoServiceUsageItem struct {
	ServiceID     string  `json:"service_id"`
	ServiceNumber string  `json:"service_number"`
	PromoCode     string  `json:"promo_code"`
	PromoAmount   float64 `json:"promo_amount"`
	TotalDiscount float64 `json:"total_discount"`
	AmountPaid    float64 `json:"amount_paid"`
	ServiceType   string  `json:"service_type"`
	CreatedAt     string  `json:"created_at"`
}

type PromoUserUsageItem struct {
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	UserPhone     string  `json:"user_phone"`
	UsageCount    int64   `json:"usage_count"`    
	TotalDiscount float64 `json:"total_discount"`
	LastUsedAt    string  `json:"last_used_at"`
}
