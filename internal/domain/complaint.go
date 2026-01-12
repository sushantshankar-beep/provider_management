package domain

import (
	"time"
)

type ComplaintStatus string

const (
	ComplaintStatusPending   ComplaintStatus = "pending"
	ComplaintStatusInReview  ComplaintStatus = "in_review"
	ComplaintStatusResolved  ComplaintStatus = "resolved"
	ComplaintStatusCancelled ComplaintStatus = "cancelled"
)

type FaultParty string

const (
	FaultPartyUser     FaultParty = "user"
	FaultPartyProvider FaultParty = "provider"
	FaultPartyBoth     FaultParty = "both"
	FaultPartySystem   FaultParty = "system"
)

type RefundType string

const (
	RefundTypeFull    RefundType = "Full Refund"
	RefundTypePartial RefundType = "Partial Refund"
	RefundTypeNone    RefundType = "No Refund"
)

type PayoutType string

const (
	PayoutTypeFull    PayoutType = "Full Payout"
	PayoutTypePartial PayoutType = "Partial Payout"
	PayoutTypeNone    PayoutType = "No Payout"
)

type PaymentActionStatus string

const (
	PaymentActionPending   PaymentActionStatus = "pending"
	PaymentActionCompleted PaymentActionStatus = "completed"
	PaymentActionFailed    PaymentActionStatus = "failed"
	PaymentActionNA        PaymentActionStatus = "n/a"
)

type ComplaintTimeline struct {
	Initiated *time.Time `bson:"initiated,omitempty" json:"initiated,omitempty"`
	InReview  *time.Time `bson:"inReview,omitempty" json:"in_review,omitempty"`
	Resolved  *time.Time `bson:"resolved,omitempty" json:"resolved,omitempty"`
}

type PaymentActionTracking struct {
	RefundStatus PaymentActionStatus `bson:"refundStatus" json:"refund_status"`
	PayoutStatus PaymentActionStatus `bson:"payoutStatus" json:"payout_status"`
	RefundID     string              `bson:"refundId,omitempty" json:"refund_id,omitempty"`
	PayoutID     string              `bson:"payoutId,omitempty" json:"payout_id,omitempty"`
}
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
type ComplaintNote struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	AddedBy   string    `bson:"addedBy" json:"added_by"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}

type Complaint struct {
	ID                string                  `bson:"_id,omitempty" json:"_id"`
	InternalID        int64                   `bson:"id" json:"id"`
	AcceptedServiceID string                  `bson:"acceptedService" json:"accepted_service_id"`
	AcceptedServiceNo int64                   `bson:"acceptedServiceId" json:"accepted_service_no"`
	UserID            string                  `bson:"userId,omitempty" json:"user_id,omitempty"`
	ProviderID        string                  `bson:"providerId,omitempty" json:"provider_id,omitempty"`
	RaisedBy          string                  `bson:"raisedBy" json:"raised_by"`
	Problem           string                  `bson:"problem" json:"problem"`
	Photos            []string                `bson:"photos" json:"photos"`
	Status            string                  `bson:"status" json:"status"`
	Timeline          ComplaintTimeline       `bson:"timeline" json:"timeline"`
	UpdatedByAdmin    string                  `bson:"updatedByAdmin,omitempty" json:"updated_by_admin,omitempty"`
	AdminUpdatedAt    *time.Time              `bson:"adminUpdatedAt,omitempty" json:"admin_updated_at,omitempty"`
	CreatedAt         time.Time               `bson:"createdAt" json:"created_at"`
	UpdatedAt         time.Time               `bson:"updatedAt" json:"updated_at"`
	Category          string                  `bson:"category,omitempty" json:"category,omitempty"`
	Assessment        *ComplaintAssessment    `bson:"assessment,omitempty" json:"assessment,omitempty"`
	Notes             []ComplaintNote         `bson:"notes,omitempty" json:"notes,omitempty"`
	ActionsTriggered  []string                `bson:"actionsTriggered,omitempty" json:"actions_triggered,omitempty"`
	PaymentTracking   *PaymentActionTracking  `bson:"paymentTracking,omitempty" json:"payment_tracking,omitempty"`
	UserName          string                  `bson:"userName,omitempty" json:"user_name,omitempty"`
	ProviderName      string                  `bson:"providerName,omitempty" json:"provider_name,omitempty"`
	BookingNumber     string                  `bson:"bookingNumber,omitempty" json:"booking_number,omitempty"`
	DeductionPayout   *DeductionPayoutRequest `bson:"deductionPayout,omitempty" json:"deduction_payout,omitempty"`
}

type UserDetails struct {
	ID         string `json:"id"`
	InternalID string `json:"internal_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	ProfilePic string `json:"profile_pic,omitempty"`
}

type ProviderDetails struct {
	ID          string `json:"id"`
	InternalID  string `json:"internal_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	ProfilePic  string `json:"profile_pic,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
}

type ComplaintWithDetails struct {
	Complaint
	UserDetails     *UserDetails     `json:"user_details,omitempty"`
	ProviderDetails *ProviderDetails `json:"provider_details,omitempty"`
	BookingDetails  *BookingDetails  `json:"booking_details,omitempty"`
}

type ComplaintFilter struct {
	Status      *string
	RaisedBy    *string
	Category    *string
	DateFrom    *time.Time
	DateTo      *time.Time
	SearchQuery *string
	UserID      *string
	ProviderID  *string
	Page        int
	Limit       int
}

type ComplaintStats struct {
	TotalComplaints    int64 `json:"total_complaints"`
	StatusResolved     int64 `json:"status_resolved"`
	StatusUnresolved   int64 `json:"status_unresolved"`
	StatusInitiated   int64 `json:"status_initiated"`
	RaisedByYou        int64 `json:"raised_by_you"`
	RaisedByProviders  int64 `json:"raised_by_providers"`
	UserComplaints     int64 `json:"user_complaints"`
	ProviderComplaints int64 `json:"provider_complaints"`
}

type BookingDetails struct {
	ID         string  `json:"id"`
	InternalID int64   `json:"internal_id"`
	BasePrice  float64 `json:"base_price"`
	FinalPrice float64 `json:"final_price"`
}

type DeductionPayoutRequest struct {
	ProviderID          string
	BookingID           string
	OriginalAmount      float64
	DeductionAmount     float64
	RemainingAmount     float64
	Reason              string
	ComplaintID         string
	ComplaintInternalID int64
}
