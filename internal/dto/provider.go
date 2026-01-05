package dto

import (
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateProviderRequest struct {
	Name              string                 `json:"name" binding:"required"`
	CompanyName       string                 `json:"companyName"`
	Phone             string                 `json:"phone" binding:"required"`
	Email             string                 `json:"email"`
	AlternateContact  string                 `json:"alternateContact"`
	City              string                 `json:"city" binding:"required"`
	Address           string                 `json:"address" binding:"required"`
	PermanentAddress  string                 `json:"permanentAddress"`
	ShopAddress       string                 `json:"shopAddress"`
	VehicleType       []string               `json:"vehicleType" binding:"required"`
	VehicleNumber     string                 `json:"vehicleNumber"`
	ProviderBrands    []string               `json:"providerBrands"`
	ProviderServices  []string               `json:"providerServices" binding:"required"`
	GSTNumber         string                 `json:"gstNumber"`
	Description       string                 `json:"description"`
	ProfileURL        string                 `json:"profileUrl"`
	IdentityProofs    []domain.Proof         `json:"identityProofs"`
	AddressProofs     []domain.Proof         `json:"addressProofs"`
	CancelCheque      *domain.CancelCheque   `json:"cancelCheque"`
	BankDetails       *domain.BankDetails    `json:"bankDetails"`
}

type UpdateProviderRequest struct {
	Name              string                 `json:"name"`
	CompanyName       string                 `json:"companyName"`
	Phone             string                 `json:"phone"`
	Email             string                 `json:"email"`
	AlternateContact  string                 `json:"alternateContact"`
	City              string                 `json:"city"`
	Address           string                 `json:"address"`
	PermanentAddress  string                 `json:"permanentAddress"`
	ShopAddress       string                 `json:"shopAddress"`
	VehicleType       []string               `json:"vehicleType"`
	VehicleNumber     string                 `json:"vehicleNumber"`
	ProviderBrands    []string               `json:"providerBrands"`
	ProviderServices  []string               `json:"providerServices"`
	GSTNumber         string                 `json:"gstNumber"`
	Description       string                 `json:"description"`
	ProfileURL        string                 `json:"profileUrl"`
	IdentityProofs    []domain.Proof         `json:"identityProofs"`
	AddressProofs     []domain.Proof         `json:"addressProofs"`
	CancelCheque      *domain.CancelCheque   `json:"cancelCheque"`
	BankDetails       *domain.BankDetails    `json:"bankDetails"`
}

type ProviderResponse struct {
	ID                primitive.ObjectID    `json:"id"`
	Name              string                `json:"name"`
	CompanyName       string                `json:"companyName"`
	Phone             string                `json:"phone"`
	Email             string                `json:"email"`
	City              string                `json:"city"`
	Status            string                `json:"status"`
	CreatedBy         string                `json:"createdBy"`
	UpdatedBy         string                `json:"updatedBy"`
	CreatedAt         string                `json:"createdAt"`
}