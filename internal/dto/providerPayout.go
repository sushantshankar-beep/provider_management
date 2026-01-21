package dto

import (
	"time"
    "provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayoutFilters struct {
	Search     string
	ProviderID string
	Status     string
	PeriodFrom string
	PeriodTo   string
	SortBy     string
	SortOrder  string
	PayoutIDInt int64
}

type PayoutSort struct {
	SortBy    string
	SortOrder string
}

type PayoutResponse struct {
	ID                string    `json:"id"`
	PayoutID          string    `json:"payout_id"`
	ProviderID        string    `json:"provider_id"`
	ProviderName      string    `json:"provider_name"`
	ServiceIDs        []primitive.ObjectID   `json:"service_ids"`
	BaseAmount        float64   `json:"base_amount"`
	CommissionPercent float64   `json:"commission_percent"`
	CommissionAmount  float64   `json:"commission_amount"`
	GSTPercent        float64   `json:"gst_percent"`
	GSTAmount         float64   `json:"gst_amount"`
	NetPayable        float64   `json:"net_payable"`
	Status            string    `json:"status"`
	PeriodFrom        time.Time `json:"period_from"`
	PeriodTo          time.Time `json:"period_to"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PayoutListResponse struct {
	Data       []PayoutResponse `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
}

type PayoutServiceResponse struct {
	ID                       string  `json:"id"`
	BookingID                string  `json:"booking_id"`
	AMCID                    string  `json:"amc_id"`
	ProviderID               string  `json:"provider_id"`
	ServiceAmount            float64 `json:"service_amount"`
	CommissionPercent        float64 `json:"commission_percent"`
	CommissionAmount         float64 `json:"commission_amount"`
	GSTPercent               float64 `json:"gst_percent"`
	GSTAmount                float64 `json:"gst_amount"`
	NetAmount                float64 `json:"net_amount"`
	PartialAmount            float64 `json:"partial_amount"`
	PayoutID                 string  `json:"payout_id"`
	SettlementStatus         string  `json:"settlement_status"`
	SettlementID             *primitive.ObjectID   `json:"settlement_id"`
	SettledAt                *time.Time `json:"settled_at"`
	HasComplaintAdjustment   bool    `json:"has_complaint_adjustment"`
	PendingSettlement        float64 `json:"pending_settlement"`
	ShowComplaintAdjustment  bool    `json:"show_complaint_adjustment"`
	PayoutStatus             domain.PayoutStatus  `json:"payout_status"`
	IsPayoutCancelled        bool    `json:"isPayoutCancelled"`
	IsSettledAfterComplaint  bool    `json:"is_settled_after_complaint"`
}

type ProviderPayoutDetailsResponse struct {
	ProviderDetails   map[string]any   `json:"provider_details"`
	AccountDetails    map[string]any   `json:"account_details"`
	EarningSummary    map[string]any   `json:"earning_summary"`
	SettlementHistory []map[string]any `json:"settlement_history"`
}

type PayoutServiceListResponse struct {
	Data []PayoutServiceResponse `json:"data"`
}
