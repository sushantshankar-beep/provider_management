package domain

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type ProviderLocation struct {
	Lat       float64    `bson:"lat" json:"lat"`
	Lon       float64    `bson:"lon" json:"lon"`
	UpdatedAt *time.Time `bson:"updatedAt,omitempty" json:"updated_at,omitempty"`
}

type ServiceLocation struct {
	Lat float64 `bson:"lat" json:"lat"`
	Lon float64 `bson:"lon" json:"lon"`
}

type PayoutStatus string

const (
	PayoutStatusRegular                  PayoutStatus = "regular_payout"
	PayoutStatusRegularComplaint         PayoutStatus = "regular_complaint"
	PayoutStatusCancelled                PayoutStatus = "payout_cancelled"
	PayoutStatusComplaintAfterSettlement PayoutStatus = "complaint_after_settlement"
)

type AcceptedService struct {
	ID                      primitive.ObjectID  `bson:"_id,omitempty"`
	ServiceRequest       primitive.ObjectID   `bson:"serviceRequest" json:"serviceRequest"`
	ServiceNumber     	 string                `bson:"serviceNumber" json:"serviceNumber"`
	NumericID            int64                `bson:"id" json:"numericId"`
	User                 primitive.ObjectID   `bson:"user" json:"user"`
	NotToSendProviders   []primitive.ObjectID `bson:"notToSendProviders,omitempty" json:"notToSendProviders,omitempty"`
	Provider             primitive.ObjectID   `bson:"provider" json:"provider"`
	AcceptedBid          primitive.ObjectID   `bson:"acceptedBid" json:"acceptedBid"`
	ServiceRequestNo        int64               `bson:"serviceRequestId,omitempty" json:"service_request_no,omitempty"`
	Status                ServiceStatus      `bson:"status" json:"status"`

	Timestamps *ServiceTimestamps `bson:"timestamps,omitempty"`
	CancelledBy             string              `bson:"cancelledBy,omitempty" json:"cancelled_by,omitempty"`
	BasePrice               float64             `bson:"basePrice" json:"base_price"`
	FinalPrice              float64             `bson:"finalPrice" json:"final_price"`
	PaymentStatus           string              `bson:"paymentStatus" json:"payment_status"`
	SettlementID            *primitive.ObjectID `bson:"settlementId,omitempty" json:"settlement_id,omitempty"`
	SettlementStatus        SettlementStatus    `bson:"settlementStatus,omitempty" json:"settlement_status,omitempty"`
	SettledAt               *time.Time          `bson:"settledAt,omitempty" json:"settled_at,omitempty"`
	TotalSettledAmount      float64             `bson:"totalSettledAmount" json:"total_settled_amount"`
	RemainingAmount         float64             `bson:"remainingAmount" json:"remaining_amount"`
	ComplaintUserID         string              `bson:"complaintUser,omitempty" json:"complaint_user_id,omitempty"`
	ComplaintProviderID     string              `bson:"complaintProvider,omitempty" json:"complaint_provider_id,omitempty"`
	OrderID                 string              `bson:"orderId,omitempty" json:"order_id,omitempty"`
	ServiceType             string              `bson:"serviceType,omitempty" json:"service_type,omitempty"`
	Issues                  []string            `bson:"issues,omitempty" json:"issues,omitempty"`
	ProviderLocation        ProviderLocation    `bson:"providerLocation,omitempty" json:"provider_location,omitempty"`
	ServiceLocation         ServiceLocation     `bson:"serviceLocation,omitempty" json:"service_location,omitempty"`
	Notes                   []BookingNote       `bson:"notes,omitempty" json:"notes,omitempty"`
	PayoutCreated           bool                `bson:"payoutCreated,omitempty" json:"payoutCreated,omitempty"`
	PayoutCreatedAt         time.Time           `bson:"payoutCreatedAt,omitempty" json:"payoutCreatedAt,omitempty"`
	IsPayoutCancelled       bool                `bson:"isPayoutCancelled" json:"isPayoutCancelled"`
	PayoutCancelledAt       *time.Time          `bson:"payoutCancelledAt" json:"payoutCancelledAt",omitempty"`
	PayoutStatus            PayoutStatus        `bson:"payoutStatus" json:"payoutStatus"`
	IsSettledAfterComplaint bool                `bson:"isSettledAfterComplaint" json:"is_settled_after_complaint"`
	SettledAfterComplaintAt *time.Time          `bson:"settledAfterComplaintAt,omitempty" json:"settled_after_complaint_at,omitempty"`
	HasComplaintAdjustment  bool                `bson:"hasComplaintAdjustment" json:"has_complaint_adjustment"`
	PendingDeductionAmount  float64             `bson:"pendingDeductionAmount" json:"pending_deduction_amount"`
	ComplaintResolvedAt     *time.Time          `bson:"complaintResolvedAt,omitempty"`
	CreatedAt               time.Time           `bson:"createdAt" json:"created_at"`
	UpdatedAt               time.Time           `bson:"updatedAt" json:"updated_at"`
	FuelType   string    `bson:"fuelType" json:"fuelType"`
	VehicleType string    `bson:"vehicleType" json:"vehicleType"`
	VehicleNumber string   `bson:"vehicleNumber" json:"vehicleNumber"`
	Brand       string    	`bson:"brand" json:"brand"`
	ModelYear   int        	`bson:"modelYear" json:"modelYear"`
	Model      string      `bson:"model" json:"model"`
	CancelledByProvider   bool                 `bson:"cancelledByProvider" json:"cancelledByProvider"`
	CancelledProviderID   string               `bson:"cancelledProviderID" json:"cancelledProviderID"`
}

type BookingNote struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	AddedBy   string    `bson:"addedBy" json:"added_by"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}


type ServiceTimestamps struct {
    CreatedAt    *time.Time `bson:"createdAt,omitempty" json:"created_at,omitempty"`
    StartedAt    *time.Time `bson:"startedAt,omitempty" json:"started_at,omitempty"`
    ReachedAt    *time.Time `bson:"reachedAt,omitempty" json:"reached_at,omitempty"`
    InProgressAt *time.Time `bson:"inProgressAt,omitempty" json:"in_progress,omitempty"`
    CompletedAt  *time.Time `bson:"completedAt,omitempty" json:"completed,omitempty"`
    CancelledAt  *time.Time `bson:"cancelledAt,omitempty" json:"cancelled,omitempty"`
    OtpVerified  *time.Time `bson:"OtpVerified,omitempty" json:"otp_verified,omitempty"`
}

type ServiceStatus string

const (
    StatusCreated   ServiceStatus = "created"
    StatusAssigned  ServiceStatus = "assigned"
    StatusStarted   ServiceStatus = "started"
    StatusCompleted ServiceStatus = "completed"
    StatusCancelled ServiceStatus = "cancelled"
	StatusSearching        ServiceStatus = "searching"
	StatusProviderAssigned ServiceStatus = "provider_assigned"
	StatusNotStarted      ServiceStatus = "not_started"
	StatusReachedLocation ServiceStatus = "reached_location"
	StatusOTPVerified     ServiceStatus = "otp_verified"

	StatusInProgress ServiceStatus = "in_progress"
	StatusConfirmed ServiceStatus = "confirmed"

)
