package domain

import (
	"time"
)

// ComplaintStatus represents the status of a complaint
type ComplaintStatus string

const (
	ComplaintStatusPending   ComplaintStatus = "pending"
	ComplaintStatusInReview  ComplaintStatus = "in_review"
	ComplaintStatusResolved  ComplaintStatus = "resolved"
	ComplaintStatusCancelled ComplaintStatus = "cancelled"
)

// FaultParty represents who is at fault
type FaultParty string

const (
	FaultPartyUser     FaultParty = "user"
	FaultPartyProvider FaultParty = "provider"
	FaultPartyBoth     FaultParty = "both"
	FaultPartySystem   FaultParty = "system"
)

// RefundType represents the type of refund
type RefundType string

const (
	RefundTypeFull    RefundType = "full"
	RefundTypePartial RefundType = "partial"
	RefundTypeNone    RefundType = "none"
)

// PayoutType represents the type of payout
type PayoutType string

const (
	PayoutTypeFull    PayoutType = "full"
	PayoutTypePartial PayoutType = "partial"
	PayoutTypeNone    PayoutType = "none"
)

// ComplaintTimeline tracks the status changes
type ComplaintTimeline struct {
	Initiated *time.Time `bson:"initiated,omitempty" json:"initiated,omitempty"`
	InReview  *time.Time `bson:"inReview,omitempty" json:"in_review,omitempty"`
	Resolved  *time.Time `bson:"resolved,omitempty" json:"resolved,omitempty"`
}

// ComplaintAssessment contains the assessment details
type ComplaintAssessment struct {
	FaultParty       FaultParty `bson:"faultParty" json:"fault_party"`
	RefundToUser     RefundType `bson:"refundToUser" json:"refund_to_user"`
	RefundAmount     float64    `bson:"refundAmount,omitempty" json:"refund_amount,omitempty"`
	PayoutToProvider PayoutType `bson:"payoutToProvider" json:"payout_to_provider"`
	PayoutAmount     float64    `bson:"payoutAmount,omitempty" json:"payout_amount,omitempty"`
	Remarks          string     `bson:"remarks,omitempty" json:"remarks,omitempty"`
	AssessedBy       string     `bson:"assessedBy" json:"assessed_by"`
	AssessedAt       time.Time  `bson:"assessedAt" json:"assessed_at"`
}

// ComplaintNote represents a note added to the complaint
type ComplaintNote struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	AddedBy   string    `bson:"addedBy" json:"added_by"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}

// Complaint represents a complaint in the system
type Complaint struct {
	ID                string            `bson:"_id,omitempty" json:"_id"`
	InternalID        int64             `bson:"id" json:"id"`
	AcceptedServiceID string            `bson:"acceptedService" json:"accepted_service_id"`
	AcceptedServiceNo int64             `bson:"acceptedServiceId" json:"accepted_service_no"`
	UserID            string            `bson:"userId,omitempty" json:"user_id,omitempty"`
	ProviderID        string            `bson:"providerId,omitempty" json:"provider_id,omitempty"`
	RaisedBy          string            `bson:"raisedBy" json:"raised_by"` // "user" or "provider"
	Problem           string            `bson:"problem" json:"problem"`
	Photos            []string          `bson:"photos" json:"photos"`
	Status            string            `bson:"status" json:"status"` // "pending", "in_review", "resolved"
	Timeline          ComplaintTimeline `bson:"timeline" json:"timeline"`
	UpdatedByAdmin    string            `bson:"updatedByAdmin,omitempty" json:"updated_by_admin,omitempty"`
	AdminUpdatedAt    *time.Time        `bson:"adminUpdatedAt,omitempty" json:"admin_updated_at,omitempty"`
	CreatedAt         time.Time         `bson:"createdAt" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updatedAt" json:"updated_at"`

	// Additional fields for enhanced functionality
	Category         string               `bson:"category,omitempty" json:"category,omitempty"` // Late Arrival, Payment Issue, etc.
	Assessment       *ComplaintAssessment `bson:"assessment,omitempty" json:"assessment,omitempty"`
	Notes            []ComplaintNote      `bson:"notes,omitempty" json:"notes,omitempty"`
	ActionsTriggered []string             `bson:"actionsTriggered,omitempty" json:"actions_triggered,omitempty"`

	// Cached names for display (optional, can be populated from relations)
	UserName      string `bson:"userName,omitempty" json:"user_name,omitempty"`
	ProviderName  string `bson:"providerName,omitempty" json:"provider_name,omitempty"`
	BookingNumber string `bson:"bookingNumber,omitempty" json:"booking_number,omitempty"`
}

// ComplaintFilter represents filter options for listing complaints
type ComplaintFilter struct {
	Status      *string
	RaisedBy    *string
	Category    *string
	DateFrom    *time.Time
	DateTo      *time.Time
	SearchQuery *string
	Page        int
	Limit       int
}

// ComplaintStats represents complaint statistics
type ComplaintStats struct {
	TotalComplaints    int64 `json:"total_complaints"`
	UserComplaints     int64 `json:"user_complaints"`
	ProviderComplaints int64 `json:"provider_complaints"`
}
