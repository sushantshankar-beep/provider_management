package dto

import (
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateProviderRequest struct {
	Name              string                 `form:"name" binding:"required"`
	CompanyName       string                 `form:"companyName"`
	Phone             string                 `form:"phone" binding:"required"`
	Email             string                 `form:"email"`
	AlternateContact  string                 `form:"alternateContact"`
	City              string                 `form:"city" binding:"required"`
	Address           string                 `form:"address" binding:"required"`
	PermanentAddress  string                 `form:"permanentAddress"`
	ShopAddress       string                 `form:"shopAddress"`
	VehicleType       []string               `form:"vehicleType" binding:"required"`
	VehicleNumber     string                 `form:"vehicleNumber"`
	ProviderBrands    []string               `form:"providerBrands"`
	ProviderServices  []string               `form:"providerServices" binding:"required"`
	GSTNumber         string                 `form:"gstNumber"`
	Description       string                 `form:"description"`
	ProfileURL        string                 `form:"-"`
	IdentityProofs []domain.Proof `form:"-"`
	AddressProofs     []domain.Proof         `form:"-"`
	CancelCheque      *domain.CancelCheque   `form:"-"`
	BankDetails       *domain.BankDetails    `form:"-"`
	AccountHolderName string                 `form:"accountHolderName"`
	AccountNumber     string                 `form:"accountNumber"`
	IfscCode          string                 `form:"ifscCode"`
	BranchName        string                 `form:"branchName"`
	Upi               string                 `form:"upi"`
}

type UpdateProviderRequest struct {
	Name              string                 `form:"name"`
	CompanyName       string                 `form:"companyName"`
	Phone             string                 `form:"phone"`
	Email             string                 `form:"email"`
	AlternateContact  string                 `form:"alternateContact"`
	City              string                 `form:"city"`
	Address           string                 `form:"address"`
	PermanentAddress  string                 `form:"permanentAddress"`
	ShopAddress       string                 `form:"shopAddress"`
	VehicleType       []string               `form:"vehicleType"`
	VehicleNumber     string                 `form:"vehicleNumber"`
	ProviderBrands    []string               `form:"providerBrands"`
	ProviderServices  []string               `form:"providerServices"`
	GSTNumber         string                 `form:"gstNumber"`
	Description       string                 `form:"description"`
	ProfileURL        string                 `form:"-"`
	IdentityProofs    []domain.Proof         `form:"-"`
	AddressProofs     []domain.Proof         `form:"-"`
	CancelCheque      *domain.CancelCheque   `form:"-"`
	BankDetails       *domain.BankDetails    `form:"-"`
	AccountHolderName string                 `form:"accountHolderName"`
	AccountNumber     string                 `form:"accountNumber"`
	IfscCode          string                 `form:"ifscCode"`
	BranchName        string                 `form:"branchName"`
	Upi               string                 `form:"upi"`
}

type ProviderResponse struct {
	ID                primitive.ObjectID    `json:"id"`
	Name              string                `json:"name"`
	CompanyName       string                `json:"companyName"`
	Phone             string                `json:"phone"`
	Email             string                `json:"email"`
	City              string                `json:"city"`
	Status            string                `json:"status"`
	CreatedBy        primitive.ObjectID     `json:"createdBy"`
	UpdatedBy         string                `json:"updatedBy"`
	CreatedAt         string                `json:"createdAt"`
}

type ZoneStats struct {
	ZoneName               string `json:"zoneName" bson:"zoneName"`
	TotalProviders         int    `json:"totalProviders" bson:"totalProviders"`
	TotalActivationMembers int    `json:"totalActivationMembers" bson:"totalActivationMembers"`
}