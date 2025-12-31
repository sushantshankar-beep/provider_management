package service

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"
    "math"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminBookingService struct {
	repo *repository.AdminBookingRepo
}

func NewAdminBookingService(repo *repository.AdminBookingRepo) *AdminBookingService {
	return &AdminBookingService{repo: repo}

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
	Problems       []string    `json:"problems"`
	BookingDate    time.Time   `json:"bookingDate"`
	Zone           string      `json:"zone"`
	Status         string      `json:"status"`
	Amount         float64     `json:"amount"`
	PaymentStatus  string      `json:"paymentStatus"`
	UserRating     *RatingInfo `json:"userRating,omitempty"`
	ProviderRating *RatingInfo `json:"providerRating,omitempty"`
}

type BookingStats struct {
	TotalBookings      int64   `json:"totalBookings"`
	InProgressBookings int64   `json:"inProgressBookings"`
	PendingBookings    int64   `json:"pendingBookings"`
	CompletedBookings  int64   `json:"completedBookings"`
	CancelledBookings  int64   `json:"cancelledBookings"`
	TotalRevenue       float64 `json:"totalRevenue"`
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
	Status             string                  `json:"status"`
	Amount             float64                 `json:"amount"`
	PaymentStatus      string                  `json:"paymentStatus"`
	PaymentDetails     *PaymentDetailsInfo     `json:"paymentDetails"`
	EstimatedTime      *EstimatedTimeInfo      `json:"estimatedTime,omitempty"`
	Distance           string                  `json:"distance,omitempty"`
	OTP                *OTPInfo                `json:"otp"`
	Timeline           *TimelineInfo           `json:"timeline"`
	UserRating         *RatingInfo             `json:"userRating"`
	ProviderRating     *RatingInfo             `json:"providerRating"`
	UserComplaint      *ComplaintInfo          `json:"userComplaint"`
	ProviderComplaint  *ComplaintInfo          `json:"providerComplaint"`
	TransactionDetails *TransactionDetailsInfo `json:"transactionDetails"`
	ProviderEarnings   *ProviderEarningsInfo   `json:"providerEarnings"`
	Notes              []domain.BookingNote      `bson:"notes,omitempty" json:"notes,omitempty"`
	AllBookings        []BookingSummary        `json:"allBookings"`
	IsSettled           bool                `json:"is_settled"`
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
}

type EstimatedTimeInfo struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type OTPInfo struct {
	Verified   bool       `json:"verified"`
	Code       string     `json:"code"`
	VerifiedAt *time.Time `json:"verifiedAt"`
}

type TimelineInfo struct {
	CreatedAt     time.Time  `json:"createdAt"`
	ReachedAt     *time.Time `json:"reachedAt"`
	StartedAt     *time.Time `json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt"`
	CancelledAt   *time.Time `json:"cancelledAt,omitempty"`
	OTPVerifiedAt *time.Time `json:"otpVerifiedAt"`
	JobStartedAt  *time.Time `json:"jobStartedAt"`
}

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
	Problem     string    `json:"problem"`
	Photos      []string  `json:"photos"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TransactionDetailsInfo struct {
	TransactionID     string  `json:"transactionId"`
	TransactionNumber int64   `json:"transactionNumber"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
	PaymentStatus     string  `json:"paymentStatus"`
	PaymentMethod     string  `json:"paymentMethod"`
}

type ProviderEarningsInfo struct {
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
	Status      string    `json:"status"`
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
	Status        string   `json:"status"`
	PaymentStatus string   `json:"paymentStatus"`
}

type InvoicePricing struct {
	ServiceCharge string `json:"serviceCharge"`
	Discount      string `json:"discount"`
	Subtotal      string `json:"subtotal"`
	GST           string `json:"gst"`
	Total         string `json:"total"`
}

const (
	StatusNotStarted  = "not_started"
	StatusStarted     = "started"
	StatusReached     = "reached_location"
	StatusOTPVerified = "otp_verified"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
)

func mapStatusLabelToDB(label string) string {
	switch label {
	case "Pending":
		return StatusNotStarted
	case "Job Started":
		return StatusStarted
	case "Reached Location":
		return StatusReached
	case "OTP Verified":
		return StatusOTPVerified
	case "Service Started":
		return StatusInProgress
	case "Completed":
		return StatusCompleted
	case "Cancelled":
		return StatusCancelled
	default:
		return label
	}
}

func (s *AdminBookingService) GetAllBookings(ctx context.Context, params map[string]string) (*GetAllBookingsResponse, error) {
	page, _ := strconv.Atoi(params["page"])
	limit, _ := strconv.Atoi(params["limit"])
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	filter := bson.M{}

	if status := params["status"]; status != "" {
		if status == "InProgress" || status == "in_progress_group" {
			filter["status"] = bson.M{
				"$in": []string{
					StatusStarted,
					StatusReached,
					StatusOTPVerified,
					StatusInProgress,
				},
			}
		} else {
			filter["status"] = mapStatusLabelToDB(status)
		}
	}


	if paymentStatus := params["paymentStatus"]; paymentStatus != "" {
		filter["paymentStatus"] = paymentStatus
	}

	if serviceType := params["serviceType"]; serviceType != "" {
		filter["serviceType"] = serviceType
	}
	
	if userID := params["userId"]; userID != "" {
		if primitive.IsValidObjectID(userID) {
			objID, _ := primitive.ObjectIDFromHex(userID)
			filter["user"] = objID
		} else {
			cleanID := strings.ReplaceAll(userID, "VW", "")
			if internalID, err := strconv.ParseInt(cleanID, 10, 64); err == nil {
				if user, err := s.repo.FindUserByInternalID(ctx, internalID); err == nil {
					objID, _ := primitive.ObjectIDFromHex(user.ID)
					filter["user"] = objID
				} else {
					return &GetAllBookingsResponse{
						Bookings:    []BookingResponse{},
						TotalPages:  0,
						CurrentPage: page,
						Total:       0,
						Stats:       BookingStats{},
					}, nil
				}
			}
		}
	}

	if providerID := params["providerId"]; providerID != "" {
		var providerObjectID primitive.ObjectID
		if primitive.IsValidObjectID(providerID) {
			providerObjectID, _ = primitive.ObjectIDFromHex(providerID)
			filter["provider"] = providerObjectID
		} else {
			if provider, err := s.repo.FindProviderByProviderID(ctx, providerID); err == nil && provider != nil {
				providerObjectID = provider.ID
				filter["provider"] = providerObjectID
			} else {
				if internalID, err := strconv.ParseInt(providerID, 10, 64); err == nil {
					if provider, err := s.repo.FindProviderByInternalID(ctx, internalID); err == nil && provider != nil {
						providerObjectID = provider.ID
						filter["provider"] = providerObjectID
					}
				}
			}
		}
	}

	if bookingID := params["bookingId"]; bookingID != "" {
		cleanID := strings.ReplaceAll(bookingID, "BK", "")
		if numericID, err := strconv.ParseInt(cleanID, 10, 64); err == nil {
			filter["id"] = numericID
		}
	}

	if search := params["search"]; search != "" {
		searchConditions := []bson.M{}
		cleanSearch := strings.ReplaceAll(search, "BK", "")

		if bookingID, err := strconv.ParseInt(cleanSearch, 10, 64); err == nil {
			searchConditions = append(searchConditions, bson.M{"id": bookingID})
		}

		cleanSearch = strings.ReplaceAll(search, "VW", "")
		if userID, err := strconv.ParseInt(cleanSearch, 10, 64); err == nil {
			if users, err := s.repo.FindUsersByInternalID(ctx, userID); err == nil && len(users) > 0 {
				userIDs := make([]primitive.ObjectID, 0, len(users))
				for _, user := range users {
					objID, _ := primitive.ObjectIDFromHex(user.ID)
					userIDs = append(userIDs, objID)
				}
				searchConditions = append(searchConditions, bson.M{"user": bson.M{"$in": userIDs}})
			}
		}

		if users, err := s.repo.FindUsersBySearch(ctx, search); err == nil && len(users) > 0 {
			userIDs := make([]primitive.ObjectID, 0, len(users))
			for _, user := range users {
				objID, _ := primitive.ObjectIDFromHex(user.ID)
				userIDs = append(userIDs, objID)
			}
			searchConditions = append(searchConditions, bson.M{"user": bson.M{"$in": userIDs}})
		}

		if providers, err := s.repo.FindProvidersBySearch(ctx, search); err == nil && len(providers) > 0 {
			providerIDs := make([]primitive.ObjectID, 0, len(providers))
			for _, provider := range providers {
				providerIDs = append(providerIDs, provider.ID)
			}
			searchConditions = append(searchConditions, bson.M{"provider": bson.M{"$in": providerIDs}})
		}

		if len(searchConditions) > 0 {
			filter["$or"] = searchConditions
		} else {
			return &GetAllBookingsResponse{
				Bookings:    []BookingResponse{},
				TotalPages:  0,
				CurrentPage: page,
				Total:       0,
				Stats:       BookingStats{},
			}, nil
		}
	}

	if startDate := params["startDate"]; startDate != "" {
		if sd, err := time.Parse(time.RFC3339, startDate); err == nil {
			sd = time.Date(sd.Year(), sd.Month(), sd.Day(), 0, 0, 0, 0, sd.Location())
			if filter["createdAt"] == nil {
				filter["createdAt"] = bson.M{}
			}
			filter["createdAt"].(bson.M)["$gte"] = sd
		}
	}
	if endDate := params["endDate"]; endDate != "" {
		if ed, err := time.Parse(time.RFC3339, endDate); err == nil {
			ed = time.Date(ed.Year(), ed.Month(), ed.Day(), 23, 59, 59, 999, ed.Location())
			if filter["createdAt"] == nil {
				filter["createdAt"] = bson.M{}
			}
			filter["createdAt"].(bson.M)["$lte"] = ed
		}
	}

	skip := int64((page - 1) * limit)
	sort := params["sort"]

	services, total, err := s.repo.FindAcceptedServices(ctx, filter, skip, int64(limit), sort)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		stats, _ := s.GetBookingStats(ctx, params)
		return &GetAllBookingsResponse{
			Bookings:    []BookingResponse{},
			TotalPages:  0,
			CurrentPage: page,
			Total:       total,
			Stats:       *stats,
		}, nil
	}

	userIDSet := make(map[string]bool)
	providerIDSet := make(map[string]bool)
	serviceRequestIDSet := make(map[string]bool)
	serviceIDs := make([]string, 0, len(services))

	for _, svc := range services {
		userIDSet[svc.UserID] = true
		providerIDSet[svc.ProviderID.Hex()] = true
		serviceRequestIDSet[svc.ServiceRequestID.Hex()] = true
		serviceIDs = append(serviceIDs, svc.ID.Hex())
	}

	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	providerIDs := make([]string, 0, len(providerIDSet))
	for id := range providerIDSet {
		providerIDs = append(providerIDs, id)
	}

	serviceRequestIDs := make([]string, 0, len(serviceRequestIDSet))
	for id := range serviceRequestIDSet {
		serviceRequestIDs = append(serviceRequestIDs, id)
	}

	userMap := make(map[string]*domain.User)
	users, _ := s.repo.FindUsersByIDs(ctx, userIDs)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	providerMap := make(map[string]*domain.Provider)
	providers, _ := s.repo.FindProvidersByIDs(ctx, providerIDs)
	for i := range providers {
		providerMap[providers[i].ID.Hex()] = &providers[i]
	}

	serviceRequestMap := make(map[string]*domain.ServiceRequest)
	serviceRequests, _ := s.repo.FindServiceRequestsByIDs(ctx, serviceRequestIDs)
	for i := range serviceRequests {
		serviceRequestMap[serviceRequests[i].ID.Hex()] = &serviceRequests[i]
	}

	ratingsMap := make(map[string][]*domain.Rating)
	ratings, _ := s.repo.FindRatingsByServiceIDs(ctx, serviceIDs)
	for i := range ratings {
		ratingsMap[ratings[i].ServiceID] = append(ratingsMap[ratings[i].ServiceID], &ratings[i])
	}

	bookings := make([]BookingResponse, 0, len(services))

	for _, svc := range services {
		booking := BookingResponse{
			ID:            svc.ID.Hex(),
			BookingID:     fmt.Sprintf("BK%d", svc.InternalID),
			ProviderID:    svc.ProviderID.Hex(),
			Status:        s.mapStatus(svc.Status),
			Amount:        svc.FinalPrice,
			PaymentStatus: s.mapPaymentStatus(svc.PaymentStatus),
			ServiceType:   svc.ServiceType,
			BookingDate:   svc.CreatedAt,
		}

		if user, ok := userMap[svc.UserID]; ok {
			booking.UserID = fmt.Sprintf("VW%d", user.InternalID)
			booking.CustomerName = user.Name
			booking.Phone = user.Phone
			booking.Email = user.Email
			booking.Zone = user.SelectedCityName
		}

		if provider, ok := providerMap[svc.ProviderID.Hex()]; ok {
			booking.ProviderName = provider.Name
			booking.ProviderPhone = provider.Phone
		}

		if sr, ok := serviceRequestMap[svc.ServiceRequestID.Hex()]; ok {
			booking.VehicleType = sr.VehicleType
			booking.ServiceBidType = sr.ServiceBidType
			booking.Problems = sr.Problems
		}

		if serviceRatings, ok := ratingsMap[svc.ID.Hex()]; ok {
			for _, rating := range serviceRatings {
				ratingInfo := &RatingInfo{
					Stars:             rating.Stars,
					Review:            rating.Review,
					RecommendToFriend: rating.RecommendToFriend,
				}
				if rating.RaterType == "user" {
					booking.UserRating = ratingInfo
				} else if rating.RaterType == "provider" {
					booking.ProviderRating = ratingInfo
				}
			}
		}

		bookings = append(bookings, booking)
	}

	stats, _ := s.GetBookingStats(ctx, params)

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &GetAllBookingsResponse{
		Bookings:    bookings,
		TotalPages:  totalPages,
		CurrentPage: page,
		Total:       total,
		Stats:       *stats,
	}, nil
}

func (s *AdminBookingService) GetBookingStats(ctx context.Context, params map[string]string) (*BookingStats, error) {
	baseFilter := bson.M{}

	if startDate := params["startDate"]; startDate != "" {
		if sd, err := time.Parse(time.RFC3339, startDate); err == nil {
			if baseFilter["createdAt"] == nil {
				baseFilter["createdAt"] = bson.M{}
			}
			baseFilter["createdAt"].(bson.M)["$gte"] = sd
		}
	}
	if endDate := params["endDate"]; endDate != "" {
		if ed, err := time.Parse(time.RFC3339, endDate); err == nil {
			if baseFilter["createdAt"] == nil {
				baseFilter["createdAt"] = bson.M{}
			}
			baseFilter["createdAt"].(bson.M)["$lte"] = ed
		}
	}

	total, _ := s.repo.CountAcceptedServices(ctx, baseFilter)


	pendingFilter := bson.M{}
	for k, v := range baseFilter {
		pendingFilter[k] = v
	}
	pendingFilter["status"] = StatusNotStarted
	pending, _ := s.repo.CountAcceptedServices(ctx, pendingFilter)
 
   inProgressFilter := bson.M{}
	for k, v := range baseFilter {
		inProgressFilter[k] = v
	}
	inProgressFilter["status"] = bson.M{
		"$in": []string{
			StatusStarted,
			StatusReached,
			StatusOTPVerified,
			StatusInProgress,
		},
	}
	inProgress, _ := s.repo.CountAcceptedServices(ctx, inProgressFilter)

	completedFilter := bson.M{}
	for k, v := range baseFilter {
		completedFilter[k] = v
	}
	completedFilter["status"] = StatusCompleted
	completed, _ := s.repo.CountAcceptedServices(ctx, completedFilter)

	cancelledFilter := bson.M{}
	for k, v := range baseFilter {
		cancelledFilter[k] = v
	}
	cancelledFilter["status"] = StatusCancelled
	cancelled, _ := s.repo.CountAcceptedServices(ctx, cancelledFilter)

	revenueFilter := bson.M{}
	for k, v := range baseFilter {
		revenueFilter[k] = v
	}
	revenueFilter["status"] = StatusCompleted
	revenueFilter["paymentStatus"] = "success"
	revenue, _ := s.repo.AggregateRevenue(ctx, revenueFilter)

	return &BookingStats{
		TotalBookings:      total,
		InProgressBookings: inProgress,
		PendingBookings:    pending,
		CompletedBookings:  completed,
		CancelledBookings:  cancelled,
		TotalRevenue:       revenue,
	}, nil
}

func (s *AdminBookingService) GetBookingByID(
	ctx context.Context,
	bookingID string,
) (*DetailedBookingResponse, error) {

	svcIDStr := strings.TrimPrefix(bookingID, "BK")
	svcInternalID, err := strconv.ParseInt(svcIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid booking id")
	}

	filter := bson.M{"id": svcInternalID}

	allServices, _, err := s.repo.FindAcceptedServices(ctx, filter, 0, 100, "-createdAt")
	if err != nil || len(allServices) == 0 {
		return nil, fmt.Errorf("booking not found")
	}

	svc := allServices[0]

	sr, err := s.repo.FindServiceRequestByID(ctx, svc.ServiceRequestID.Hex())
	if err != nil {
		sr = &domain.ServiceRequest{}
	}

	status := s.mapStatus(svc.Status)

	booking := &DetailedBookingResponse{
		ID:            svc.ID.Hex(),
		BookingID:     fmt.Sprintf("BK%d", svc.InternalID),
		ProviderID:    svc.ProviderID.Hex(),
		Status:        status,
		Amount:        svc.FinalPrice,
		PaymentStatus: svc.PaymentStatus,
		ServiceType:   svc.ServiceType,
		BookingDate:   svc.CreatedAt,
		VehicleType:   sr.VehicleType,
		VehicleNumber: sr.VehicleNumber,
		Brand:         sr.Brand,
		Model:         sr.Model,
		Year:          sr.Year,
		FuelType:      sr.FuelType,
		Problems:      sr.Problems,
		Description:   sr.Description,
		Location:      sr.Address,
		IsSettled:     svc.IsSettled,
		Notes: svc.Notes,
	}

	addressParts := strings.Split(sr.Address, ",")
	if len(addressParts) >= 2 {
		zone := strings.TrimSpace(addressParts[len(addressParts)-2])
		zone = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return -1
			}
			return r
		}, zone)
		booking.Zone = strings.TrimSpace(zone)
	}
	if booking.Zone == "" {
		booking.Zone = "N/A"
	}

	if user, err := s.repo.FindUserByID(ctx, svc.UserID); err == nil {
		booking.UserID = fmt.Sprintf("VW%06d", user.InternalID)
		booking.CustomerName = user.Name
		booking.Phone = user.Phone
		booking.Email = user.Email
	}

	if provider, err := s.repo.FindProviderByID(ctx, svc.ProviderID.Hex()); err == nil {
		booking.ProviderName = provider.Name
		booking.ProviderPhone = provider.Phone
		booking.MechanicType = provider.VehicleType
	}

	basePrice := svc.BasePrice
	if basePrice == 0 {
		basePrice = svc.FinalPrice
	}
	finalPrice := svc.FinalPrice
	gstAmount := finalPrice * 0.18
	subtotal := finalPrice - gstAmount
	discount := 0.0
	if basePrice > finalPrice {
		discount = basePrice - finalPrice
	}
	


	booking.PaymentDetails = &PaymentDetailsInfo{
		ServiceCharge: basePrice,
		Discount:      discount,
		Subtotal:      round2(subtotal),
		GST:           gstAmount,
		Total:         finalPrice,
	}

	booking.OTP = &OTPInfo{
		Verified:   svc.OTP.Verified,
		Code:       svc.OTP.Code,
		VerifiedAt: svc.OTP.VerifiedAt,
	}

	booking.Timeline = &TimelineInfo{
		CreatedAt:     svc.CreatedAt,
		ReachedAt:     svc.ReachedAt,
		StartedAt:     svc.StartedAt,
		CompletedAt:   svc.CompletedAt,
		CancelledAt:   svc.CancelledAt,
		OTPVerifiedAt: svc.OTPVerifiedAt,
		JobStartedAt:  svc.JobStartedAt,
	}

	var transaction *domain.Transaction
	if svc.OrderID != "" {
		if txn, err := s.repo.FindTransactionByOrderID(ctx, svc.OrderID); err == nil {
			transaction = txn
			paymentStatus := "failure"
			if txn.TxnResponse != nil {
				if respMap, ok := txn.TxnResponse.(map[string]interface{}); ok {
					if txnStatus, ok := respMap["status"].(string); ok {
						paymentStatus = txnStatus
					}
				}
			}

			booking.TransactionDetails = &TransactionDetailsInfo{
				TransactionID:     txn.TxnID,
				TransactionNumber: txn.InternalID,
				Amount:            txn.Amount,
				Status:            txn.Status,
				PaymentStatus:     paymentStatus,
				PaymentMethod:     txn.PaymentSource,
			}
		}
	}

	ratings, _ := s.repo.FindRatingsByServiceIDs(ctx, []string{svc.ID.Hex()})
	var userRatings []int
	var providerRatings []int

	for _, rating := range ratings {
		ratingInfo := &RatingInfo{
			Stars:             rating.Stars,
			Review:            rating.Review,
			RecommendToFriend: rating.RecommendToFriend,
			CreatedAt:         rating.CreatedAt,
			RatedBy:           rating.RatedBy,
			RatedTo:           rating.RatedTo,
		}
		if rating.RaterType == "user" {
			booking.UserRating = ratingInfo
			userRatings = append(userRatings, rating.Stars)
		} else if rating.RaterType == "provider" {
			booking.ProviderRating = ratingInfo
			providerRatings = append(providerRatings, rating.Stars)
		}
	}

	userAvg := 0.0
	if len(userRatings) > 0 {
		sum := 0
		for _, r := range userRatings {
			sum += r
		}
		userAvg = float64(sum) / float64(len(userRatings))
		userAvg = float64(int(userAvg*10+0.5)) / 10
	}

	providerAvg := 0.0
	if len(providerRatings) > 0 {
		sum := 0
		for _, r := range providerRatings {
			sum += r
		}
		providerAvg = float64(sum) / float64(len(providerRatings))
		providerAvg = float64(int(providerAvg*10+0.5)) / 10
	}

	booking.Rating = &RatingsSummary{
		User:     userAvg,
		Provider: providerAvg,
	}

	if svc.ComplaintUserID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintUserID); err == nil {
			booking.UserComplaint = &ComplaintInfo{
				ID:          complaint.ID,
				ComplaintID: fmt.Sprintf("CMPL%06d", complaint.InternalID),
				RaisedBy:    complaint.RaisedBy,
				Problem:     complaint.Problem,
				Photos:      complaint.Photos,
				Status:      complaint.Status,
				CreatedAt:   complaint.CreatedAt,
			}
		}
	}

	if svc.ComplaintProviderID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintProviderID); err == nil {
			booking.ProviderComplaint = &ComplaintInfo{
				ID:          complaint.ID,
				ComplaintID: fmt.Sprintf("CMPL%06d", complaint.InternalID),
				RaisedBy:    complaint.RaisedBy,
				Problem:     complaint.Problem,
				Photos:      complaint.Photos,
				Status:      complaint.Status,
				CreatedAt:   complaint.CreatedAt,
			}
		}
	}

	if finalPrice > 0 && transaction != nil && transaction.Status == "paid" {
		commissionAmount := finalPrice * 0.2
		netPayout := finalPrice * 0.62

		payoutStatus := "Pending"
		if transaction.Status == "paid" {
			payoutStatus = "Completed"
		} else if transaction.Status == "refunded" {
			payoutStatus = "Refunded"
		} else if transaction.Status == "failed" {
			payoutStatus = "Failed"
		}

		var expectedPayoutDate *time.Time
		if transaction.Status == "paid" {
			date := transaction.CreatedAt.Add(24 * time.Hour)
			expectedPayoutDate = &date
		}

		booking.ProviderEarnings = &ProviderEarningsInfo{
			UserPaid: finalPrice,
			Commission: CommissionInfo{
				Percentage: 20,
				Amount:     commissionAmount,
			},
			GST: GSTInfo{
				Percentage: 18,
				Amount:     gstAmount,
			},
			NetPayout:          netPayout,
			PayoutStatus:       payoutStatus,
			PaymentMode:        transaction.PaymentSource,
			ExpectedTime:       "24 Hours",
			ExpectedPayoutDate: expectedPayoutDate,
		}
	}

	var allBookings []BookingSummary
	for _, service := range allServices {
		allBookings = append(allBookings, BookingSummary{
			ID:          service.ID.Hex(),
			Status:      service.Status,
			CreatedAt:   service.CreatedAt,
			CancelledBy: service.CancelledBy,
		})
	}
	booking.AllBookings = allBookings

	return booking, nil
}

func (s *AdminBookingService) CancelBooking(ctx context.Context, bookingID string) (*DetailedBookingResponse, error) {
	svcIDStr := strings.TrimPrefix(bookingID, "BK")
	svcInternalID, err := strconv.ParseInt(svcIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid booking id")
	}
	filter := bson.M{"id": svcInternalID}
	services, _, err := s.repo.FindAcceptedServices(ctx, filter, 0, 1, "-createdAt")
	if err != nil {
		return nil, fmt.Errorf("error finding booking: %v", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("booking not found")
	}

	svc := services[0]

	if svc.Status == "cancelled" {
		return nil, fmt.Errorf("booking already cancelled")
	}
	if svc.Status == "completed" {
		return nil, fmt.Errorf("cannot cancel completed booking")
	}

	now := time.Now()
	update := bson.M{
		"status":      "cancelled",
		"cancelledBy": "admin",
		"cancelledAt": now,
		"updatedAt":   now,
	}

	_, err = s.repo.UpdateAcceptedService(ctx, svc.ID.Hex(), update)
	if err != nil {
		return nil, err
	}

	if svc.ProviderID.Hex() != "" {
		err = s.repo.UpdateProviderIsAssigned(ctx, svc.ProviderID.Hex(), false)
		if err != nil {
			log.Printf("Failed to update provider status: %v", err)
		}
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) MarkBookingCompleted(ctx context.Context, bookingID string) (*DetailedBookingResponse, error) {
	svcIDStr := strings.TrimPrefix(bookingID, "BK")
	svcInternalID, err := strconv.ParseInt(svcIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid booking id")
	}

	filter := bson.M{"id": svcInternalID}
	services, _, err := s.repo.FindAcceptedServices(ctx, filter, 0, 1, "-createdAt")
	if err != nil {
		return nil, fmt.Errorf("error finding booking: %v", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("booking not found")
	}

	svc := services[0]

	if svc.Status == "completed" {
		return nil, fmt.Errorf("booking already completed")
	}
	if svc.Status == "cancelled" {
		return nil, fmt.Errorf("cannot complete cancelled booking")
	}

	now := time.Now()
	update := bson.M{
		"status":      "completed",
		"completedAt": now,
		"updatedAt":   now,
	}

	_, err = s.repo.UpdateAcceptedService(ctx, svc.ID.Hex(), update)
	if err != nil {
		return nil, err
	}

	if svc.ProviderID.Hex() != "" {
		err = s.repo.UpdateProviderIsAssigned(ctx, svc.ProviderID.Hex(), false)
		if err != nil {
			log.Printf("Failed to update provider status: %v", err)
		}
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) mapStatus(status string) string {
	switch status {
	case StatusNotStarted:
		return "Pending"
	case StatusStarted:
		return "Job Started"
	case StatusReached:
		return "Reached Location"
	case StatusOTPVerified:
		return "OTP Verified"
	case StatusInProgress:
		return "Service Started"
	case StatusCompleted:
		return "Completed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return status
	}
}


func (s *AdminBookingService) mapPaymentStatus(status string) string {
	if status == "success" || status == "paid" {
		return "paid"
	}
	return status
}

func (s *AdminBookingService) GetInvoiceData(
	ctx context.Context,
	serviceID string,
) (*InvoiceData, error) {

	if serviceID == "" {
		return nil, fmt.Errorf("service ID is required")
	}

	serviceObjectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid service id")
	}

	filter := bson.M{"_id": serviceObjectID}

	services, _, err := s.repo.FindAcceptedServices(ctx, filter, 0, 1, "-createdAt")
	if err != nil {
		return nil, fmt.Errorf("error finding service: %v", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service not found")
	}

	svc := services[0]

	user, err := s.repo.FindUserByID(ctx, svc.UserID)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %v", err)
	}

	provider, err := s.repo.FindProviderByID(ctx, svc.ProviderID.Hex())
	if err != nil {
		return nil, fmt.Errorf("error finding provider: %v", err)
	}

	sr, err := s.repo.FindServiceRequestByID(ctx, svc.ServiceRequestID.Hex())
	if err != nil {
		return nil, fmt.Errorf("error finding service request: %v", err)
	}

	finalPrice := svc.FinalPrice
	gst := finalPrice * 0.18
	subtotal := finalPrice - gst

	var serviceDate string
	if svc.CompletedAt != nil {
		serviceDate = svc.CompletedAt.Format("2006-01-02")
	} else {
		serviceDate = svc.CreatedAt.Format("2006-01-02")
	}

	var gstNumber string
	if provider.GSTNumber != "" {
		gstNumber = provider.GSTNumber
	}

	invoiceSuffix := svc.ID.Hex()

	if len(invoiceSuffix) >= 8 {
		invoiceSuffix = invoiceSuffix[len(invoiceSuffix)-8:]
	}

	invoice := &InvoiceData{
		InvoiceNumber: fmt.Sprintf("INV-%s", strings.ToUpper(invoiceSuffix)),
		InvoiceDate:   time.Now().Format("2006-01-02"),
		ServiceDate:   serviceDate,

		Provider: InvoiceProvider{
			Name:      provider.Name,
			Phone:     provider.Phone,
			Address:   provider.Address,
			GSTNumber: gstNumber,
		},

		Customer: InvoiceCustomer{
			Name:    user.Name,
			Phone:   user.Phone,
			Address: user.Address,
		},

		Vehicle: InvoiceVehicle{
			Brand:         sr.Brand,
			Model:         sr.Model,
			Year:          sr.Year,
			FuelType:      sr.FuelType,
			VehicleNumber: sr.VehicleNumber,
			VehicleType:   sr.VehicleType,
		},

		Service: InvoiceService{
			Type:          sr.ServiceType,
			Problems:      sr.Problems,
			Status:        s.mapStatus(svc.Status),
			PaymentStatus: s.mapPaymentStatus(svc.PaymentStatus),
		},

		Pricing: InvoicePricing{
			ServiceCharge: fmt.Sprintf("%.2f", finalPrice),
			Discount:      "0.00",
			Subtotal:      fmt.Sprintf("%.2f", subtotal),
			GST:           fmt.Sprintf("%.2f", gst),
			Total:         fmt.Sprintf("%.2f", finalPrice),
		},
	}

	return invoice, nil
}

func getProviderName(provider *domain.Provider) string {
	if provider.CompanyName != "" {
		return provider.CompanyName
	}
	return provider.Name
}


func (s *AdminBookingService) AddNote(
	ctx context.Context,
	bookingID string,
	req AddNoteRequest,
) error {

	if req.Content == "" {
		return fmt.Errorf("note content is required")
	}
	if req.AddedBy == "" {
		return fmt.Errorf("addedBy is required")
	}

	// 🔥 BK123 → 123
	cleanID := strings.TrimPrefix(bookingID, "BK")
	internalID, err := strconv.ParseInt(cleanID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid booking id")
	}

	// Find booking
	svcs, _, err := s.repo.FindAcceptedServices(
		ctx,
		bson.M{"id": internalID},
		0, 1, "",
	)
	if err != nil || len(svcs) == 0 {
		return fmt.Errorf("booking not found")
	}

	note := domain.BookingNote{
		ID:        primitive.NewObjectID().Hex(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	return s.repo.AddBookingNote(ctx, svcs[0].ID, note)
}


func round2(val float64) float64 {
	return math.Round(val*100) / 100
}
