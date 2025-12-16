package service

import (
	"context"
	"fmt"

	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AdminBookingService struct {
	repo *repository.AdminBookingRepo
}

func NewAdminBookingService(repo *repository.AdminBookingRepo) *AdminBookingService {
	return &AdminBookingService{repo: repo}
}

type BookingResponse struct {
	ID               string      `json:"id"`
	BookingID        string      `json:"bookingId"`
	UserID           string      `json:"userId"`
	CustomerName     string      `json:"customerName"`
	Phone            string      `json:"phone"`
	Email            string      `json:"email"`
	ProviderName     string      `json:"providerName,omitempty"`
	ProviderPhone    string      `json:"providerPhone,omitempty"`
	ProviderID       string      `json:"providerId,omitempty"`
	VehicleType      string      `json:"vehicleType"`
	ServiceBidType   string      `json:"serviceBidType"`
	ServiceType      string      `json:"serviceType"`
	Problems         []string    `json:"problems"`
	BookingDate      time.Time   `json:"bookingDate"`
	Zone             string      `json:"zone"`
	Status           string      `json:"status"`
	Amount           float64     `json:"amount"`
	PaymentStatus    string      `json:"paymentStatus"`
	UserRating       *RatingInfo `json:"userRating,omitempty"`
	ProviderRating   *RatingInfo `json:"providerRating,omitempty"`
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
	Rating             *RatingsSummary         `json:"rating,omitempty"`
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
	PaymentDetails     *PaymentDetailsInfo     `json:"paymentDetails,omitempty"`
	EstimatedTime      *EstimatedTimeInfo      `json:"estimatedTime,omitempty"`
	Distance           string                  `json:"distance,omitempty"`
	OTP                *OTPInfo                `json:"otp,omitempty"`
	Timeline           *TimelineInfo           `json:"timeline,omitempty"`
	UserRating         *RatingInfo             `json:"userRating,omitempty"`
	ProviderRating     *RatingInfo             `json:"providerRating,omitempty"`
	UserComplaint      *ComplaintInfo          `json:"userComplaint,omitempty"`
	ProviderComplaint  *ComplaintInfo          `json:"providerComplaint,omitempty"`
	TransactionDetails *TransactionDetailsInfo `json:"transactionDetails,omitempty"`
	ProviderEarnings   *ProviderEarningsInfo   `json:"providerEarnings,omitempty"`
	AllBookings        []BookingSummary        `json:"allBookings,omitempty"`
}

type RatingsSummary struct {
	User     int `json:"user"`
	Provider int `json:"provider"`
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
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
}

type TimelineInfo struct {
	CreatedAt     time.Time  `json:"createdAt"`
	ReachedAt     *time.Time `json:"reachedAt,omitempty"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	OTPVerifiedAt *time.Time `json:"otpVerifiedAt,omitempty"`
	JobStartedAt  *time.Time `json:"jobStartedAt,omitempty"`
}

type RatingInfo struct {
	Stars             int        `json:"stars"`
	Review            string     `json:"review"`
	RecommendToFriend bool       `json:"recommendToFriend"`
	CreatedAt         time.Time  `json:"createdAt"`
	RatedBy           string     `json:"ratedBy"`
	RatedTo           string     `json:"ratedTo"`
}

type ComplaintInfo struct {
	ID        string    `json:"id"`
	RaisedBy  string    `json:"raisedBy"`
	Problem   string    `json:"problem"`
	Photos    []string  `json:"photos"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
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
	UserPaid            float64         `json:"userPaid"`
	Commission          CommissionInfo  `json:"commission"`
	GST                 GSTInfo         `json:"gst"`
	NetPayout           float64         `json:"netPayout"`
	PayoutStatus        string          `json:"payoutStatus"`
	PaymentMode         string          `json:"paymentMode"`
	ExpectedTime        string          `json:"expectedTime"`
	ExpectedPayoutDate  time.Time       `json:"expectedPayoutDate"`
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
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
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
		filter["status"] = status
	}
	if paymentStatus := params["paymentStatus"]; paymentStatus != "" {
		filter["paymentStatus"] = paymentStatus
	}
	if serviceType := params["serviceType"]; serviceType != "" {
		filter["serviceType"] = serviceType
	}
	if userID := params["userId"]; userID != "" {
		filter["user"] = userID
	}
	if providerID := params["providerId"]; providerID != "" {
		filter["provider"] = providerID
	}

	if startDate := params["startDate"]; startDate != "" {
		if sd, err := time.Parse(time.RFC3339, startDate); err == nil {
			if filter["createdAt"] == nil {
				filter["createdAt"] = bson.M{}
			}
			filter["createdAt"].(bson.M)["$gte"] = sd
		}
	}
	if endDate := params["endDate"]; endDate != "" {
		if ed, err := time.Parse(time.RFC3339, endDate); err == nil {
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

	bookings := make([]BookingResponse, 0, len(services))
	serviceIDs := make([]string, 0, len(services))
	
	for _, svc := range services {
		serviceIDs = append(serviceIDs, svc.ID)
	}

	ratingsMap := make(map[string][]*domain.Rating)
	if len(serviceIDs) > 0 {
		ratings, _ := s.repo.FindRatingsByServiceIDs(ctx, serviceIDs)
		for i := range ratings {
			ratingsMap[ratings[i].ServiceID] = append(ratingsMap[ratings[i].ServiceID], &ratings[i])
		}
	}
	
	for _, svc := range services {

		var (
			sr            *domain.ServiceRequest
			srInternalID  int64
		)
	
		if sreq, err := s.repo.FindServiceRequestByID(ctx, svc.ServiceRequestID.Hex()); err == nil {
			sr = sreq
			srInternalID = sreq.InternalID
		}
	
		booking := BookingResponse{
			ID:            svc.ID,
			BookingID:     fmt.Sprintf("BK%d", srInternalID), // ✅ SERVICE REQUEST ID
			ProviderID:    svc.ProviderID,
			Status:        s.mapStatus(svc.Status),
			Amount:        svc.FinalPrice,
			PaymentStatus: s.mapPaymentStatus(svc.PaymentStatus),
			ServiceType:   svc.ServiceType,
			BookingDate:   svc.CreatedAt,
		}
	
		if user, err := s.repo.FindUserByID(ctx, svc.UserID); err == nil {
			booking.UserID = fmt.Sprintf("VW%d", user.InternalID)
			booking.CustomerName = user.Name
			booking.Phone = user.Phone
			booking.Email = user.Email
			booking.Zone = user.SelectedCityName
		}
	
		if provider, err := s.repo.FindProviderByID(ctx, svc.ProviderID); err == nil {
			booking.ProviderName = provider.Name
			booking.ProviderPhone = provider.Phone
		}
	
		if sr != nil {
			booking.VehicleType = sr.VehicleType
			booking.ServiceBidType = sr.ServiceBidType
			booking.Problems = sr.Problems
		}
	
		if serviceRatings, ok := ratingsMap[svc.ID]; ok {
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
	pendingFilter["status"] = bson.M{"$in": []string{"pending", "accepted", "reached"}}
	pending, _ := s.repo.CountAcceptedServices(ctx, pendingFilter)

	inProgressFilter := bson.M{}
	for k, v := range baseFilter {
		inProgressFilter[k] = v
	}
	inProgressFilter["status"] = "in_progress"
	inProgress, _ := s.repo.CountAcceptedServices(ctx, inProgressFilter)

	completedFilter := bson.M{}
	for k, v := range baseFilter {
		completedFilter[k] = v
	}
	completedFilter["status"] = "completed"
	completed, _ := s.repo.CountAcceptedServices(ctx, completedFilter)

	cancelledFilter := bson.M{}
	for k, v := range baseFilter {
		cancelledFilter[k] = v
	}
	cancelledFilter["status"] = "cancelled"
	cancelled, _ := s.repo.CountAcceptedServices(ctx, cancelledFilter)

	revenueFilter := bson.M{}
	for k, v := range baseFilter {
		revenueFilter[k] = v
	}
	revenueFilter["status"] = "completed"
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

	srIDStr := strings.TrimPrefix(bookingID, "BK")
	srInternalID, err := strconv.ParseInt(srIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid booking id")
	}

	sr, err := s.repo.FindServiceRequestByInternalID(ctx, srInternalID)
	if err != nil {
		return nil, err
	}

	svc, err := s.repo.FindAcceptedServiceByServiceRequestID(ctx, sr.ID)
	if err != nil {
		return nil, err
	}

	booking := &DetailedBookingResponse{
		ID:            svc.ID,
		BookingID:     fmt.Sprintf("BK%d", sr.InternalID),
		ProviderID:    svc.ProviderID,
		Status:        strings.ToLower(svc.Status),
		Amount:        svc.FinalPrice,
		PaymentStatus: strings.ToLower(svc.PaymentStatus),
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
	}

	if user, err := s.repo.FindUserByID(ctx, svc.UserID); err == nil {
		booking.UserID = fmt.Sprintf("VW%d", user.InternalID)
		booking.CustomerName = user.Name
		booking.Phone = user.Phone
		booking.Email = user.Email
		booking.Zone = user.SelectedCityName
	}

	if provider, err := s.repo.FindProviderByID(ctx, svc.ProviderID); err == nil {
		booking.ProviderName = provider.Name
		booking.ProviderPhone = provider.Phone
		booking.MechanicType = provider.VehicleType
	}

	gstAmount := svc.FinalPrice * 0.18
	subtotal := svc.FinalPrice - gstAmount
	booking.PaymentDetails = &PaymentDetailsInfo{
		ServiceCharge: svc.FinalPrice,
		Discount:      0,
		Subtotal:      subtotal,
		GST:           gstAmount,
		Total:         svc.FinalPrice,
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
		OTPVerifiedAt: svc.OTPVerifiedAt,
		JobStartedAt:  svc.JobStartedAt,
	}

	if svc.OrderID != "" {
		if txn, err := s.repo.FindTransactionByOrderID(ctx, svc.OrderID); err == nil {
			paymentStatus := "success"
			if txn.Status != "success" && txn.Status != "completed" {
				paymentStatus = txn.Status
			}
			booking.TransactionDetails = &TransactionDetailsInfo{
				TransactionID:     txn.TxnID,
				TransactionNumber: txn.InternalID,
				Amount:            txn.Amount,
				Status:            strings.ToLower(svc.PaymentStatus),
				PaymentStatus:     paymentStatus,
				PaymentMethod:     strings.ToLower(txn.PaymentSource),
			}
		}
	}

	ratings, _ := s.repo.FindRatingsByServiceIDs(ctx, []string{svc.ID})
	userRatingStars := 0
	providerRatingStars := 0
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
			userRatingStars = rating.Stars
		} else if rating.RaterType == "provider" {
			booking.ProviderRating = ratingInfo
			providerRatingStars = rating.Stars
		}
	}

	if userRatingStars > 0 || providerRatingStars > 0 {
		booking.Rating = &RatingsSummary{
			User:     userRatingStars,
			Provider: providerRatingStars,
		}
	}

	if svc.ComplaintUserID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintUserID); err == nil {
			booking.UserComplaint = &ComplaintInfo{
				ID:        complaint.ID,
				RaisedBy:  complaint.RaisedBy,
				Problem:   complaint.Problem,
				Photos:    complaint.Photos,
				Status:    complaint.Status,
				CreatedAt: complaint.CreatedAt,
			}
		}
	}

	if svc.ComplaintProviderID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintProviderID); err == nil {
			booking.ProviderComplaint = &ComplaintInfo{
				ID:        complaint.ID,
				RaisedBy:  complaint.RaisedBy,
				Problem:   complaint.Problem,
				Photos:    complaint.Photos,
				Status:    complaint.Status,
				CreatedAt: complaint.CreatedAt,
			}
		}
	}

	commissionPercentage := 20.0
	if provider, err := s.repo.FindProviderByID(ctx, svc.ProviderID); err == nil {
		if provider.CommissionPercentage > 0 {
			commissionPercentage = provider.CommissionPercentage
		}
	}

	commissionAmount := svc.FinalPrice * (commissionPercentage / 100)
	gstPercentage := 18.0
	gstAmountEarnings := svc.FinalPrice * (gstPercentage / 100)
	netPayout := svc.FinalPrice - commissionAmount - gstAmountEarnings

	payoutStatus := "Pending"
	if svc.Status == "completed" {
		payoutStatus = "Completed"
	}

	expectedPayoutDate := svc.CreatedAt.Add(24 * time.Hour)

	booking.ProviderEarnings = &ProviderEarningsInfo{
		UserPaid: svc.FinalPrice,
		Commission: CommissionInfo{
			Percentage: commissionPercentage,
			Amount:     commissionAmount,
		},
		GST: GSTInfo{
			Percentage: gstPercentage,
			Amount:     gstAmountEarnings,
		},
		NetPayout:          netPayout,
		PayoutStatus:       payoutStatus,
		PaymentMode:        "payu",
		ExpectedTime:       "24 Hours",
		ExpectedPayoutDate: expectedPayoutDate,
	}

	userServices, _, _ := s.repo.FindAcceptedServices(ctx, map[string]interface{}{
		"user": svc.UserID,
	}, 0, 100, "-createdAt")
	var allBookings []BookingSummary
	for _, service := range userServices {
		allBookings = append(allBookings, BookingSummary{
			ID:        service.ID,
			Status:    strings.ToLower(service.Status),
			CreatedAt: service.CreatedAt,
		})
	}
	booking.AllBookings = allBookings

	return booking, nil
}



func (s *AdminBookingService) CancelBooking(ctx context.Context, bookingID string) (*DetailedBookingResponse, error) {
	svc, err := s.repo.FindAcceptedServiceByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

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

	_, err = s.repo.UpdateAcceptedService(ctx, bookingID, update)
	if err != nil {
		return nil, err
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) MarkBookingCompleted(ctx context.Context, bookingID string) (*DetailedBookingResponse, error) {
	svc, err := s.repo.FindAcceptedServiceByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

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

	_, err = s.repo.UpdateAcceptedService(ctx, bookingID, update)
	if err != nil {
		return nil, err
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) mapStatus(status string) string {
	statusMap := map[string]string{
		"pending":     "Pending",
		"accepted":    "Accepted",
		"reached":     "Reached",
		"in_progress": "In Progress",
		"completed":   "Completed",
		"cancelled":   "Cancelled",
	}
	if mapped, ok := statusMap[status]; ok {
		return mapped
	}
	return status
}

func (s *AdminBookingService) mapPaymentStatus(status string) string {
	if status == "success" || status == "paid" {
		return "paid"
	}
	return status
}