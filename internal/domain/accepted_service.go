package domain

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type OTPInfo struct {
	Code       string     `bson:"code,omitempty" json:"code,omitempty"`
	Verified   bool       `bson:"verified" json:"verified"`
	VerifiedAt *time.Time `bson:"verifiedAt,omitempty" json:"verified_at,omitempty"`
}

type ProviderLocation struct {
	Lat       float64    `bson:"lat" json:"lat"`
	Lon       float64    `bson:"lon" json:"lon"`
	UpdatedAt *time.Time `bson:"updatedAt,omitempty" json:"updated_at,omitempty"`
}

type ServiceLocation struct {
	Lat float64 `bson:"lat" json:"lat"`
	Lon float64 `bson:"lon" json:"lon"`
}

type AcceptedService struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty"`
	InternalID          int64               `bson:"id" json:"id"`
	ServiceRequestID    primitive.ObjectID  `bson:"serviceRequest" json:"-"`
	ServiceRequestNo    int64               `bson:"serviceRequestId,omitempty" json:"service_request_no,omitempty"`
	UserID              string              `bson:"user" json:"user_id"`
	ProviderID          primitive.ObjectID  `bson:"provider" json:"provider_id"`
	AcceptedBidID       string              `bson:"acceptedBid" json:"accepted_bid_id"`
	NotToSendProviders  []string            `bson:"notToSendProviders,omitempty" json:"not_to_send_providers,omitempty"`
	OTP                 OTPInfo             `bson:"otp,omitempty" json:"otp,omitempty"`
	Status              string              `bson:"status" json:"status"`
	ReachedAt           *time.Time          `bson:"reachedAt,omitempty" json:"reached_at,omitempty"`
	StartedAt           *time.Time          `bson:"startedAt,omitempty" json:"started_at,omitempty"`
	CompletedAt         *time.Time          `bson:"completedAt,omitempty" json:"completed_at,omitempty"`
	CancelledAt         *time.Time          `bson:"cancelledAt,omitempty" json:"cancelled_at,omitempty"`
	ExpiresAt           *time.Time          `bson:"expiresAt,omitempty" json:"expires_at,omitempty"`
	OTPVerifiedAt       *time.Time          `bson:"otpVerifiedAt,omitempty" json:"otp_verified_at,omitempty"`
	JobStartedAt        *time.Time          `bson:"jobStartedAt,omitempty" json:"job_started_at,omitempty"`
	CancelledBy         string              `bson:"cancelledBy,omitempty" json:"cancelled_by,omitempty"`
	BasePrice           float64             `bson:"basePrice" json:"base_price"`
	FinalPrice          float64             `bson:"finalPrice" json:"final_price"`
	PaymentStatus       string              `bson:"paymentStatus" json:"payment_status"`
	IsSettled           bool                `bson:"isSettled" json:"is_settled"`
	SettlementID        *primitive.ObjectID `bson:"settlementId,omitempty" json:"settlement_id,omitempty"`
	SettledAt           *time.Time          `bson:"settledAt,omitempty" json:"settled_at,omitempty"`
	ComplaintUserID     string              `bson:"complaintUser,omitempty" json:"complaint_user_id,omitempty"`
	ComplaintProviderID string              `bson:"complaintProvider,omitempty" json:"complaint_provider_id,omitempty"`
	OrderID             string              `bson:"orderId,omitempty" json:"order_id,omitempty"`
	ServiceType         string              `bson:"serviceType,omitempty" json:"service_type,omitempty"`
	Issues              []string            `bson:"issues,omitempty" json:"issues,omitempty"`
	ProviderLocation    ProviderLocation    `bson:"providerLocation,omitempty" json:"provider_location,omitempty"`
	ServiceLocation     ServiceLocation     `bson:"serviceLocation,omitempty" json:"service_location,omitempty"`
	Notes               []BookingNote      `bson:"notes,omitempty" json:"notes,omitempty"`
	PayoutCreated   bool      `bson:"payoutCreated,omitempty" json:"payoutCreated,omitempty"`
	PayoutCreatedAt time.Time `bson:"payoutCreatedAt,omitempty" json:"payoutCreatedAt,omitempty"`
	CreatedAt           time.Time           `bson:"createdAt" json:"created_at"`
	UpdatedAt           time.Time           `bson:"updatedAt" json:"updated_at"`
}

type BookingNote struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	AddedBy   string    `bson:"addedBy" json:"added_by"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}
