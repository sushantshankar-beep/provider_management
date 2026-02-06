package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)


const (
	VehicleTypeCar  = "car"
	VehicleTypeBike = "bike"
)
type Provider struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProviderCode         string             `bson:"providerCode" json:"providerCode"`
	InternalID           int64              `bson:"id" json:"-"`
	Name                 string             `bson:"name" json:"name"`
	CompanyName          string             `bson:"companyName,omitempty" json:"companyName,omitempty"`
	Phone                string             `bson:"phone" json:"phone"`
	Email                string             `bson:"email,omitempty" json:"email,omitempty"`
	AlternateContact     string             `bson:"alternateContact,omitempty" json:"alternateContact,omitempty"`
	ProfileURL           string             `bson:"profileUrl" json:"profileUrl"`
	Address              string             `bson:"address,omitempty" json:"address,omitempty"`
	PermanentAddress     string             `bson:"permanentAddress,omitempty" json:"permanentAddress,omitempty"`
	City                 string             `bson:"city,omitempty" json:"city,omitempty"`
	FCMToken             string             `bson:"fcmToken,omitempty" json:"fcmToken"`
	VehicleNumber        string             `bson:"vehicleNumber,omitempty" json:"vehicleNumber,omitempty"`
	Description          string             `bson:"description,omitempty" json:"description,omitempty"`
	VehicleType          []string           `bson:"vehicleType,omitempty" json:"vehicleType,omitempty"`
	ProviderBrands       []string           `bson:"providerBrands,omitempty" json:"providerBrands,omitempty"`
	ProviderServices     []string           `bson:"providerServices,omitempty" json:"providerServices,omitempty"`
	KYCID                primitive.ObjectID `bson:"kycId,omitempty" json:"kycId,omitempty"`
	FormSubmitted        int                `bson:"formSubmitted" json:"formSubmitted"`
	IsAgreementSubmitted bool               `bson:"isAgreementSubmitted" json:"isAgreementSubmitted"`
	AgreementSubmittedAt *time.Time         `bson:"agreementSubmittedAt,omitempty" json:"agreementSubmittedAt,omitempty"`
	AgreementPDF         string             `bson:"agreementPdf,omitempty" json:"agreementPdf,omitempty"`
	CommissionPercentage float64            `bson:"commissionPercentage,omitempty" json:"commissionPercentage,omitempty"`
	IsActive             string             `bson:"isActive" json:"isActive"`
	Rating               string             `bson:"rating" json:"rating"`
	Notes                []ProviderNote     `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
	CreatedBy            primitive.ObjectID `bson:"createdBy" json:"createdBy"`
AgreementUrl         string             `bson:"agreementUrl,omitempty" json:"agreementUrl,omitempty"`
	//no need
	Slug           string   `bson:"slug,omitempty" json:"slug,omitempty"`
	AppStateStatus string   `bson:"appStateStatus" json:"app_state_status"`
	OTP            OTP      `bson:"otp,omitempty" json:"-"`

	Status         string   `bson:"status" json:"status"`
	IsAssigned     bool     `bson:"isAssigned" json:"is_assigned"`

	PreferredLanguage  string `bson:"preferredLanguage" json:"preferred_language"`
	TermsAndConditions bool   `bson:"termsAndConditions,omitempty" json:"terms_and_conditions,omitempty"`

	IsSocketConnected bool          `bson:"isSocketConnected" json:"is_socket_connected"`
	IsServiceOn       bool          `bson:"isServiceOn" json:"is_service_on"`
	Tokens            []string      `bson:"tokens,omitempty" json:"-"`
}

// no need
type Service struct {
	ID               string  `bson:"_id,omitempty" json:"id"`
	ServiceRequestID string  `bson:"serviceRequest,omitempty" json:"service_request_id"`
	Problem          string  `bson:"-" json:"problem"`
	VehicleNumber    string  `bson:"-" json:"vehicle_number"`
	Amount           float64 `bson:"-" json:"amount"`
	Status          ServiceStatus  `bson:"status" json:"status"`
	Date             string  `bson:"-" json:"date"`
	ServiceType      string  `bson:"serviceType,omitempty" json:"service_type"`
}

type DocumentURLResponse struct {
	DocumentType string `json:"documentType"`
	DocumentName string `json:"documentName"`
	URL          string `json:"url"`
	FileType     string `json:"fileType"`
}
type ProviderNote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Content   string             `bson:"content" json:"content"`
	AddedBy   string             `bson:"addedBy" json:"addedBy"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
