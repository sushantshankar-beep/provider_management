package dto

import (
	"time"
	"provider_management/internal/domain"
)
type UserPagination struct {
	Page  int64
	Limit int64
	Sort  string
	Skip  int64
}

type UserFilters struct {
	Search        string
	Status        string
	Zone         string
    AMCStatus     string
	PlatformUsed   string
	VehicleType     string
	StartDate     string
}

type UserAdminListResponse struct {
	Users      []UserAdminResponse `json:"users"`
	Pagination  UserMetaPagination          `json:"pagination"`
	Stats      *UserStatsResponse  `json:"stats"`
}

type UserStatsResponse struct {
	TotalRegisteredUsers int64 `json:"total_registered_users"`
	TotalActiveUsers     int64 `json:"total_active_users"`
	TotalInactiveUsers   int64 `json:"total_inactive_users"`
	ActiveAMCUsers       int64 `json:"active_amc_users"`
	UsersWithBookings    int64 `json:"users_with_bookings"`
}

type UserAdminResponse struct {
	ID       string `json:"_id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	UserID        string `json:"userId"`
	Zone          string `json:"zone"`
	Status        string `json:"status"`
	CreatedDate   string `json:"createdDate"`
	PlatformUsed  string `json:"platformUsed"`
	VehicleType   string `json:"vehicleType"`
	VehicleCount  int    `json:"vehicleCount"`
	TotalBookings int    `json:"totalBookings"`
	AMCStatus     string `json:"amcStatus"`
	Notes         []domain.UserNote  `json:"notes"`
	ProfileURL    string `json:"profileUrl,omitempty"`
	Address       string `json:"address,omitempty"`
	ISActive   string   		`json:"is_active"`
}

type UserDetailResponse struct {
	UserAdminResponse
	TotalExpenses int64         `json:"totalExpenses"`
	Vehicles      []VehicleInfo `json:"vehicles"`
	AMCInfo       AMCInfo       `json:"amcInfo"`
	PreferredLang string        `json:"preferredLanguage"`
}

type VehicleInfo struct {
	ID            string `json:"_id"`
	VehicleNumber string `json:"vehicleNumber"`
	Brand         string `json:"brand"`
	Model         string `json:"model"`
	Year          string `json:"year"`
	FuelType      string `json:"fuelType"`
	VehicleType   string `json:"vehicleType"`
}

type AMCInfo struct {
	AMCID        string    `json:"amcId"`
	AMCStatus          string    `json:"amcStatus"`
	CurrentPlanName    string    `json:"currentPlanName"`
	StartDate          time.Time `json:"startDate,omitempty"`
	EndDate            time.Time `json:"endDate,omitempty"`
	ActivationDate     time.Time `json:"activationDate,omitempty"`
	BoundVehicleNumber string    `json:"boundVehicleNumber"`
	TotalServices      int       `json:"totalServices"`
	ServicesUsed       int       `json:"servicesUsed"`
	ServicesRemaining  int       `json:"servicesRemaining"`
	PaymentStatus      string    `json:"paymentStatus"`
}

type UserAdminStatusResponse struct {
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UserMetaPagination struct {
	CurrentPage int64 `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
	Total       int64 `json:"total"`
	TotalUsers  int64 `json:"total_users"`
	Limit       int64 `json:"limit"`
	HasNext     bool  `json:"has_next,omitempty"`
	HasPrev     bool  `json:"has_prev,omitempty"`
}

type OfferUsageItem struct {
	ServiceID     string  `json:"service_id"`
	ServiceNumber string  `json:"service_number"`
	UserID        string  `json:"user_id"`

	Promo    *OfferPromoInfo    `json:"promo,omitempty"`
	Discount *OfferDiscountInfo `json:"discount,omitempty"`

	TotalDiscount float64 `json:"total_discount"`
	AmountPaid    float64 `json:"amount_paid"`
	CreatedAt     string  `json:"created_at"`
}

type OfferPromoInfo struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type OfferDiscountInfo struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}