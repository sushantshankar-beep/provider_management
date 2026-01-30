package dto

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/domain"
)

type ProviderPagination struct {
	Page  int64
	Limit int64
	Sort  string
	Skip  int64
}

type ProviderFilters struct {
	Search        string
	Status        string
	Name          string
	Mobile        string
	ProviderID    string
	KYCStatus     string
	AccountStatus string
	VehicleType   string
	Zone          string
	StartDate     string
	Filter        string
}

type CreateProviderRequest struct {
	Name              string               `form:"name"`
	CompanyName       string               `form:"companyName"`
	Phone             string               `form:"phone"`
	Email             string               `form:"email"`
	AlternateContact  string               `form:"alternateContact"`
	City              string               `form:"city"`
	Address           string               `form:"address""`
	PermanentAddress  string               `form:"permanentAddress"`
	ShopAddress       string               `form:"shopAddress"`
	VehicleType       []string             `form:"vehicleType"`
	VehicleNumber     string               `form:"vehicleNumber"`
	ProviderBrands    []string             `form:"providerBrands"`
	ProviderServices  []string             `form:"providerServices"`
	GSTNumber         string               `form:"gstNumber"`
	Description       string               `form:"description"`
	ProfileURL        string               `form:"-"`
	AccountHolderName string               `form:"accountHolderName"`
	AccountNumber     string               `form:"accountNumber"`
	IfscCode          string               `form:"ifscCode"`
	BranchName        string               `form:"branchName"`
	Upi               string               `form:"upiId"`
}

type UpdateProviderRequest struct {
	Name              string               `form:"name"`
	CompanyName       string               `form:"companyName"`
	Phone             string               `form:"phone"`
	Email             string               `form:"email"`
	AlternateContact  string               `form:"alternateContact"`
	City              string               `form:"city"`
	Address           string               `form:"address"`
	PermanentAddress  string               `form:"permanentAddress"`
	ShopAddress       string               `form:"shopAddress"`
	VehicleType       []string             `form:"vehicleType"`
	VehicleNumber     string               `form:"vehicleNumber"`
	ProviderBrands    []string             `form:"providerBrands"`
	ProviderServices  []string             `form:"providerServices"`
	GSTNumber         string               `form:"gstNumber"`
	Description       string               `form:"description"`
	ProfileURL        string               `form:"-"`
	IdentityProofs    []domain.Proof       `form:"-"`
	AddressProofs     []domain.Proof       `form:"-"`
	CancelCheque      *domain.CancelCheque `form:"-"`
	BankDetails       *domain.BankDetails  `form:"-"`
	AccountHolderName string               `form:"accountHolderName"`
	AccountNumber     string               `form:"accountNumber"`
	IfscCode          string               `form:"ifscCode"`
	BranchName        string               `form:"branchName"`
	Upi               string               `form:"upi"`
}

type ProviderListResponse struct {
	Providers  []ProviderAllResponse  `json:"providers"`
	Counts     ProviderCounts         `json:"counts"`
	Pagination ProviderMetaPagination `json:"pagination"`
}

type ProviderAllResponse struct {
	ID            string               `json:"id"`
	ProviderID    string               `json:"provider_id"`
	Name          string               `json:"name"`
	Mobile        string               `json:"mobile"`
	Email         string               `json:"email"`
	KYC           string               `json:"kyc"`
	Account       string               `json:"account"`
	Vehicle       string               `json:"vehicle"`
	Zone          string               `json:"zone"`
	DOJ           string               `json:"doj"`
	ProfileURL    string               `json:"profile_url"`
	IsServiceOn   bool                 `json:"is_service_on"`
	IsActive      string               `json:"is_active"`
	TotalJobs     int64                `json:"total_jobs"`
	CompletedJobs int64                `json:"completed_jobs"`
}

type ProviderCounts struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Inactive   int64 `json:"inactive"`
	PendingKYC int64 `json:"pending_kyc"`
}

type ProviderMetaPagination struct {
	CurrentPage int64 `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
	Total       int64 `json:"total"`
	TotalUsers  int64 `json:"total_users"`
	Limit       int64 `json:"limit"`
	HasNext     bool  `json:"has_next,omitempty"`
	HasPrev     bool  `json:"has_prev,omitempty"`
}

type ProviderDetailResponse struct {
	ID                   string                      `json:"id"`
	ProviderID           string                      `json:"providerId"`
	Name                 string                      `json:"name"`
	Phone                string                      `json:"phone"`
	Email                string                      `json:"email"`
	AlternateContact     string                      `json:"alternateContact"`
	ProfileURL           string                      `json:"profileUrl"`
	Address              string                      `json:"address"`
	PermanentAddress     string                      `json:"permanentAddress"`
	City                 string                      `json:"city"`
	Account              string                      `json:"account"`
	VehicleType          []string                    `json:"vehicleType"`
	VehicleNumber        string                      `json:"vehicleNumber"`
	ProviderBrands       []string                    `json:"providerBrands"`
	ProviderServices     []string                    `json:"providerServices"`
	CompanyName          string                      `json:"companyName"`
	Description          string                      `json:"description"`
	Zone                 string                      `json:"zone"`
	DOJ                  string                      `json:"doj"`
	KYCStatus            string                      `json:"kycStatus"`
	KYCID                primitive.ObjectID           `json:"kycId,omitempty"`
	KYCDocuments         []domain.KYCDocument        `json:"providerDocuments"`
	BankDetails          *domain.ProviderBankDetails `json:"bankDetails"`
	IsActive             string                      `json:"isActive"`
	TotalJobs            int64                       `json:"totalJobs"`
	CompletedJobs        int64                       `json:"completedJobs"`
	CommissionPercentage float64                     `json:"commissionPercentage,omitempty"`
	AgreementSubmittedAt *time.Time                   `json:"agreementSubmittedAt,omitempty"`
	Notes                []domain.ProviderNote       `json:"notes,omitempty"`
	ApprovedAt           string                      `json:"approvedAt"`
	IsAgreementSubmitted bool                        `json:"isAgreementSubmitted"`
}

type DocumentResponse struct {
	DocumentID   string `json:"document_id,omitempty"`
	DocumentType string `json:"document_type"`
	Type         string `json:"type,omitempty"`
	File         string `json:"file"`
	Verified     string `json:"verified"`
}

type ProviderActivationTeamResponse struct {
	ZoneName        string                         `json:"zoneName"`
	TotalProviders  int64                          `json:"totalProviders"`
	TotalActivators int64                          `json:"totalActivators"`
	Team            []ProviderActivationTeamMember `json:"team"`
}

type ProviderActivationTeamMember struct {
	PersonID       primitive.ObjectID `json:"personId" bson:"_id"`
	PersonName     string             `json:"personName" bson:"personName"`
	AssignZone     string             `json:"assignZone" bson:"assignZone"`
	TotalProviders int64              `json:"totalProviders" bson:"totalProviders"`
}
