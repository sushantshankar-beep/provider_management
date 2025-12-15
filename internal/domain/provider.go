package domain

import (
	"time"
)

const (
	IdentityProofTypeAadhaar        = "aadhaar"
	IdentityProofTypePan            = "pan"
	IdentityProofTypeVoterID        = "voter_id"
	AddressProofTypeElectricityBill = "electricity_bill"
	AddressProofTypeRentAgreement   = "rent_agreement"
	AddressProofTypeVoterID         = "voter_id"
)

const (
	VehicleTypeCar  = "car"
	VehicleTypeBike = "bike"
)

type Proof struct {
	Type     string `bson:"type" json:"type"`
	File     string `bson:"file" json:"file"`
	Verified string `bson:"verified" json:"verified"`
	ID       string `bson:"_id,omitempty" json:"document_id,omitempty"`
}

type CancelCheque struct {
	File     string `bson:"file" json:"file"`
	Verified string `bson:"verified" json:"verified"`
}

type BankDetails struct {
	AccountHolderName string `bson:"accountHolderName" json:"account_holder_name"`
	AccountNumber     string `bson:"accountNumber" json:"account_number"`
	IfscCode          string `bson:"ifscCode" json:"ifsc_code"`
	BranchName        string `bson:"branchName" json:"branch_name"`
	Upi               string `bson:"upi" json:"upi"`
}

type Provider struct {
	ID                   string        `bson:"_id,omitempty" json:"id"`
	InternalID           int64         `bson:"id" json:"-"`
	Name                 string        `bson:"name" json:"name"`
	Slug                 string        `bson:"slug,omitempty" json:"slug,omitempty"`
	AppStateStatus       string        `bson:"appStateStatus" json:"app_state_status"`
	Phone                string        `bson:"phone" json:"phone"`
	Email                string        `bson:"email,omitempty" json:"email,omitempty"`
	AlternateContact     string        `bson:"alternateContact,omitempty" json:"alternate_contact,omitempty"`
	ProfileURL           string        `bson:"profileUrl,omitempty" json:"profile_url,omitempty"`
	OTP                  OTP           `bson:"otp,omitempty" json:"-"`
	Location             GeoPoint      `bson:"location,omitempty" json:"location,omitempty"`
	Address              string        `bson:"address,omitempty" json:"address,omitempty"`
	PermanentAddress     string        `bson:"permanentAddress,omitempty" json:"permanent_address,omitempty"`
	Status               string        `bson:"status" json:"status"`
	GSTNumber            string        `bson:"GSTNumber,omitempty" json:"gst_number,omitempty"`
	IdentityProof        []Proof       `bson:"identityProof,omitempty" json:"identity_proof,omitempty"`
	AddressProof         []Proof       `bson:"addressProof,omitempty" json:"address_proof,omitempty"`
	CancelCheque         *CancelCheque `bson:"cancelCheque,omitempty" json:"cancel_cheque,omitempty"`
	BankDetails          *BankDetails  `bson:"bankDetails,omitempty" json:"bank_details,omitempty"`
	VehicleNumber        string        `bson:"vehicleNumber,omitempty" json:"vehicle_number,omitempty"`
	FormSubmitted        int           `bson:"formSubbmitted" json:"form_submitted"`
	IsAssigned           bool          `bson:"isAssigned" json:"is_assigned"`
	Description          string        `bson:"description,omitempty" json:"description,omitempty"`
	VehicleType          []string      `bson:"vehicleType,omitempty" json:"vehicle_type,omitempty"`
	ProviderBrands       []string      `bson:"providerBrands,omitempty" json:"provider_brands,omitempty"`
	ProviderServices     []string      `bson:"providerServices,omitempty" json:"provider_services,omitempty"`
	CompanyName          string        `bson:"companyName,omitempty" json:"company_name,omitempty"`
	City                 string        `bson:"city,omitempty" json:"city,omitempty"`
	PreferredLanguage    string        `bson:"preferredLanguage" json:"preferred_language"`
	TermsAndConditions   bool          `bson:"termsAndConditions,omitempty" json:"terms_and_conditions,omitempty"`
	FCMToken             string        `bson:"fcmToken,omitempty" json:"-"`
	IsSocketConnected    bool          `bson:"isSocketConnected" json:"is_socket_connected"`
	IsServiceOn          bool          `bson:"isServiceOn" json:"is_service_on"`
	Tokens               []string      `bson:"tokens,omitempty" json:"-"`
	IsActive             string        `bson:"isActive" json:"is_active"`
	CommissionPercentage float64       `bson:"commissionPercentage,omitempty" json:"commission_percentage,omitempty"`
	CreatedAt            time.Time     `bson:"createdAt" json:"created_at"`
	UpdatedAt            time.Time     `bson:"updatedAt" json:"updated_at"`
}

type Service struct {
    ID            string  `bson:"_id,omitempty" json:"id"`
    ServiceRequestID string `bson:"serviceRequest,omitempty" json:"service_request_id"`
    Problem       string  `bson:"-" json:"problem"`
    VehicleNumber string  `bson:"-" json:"vehicle_number"`
    Amount        float64 `bson:"-" json:"amount"`
    Status        string  `bson:"status" json:"status"`
    Date          string  `bson:"-" json:"date"`
    ServiceType   string  `bson:"serviceType,omitempty" json:"service_type"`
}