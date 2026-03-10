package dto

import (
	"time"
	 "provider_management/internal/domain"
)

type BookingFilters struct {
	Search         string `form:"search"`
	Status         string `form:"status"`
	PaymentStatus  string `form:"paymentStatus"`
	ServiceType    string `form:"serviceType"`
	VehicleType    string `form:"vehicleType"`
	UserID         string `form:"userId"`
	ProviderID     string `form:"providerId"`
	BookingID      string `form:"bookingId"`
	StartDate      string `form:"startDate"`
	EndDate        string `form:"endDate"`
	CreatedAt      string `form:"createdAt"`
	Sort           string `form:"sort"`
}

type BookingPagination struct {
	Page  int64 `form:"page"`
	Limit int64 `form:"limit"`
	Skip  int64
	Sort  string
}

type BookingStats struct {
	TotalBookings      int64   `json:"totalBookings"`
	InProgressBookings int64   `json:"inProgressBookings"`
	PendingBookings    int64   `json:"pendingBookings"`
	CompletedBookings  int64   `json:"completedBookings"`
	CancelledBookings  int64   `json:"cancelledBookings"`
	TotalRevenue       float64 `json:"totalRevenue"`
}


type BookingResponse struct {
	ID             string      `json:"id"`
	BookingID      string      `json:"bookingId"`
	UserID         string      `json:"userId"`
	CustomerName   string      `json:"customerName"`
	Phone          string      `json:"phone"`
	Email          string      `json:"email"`
	ProviderName   string      `json:"providerName,omitempty"`
	ProviderPhone  string      `json:"providerPhone,omitempty"`
	ProviderID     string      `json:"providerId,omitempty"`
	VehicleType    string      `json:"vehicleType"`
	ServiceBidType string      `json:"serviceBidType"`
	ServiceType    string      `json:"serviceType"`
	BookingDate    *time.Time   `json:"bookingDate"`
	CompletedAt    *time.Time  `json:"completedAt"`
	Zone           string      `json:"zone"`
	Status         domain.ServiceStatus      `json:"status"`
	Amount         float64     `json:"amount"`
	PaymentStatus  string      `json:"paymentStatus"`
	UserRating     *RatingInfo `json:"userRating,omitempty"`
	ProviderRating *RatingInfo `json:"providerRating,omitempty"`
}

type RatingsInfo struct {
	Stars             int  `json:"stars"`
	Review            string `json:"review"`
	RecommendToFriend bool   `json:"recommendToFriend"`
}

type BookingListResponse struct {
	Bookings   []BookingResponse `json:"bookings"`
	Stats      BookingStats      `json:"stats"`
	Pagination BookingPaginationMeta    `json:"pagination"`
}

type BookingPaginationMeta struct {
	CurrentPage int64 `json:"currentPage"`
	TotalPages  int64 `json:"totalPages"`
	Total       int64 `json:"total"`
	Limit       int64 `json:"limit"`
	HasNext     bool  `json:"hasNext"`
	HasPrev     bool  `json:"hasPrev"`
}
type GetAllBookingsResponse struct {
	Bookings    []BookingResponse `json:"bookings"`
	TotalPages  int               `json:"totalPages"`
	CurrentPage int               `json:"currentPage"`
	Total       int64             `json:"total"`
	Stats       BookingStats      `json:"stats"`
}

type DetailedBookingResponse struct {
	ID                 string                  `json:"id"`
	BookingID          string                  `json:"bookingId"`
	UserID             string                  `json:"userId"`
	ProviderID         string                  `json:"providerId"`
	CustomerName       string                  `json:"customerName"`
	Phone              string                  `json:"phone"`
	Email              string                  `json:"email"`
	Location           string                  `json:"location"`
	ProviderName       string                  `json:"providerName"`
	ProviderPhone      string                  `json:"providerPhone"`
	MechanicType       []string                `json:"mechanicType"`
	RatingProvider     string                   `json:"ratingProvider"`
	Rating             *RatingsSummary         `json:"rating"`
	VehicleType        string                  `json:"vehicleType"`
	VehicleNumber      string                  `json:"vehicleNumber"`
	Brand              string                  `json:"brand"`
	Model              string                  `json:"model"`
	Year               int                     `json:"year"`
	FuelType           string                  `json:"fuelType"`
	ServiceType        string                  `json:"serviceType"`
	Problems           []string                `json:"problems"`
	Description        string                  `json:"description"`
	BookingDate        time.Time               `json:"bookingDate"`
	Zone               string                  `json:"zone"`
	Status             domain.ServiceStatus                  `json:"status"`
	Amount             float64                 `json:"amount"`
	PaymentStatus      string                  `json:"paymentStatus"`
	PaymentDetails     *PaymentDetailsInfo     `json:"paymentDetails"`
	EstimatedTime      *EstimatedTimeInfo      `json:"estimatedTime,omitempty"`
	Distance           string                  `json:"distance,omitempty"`
	OTP                string               `json:"otp"`
	Timestamps           domain.ServiceTimestamps        `json:"timestamps"`
	UserRating         *RatingInfo             `json:"userRating"`
	ProviderRating     *RatingInfo             `json:"providerRating"`
	UserComplaint      *ComplaintInfo          `json:"userComplaint"`
	ProviderComplaint  *ComplaintInfo          `json:"providerComplaint"`
	TransactionDetails *TransactionDetailsInfo `json:"transactionDetails"`
	ProviderEarnings   *ProviderEarningsInfo   `json:"providerEarnings"`
	Notes              []domain.BookingNote    `bson:"notes,omitempty" json:"notes,omitempty"`
	AllBookings        []BookingSummary        `json:"allBookings"`
	SettlementStatus   SettlementStatus        `json:"settlement_status"`
	CancelledByProvider   bool                 `json:"cancelledByProvider"`
}


type BookingNote struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
}

type RatingsSummary struct {
	User     float64 `json:"user"`
	Provider float64 `json:"provider"`
}

type PaymentDetailsInfo struct {
	ServiceCharge float64 `json:"serviceCharge"`
	Discount      float64 `json:"discount"`
	Subtotal      float64 `json:"subtotal"`
	GST           float64 `json:"gst"`
	Total         float64 `json:"total"`
	TotalDiscount  float64 `json:"totalDiscount"`
	AmountAfterDiscount float64 `json:"amountAfterDiscount"` 
	AppliedPromo        *AppliedPromoInfo    `json:"appliedPromo,omitempty"`
    AppliedDiscount     *AppliedDiscountInfo `json:"appliedDiscount,omitempty"`
}

type EstimatedTimeInfo struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

// type TimelineInfo struct {
// 	CreatedAt     time.Time  `json:"createdAt"`
// 	ReachedAt     *time.Time `json:"reachedAt"`
// 	StartedAt     *time.Time `json:"startedAt"`
// 	CompletedAt   *time.Time `json:"completedAt"`
// 	CancelledAt   *time.Time `json:"cancelledAt,omitempty"`
// 	OTPVerifiedAt *time.Time `json:"otpVerified"`
// 	JobStartedAt  *time.Time `json:"jobStartedAt"`
// }

type RatingInfo struct {
	Stars             int       `json:"stars"`
	Review            string    `json:"review"`
	RecommendToFriend bool      `json:"recommendToFriend"`
	CreatedAt         time.Time `json:"createdAt"`
	RatedBy           string    `json:"ratedBy"`
	RatedTo           string    `json:"ratedTo"`
}

type ComplaintInfo struct {
	ID          string    `json:"id"`
	ComplaintID string    `json:"complaintId"`
	RaisedBy    string    `json:"raisedBy"`
	Status      domain.ComplaintStatus    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UserComplaint     domain.ComplaintSide `json:"userComplaint,omitempty"`
	ProviderComplaint domain.ComplaintSide `json:"providerComplaint,omitempty"`
}

type TransactionDetailsInfo struct {
	TransactionID     string  `json:"transactionId"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
	// PaymentStatus     string  `json:"paymentStatus"`
	PaymentMethod     string  `json:"method"`
}

type ProviderEarningsInfo struct {
	ServiceAmount      float64        `json:"serviceAmount"`
	UserPaid           float64        `json:"userPaid"`
	Commission         CommissionInfo `json:"commission"`
	GST                GSTInfo        `json:"gst"`
	NetPayout          float64        `json:"netPayout"`
	PayoutStatus       string         `json:"payoutStatus"`
	PaymentMode        string         `json:"paymentMode"`
	ExpectedTime       string         `json:"expectedTime"`
	ExpectedPayoutDate *time.Time     `json:"expectedPayoutDate"`
}

type CommissionInfo struct {
	Percentage float64 `json:"percentage"`
	Amount     float64 `json:"amount"`
}

type GSTInfo struct {
	Percentage float64 `json:"percentage"`
	Amount     float64 `json:"amount"`
}

type BookingSummary struct {
	ID          string    `json:"id"`
	Status      domain.ServiceStatus   `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	CancelledBy string    `json:"cancelledBy,omitempty"`
}

type InvoiceData struct {
	InvoiceNumber string          `json:"invoiceNumber"`
	InvoiceDate   string          `json:"invoiceDate"`
	ServiceDate   string          `json:"serviceDate,omitempty"`
	Provider      InvoiceProvider `json:"provider"`
	Customer      InvoiceCustomer `json:"customer"`
	Vehicle       InvoiceVehicle  `json:"vehicle"`
	Service       InvoiceService  `json:"service"`
	Pricing       InvoicePricing  `json:"pricing"`
}

type InvoiceProvider struct {
	Name      string `json:"name"`
	Address   string `json:"address,omitempty"`
	GSTNumber string `json:"gstNumber,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type InvoiceCustomer struct {
	Name    string `json:"name"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

type InvoiceVehicle struct {
	Brand         string `json:"brand,omitempty"`
	Model         string `json:"model,omitempty"`
	Year          int    `json:"year,omitempty"`
	VehicleType   string `json:"vehicleType,omitempty"`
	VehicleNumber string `json:"vehicleNumber,omitempty"`
	FuelType      string `json:"fuelType,omitempty"`
}

type InvoiceService struct {
	Type          string   `json:"type"`
	Problems      []string `json:"problems"`
	Status        domain.ServiceStatus   `json:"status"`
	PaymentStatus string   `json:"paymentStatus"`
}

type InvoicePricing struct {
	ServiceCharge string `json:"serviceCharge"`
	Discount      string `json:"discount"`
	Subtotal      string `json:"subtotal"`
	GST           string `json:"gst"`
	Total         string `json:"total"`
}

type SettlementStatus string

const (
	SettleStatusPending SettlementStatus = "pending"
	SettleStatusSettled SettlementStatus = "settled"
)

type AppliedPromoInfo struct {
    Code        string  `json:"code"`
    DiscountAmt float64 `json:"discountAmt"`
}

type AppliedDiscountInfo struct {
    Code        string  `json:"code"`
    DiscountAmt float64 `json:"discountAmt"`
}
