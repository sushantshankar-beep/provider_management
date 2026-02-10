package dto

import (
	"provider_management/internal/domain"
	"time"
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
	ComplaintInternalID string
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
	ComplaintInternalID string
	CancelPayout        bool
	CreateDeduction     bool
}

type ComplaintListItem struct {
	ID                string                 `json:"id"`
	InternalID        string                 `json:"complaint_id"`
	AcceptedServiceNo string                 `json:"booking_no"`
	RaisedBy          string                 `json:"raised_by"`
	Against           string                 `json:"against"`
	Status            domain.ComplaintStatus `json:"status"`
	UserName     string `json:"userName"`
	ProviderName string `json:"providerName"`
	CreatedAt         string                 `json:"created_at"`
}

type ComplaintListResponse struct {
	Data       []ComplaintListItem `json:"data"`
	Stats      *ComplaintStats     `json:"stats"`
	Pagination PaginationMeta      `json:"pagination"`
}

type AssessComplaintRequest struct {
	FaultParty        domain.FaultParty `json:"fault_party" binding:"required"`
	RefundToUser      domain.RefundType `json:"refund_to_user" binding:"required"`
	RefundAmount      float64           `json:"refund_amount,omitempty"`
	PayoutToProvider  domain.PayoutType `json:"payout_to_provider" binding:"required"`
	PayoutAmount      float64           `json:"payout_amount,omitempty"`
	RemarkForUser     string            `json:"remarkForUser,omitempty"`
	RemarkForProvider string            `json:"remarkForProvider,omitempty"`
	AssessedBy        string            `json:"assessed_by"`
	TxnID             string            `json:"-"`
	GSTAmount       float64 
    TDSAmount       float64 
    NetAmount       float64 
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

type ComplaintDetailResponse struct {
	ComplaintID          string                   `json:"_id"`
	ComplaintNumber      string                   `json:"complaintNumber"`
	Status               domain.ComplaintStatus   `json:"status"`
	ComplaintInformation ComplaintInformation     `json:"complaintInformation"`
	UserToProvider       *ComplaintSideUI         `json:"userToProvider,omitempty"`
	ProviderToUser       *ComplaintSideUI         `json:"providerToUser,omitempty"`
	Tracking             domain.ComplaintTimeline `json:"tracking"`
	Notes                []domain.ComplaintNote   `bson:"notes,omitempty" json:"notes,omitempty"`
	Assessment       *domain.ComplaintAssessment   `json:"assessment"`
	ActionsTriggered []string                      `json:"actions_triggered"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
	UpdatedByAdmin   string                        `json:"updated_by_admin"`
	AdminUpdatedAt   *time.Time                    `json:"admin_updated_at"`
	PaymentTracking  *domain.PaymentActionTracking `json:"payment_tracking"`

}

type ComplaintInformation struct {
	ComplaintID   string    `json:"complaintId"`
	BookingID     string    `json:"bookingId,omitempty"`
	AMCBookingID  string    `json:"amcBookingId,omitempty"`
	BookingAmount float64   `json:"bookingAmount,omitempty"`
	SubmittedAt   time.Time `json:"submittedAt"`
}

type ComplaintSideUI struct {
	SubmittedBy      PartyInfo `json:"submittedBy"`
	ComplaintAgainst PartyInfo `json:"complaintAgainst"`

	Description string    `json:"description"`
	Images      []string  `json:"images"`
	RaisedAt    time.Time `json:"raisedAt"`
}

type PartyInfo struct {
	ID         string `json:"_id"`
	InternalID string `json:"internalID"`
	Name       string `json:"name"`
	Mobile     string `json:"mobile"`
	Role       string `json:"role"`
}
