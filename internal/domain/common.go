package domain

import "time"

const (
	AccountStatusActive      = "active"
	AccountStatusSuspended   = "suspended"
	AccountStatusBlacklisted = "blacklisted"
	AccountStatusDeactivated = "deactivated"
	StatusActive             = "active"
	StatusPending            = "Pending"
	StatusRejected           = "rejected"
	VerificationPending      = "pending"
	VerificationApproved     = "approved"
	VerificationRejected     = "rejected"
)

type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type OTP struct {
	Code      string    `bson:"code,omitempty" json:"-"`
	ExpiresAt time.Time `bson:"expiresAt,omitempty" json:"-"`
	Verified  bool      `bson:"verified" json:"-"`
}
