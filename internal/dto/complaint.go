package dto

import (
	"time"
	"provider_management/internal/domain"
)

type ComplaintPagination struct {
	Page  int
	Limit int
}

type ComplaintFilters struct {
	Status        *string
	RaisedBy      *string
	Category      *string
	SearchQuery   *string
	UserID        *string
	ProviderID    *string
	CreatedAtFrom *time.Time
	CreatedAtTo   *time.Time
}

type ComplaintFilter struct {
	Status        *string
	RaisedBy      *string
	Category      *string
	DateFrom      *time.Time
	DateTo        *time.Time
	SearchQuery   *string
	UserID        *string
	ProviderID    *string
	Page          int
	Limit         int
	CreatedAtFrom *time.Time
	CreatedAtTo   *time.Time
}

type RefundRequest struct {
	UserID              string
	BookingID           string
	BookingInternalID   int64
	ComplaintID         string
	ComplaintInternalID int64
	Amount              float64
	Reason              string
	TxnID               string
}

type PayoutRequest struct {
	ProviderID          string
	Amount              float64
	Reason              string
	ComplaintID         string
	BookingID           string
	PartialAmount       float64
	ComplaintInternalID int64
	CancelPayout        bool
	CreateDeduction     bool
}

type ComplaintListItem struct {
	ID                string                 `json:"id"`
	InternalID        string                 `json:"complaint_id"`
	AcceptedServiceNo string                 `json:"booking_no"`
	RaisedBy          string                 `json:"raised_by"`
	Problem           string                 `json:"category"`
	Status            domain.ComplaintStatus `json:"status"`
	CreatedAt         string                 `json:"created_at"`
}

type ComplaintDetailResponse struct {
	ID               string                        `json:"_id"`
	ComplaintID      string                        `json:"complaint_id"`
	BookingID        string                        `json:"booking_id"`
	BookingNo        string                        `json:"booking_no"`
	UserID           string                        `json:"user_id"`
	ProviderID       string                        `json:"provider_id"`
	RaisedBy         string                        `json:"raised_by"`
	Problem          string                        `json:"problem"`
	Photos           []string                      `json:"photos"`
	Status           domain.ComplaintStatus        `json:"status"`
	Timeline         domain.ComplaintTimeline      `json:"timeline"`
	Assessment       *domain.ComplaintAssessment   `json:"assessment"`
	Notes            []domain.ComplaintNote        `json:"notes"`
	ActionsTriggered []string                      `json:"actions_triggered"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
	UpdatedByAdmin   string                        `json:"updated_by_admin"`
	AdminUpdatedAt   *time.Time                    `json:"admin_updated_at"`
	PaymentTracking  *domain.PaymentActionTracking `json:"payment_tracking"`
	UserDetails      *domain.UserDetails           `json:"user_details,omitempty"`
	ProviderDetails  *domain.ProviderDetails       `json:"provider_details,omitempty"`
	BookingDetails   *domain.BookingDetails        `json:"booking_details,omitempty"`
}

type ComplaintListResponse struct {
	Data       []ComplaintListItem `json:"data"`
	Stats      *ComplaintStats     `json:"stats"`
	Pagination PaginationMeta      `json:"pagination"`
}

type AssessComplaintRequest struct {
	FaultParty       domain.FaultParty `json:"fault_party" binding:"required"`
	RefundToUser     domain.RefundType `json:"refund_to_user" binding:"required"`
	RefundAmount     float64           `json:"refund_amount,omitempty"`
	PayoutToProvider domain.PayoutType `json:"payout_to_provider" binding:"required"`
	PayoutAmount     float64           `json:"payout_amount,omitempty"`
	Remarks          string            `json:"remarks,omitempty"`
	AssessedBy       string            `json:"assessed_by"`
	TxnID            string            `json:"-"`
}

type UpdateComplaintStatusRequest struct {
	Status domain.ComplaintStatus `json:"status" binding:"required"`
}

type AddNoteRequest struct {
	Content string `json:"content" binding:"required"`
	AddedBy string `json:"addedBy" binding:"required"`
}

type ComplaintStats struct {
	TotalComplaints    int64 `json:"total_complaints"`
	StatusResolved     int64 `json:"status_resolved"`
	StatusUnresolved   int64 `json:"status_unresolved"`
	StatusInitiated    int64 `json:"status_initiated"`
	RaisedByYou        int64 `json:"raised_by_you"`
	RaisedByProviders  int64 `json:"raised_by_providers"`
	UserComplaints     int64 `json:"user_complaints"`
	ProviderComplaints int64 `json:"provider_complaints"`
}
