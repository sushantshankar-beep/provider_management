package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
)

type AdminBookingService struct {
	repo *repository.AdminBookingRepo
}

func NewAdminBookingService(repo *repository.AdminBookingRepo) *AdminBookingService {
	return &AdminBookingService{repo: repo}
}

type RatingInfo struct {
	Stars             int    `json:"stars"`
	Review            string `json:"review"`
	RecommendToFriend bool   `json:"recommendToFriend"`
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

type ComplaintInfo struct {
	ID        string    `json:"id"`
	RaisedBy  string    `json:"raisedBy"`
	Problem   string    `json:"problem"`
	Photos    []string  `json:"photos"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type DetailedBookingResponse struct {
	ID                string                 `json:"id"`
	BookingID         string                 `json:"bookingId"`
	UserID            string                 `json:"userId"`
	CustomerName      string                 `json:"customerName"`
	Phone             string                 `json:"phone"`
	Email             string                 `json:"email"`
	ProviderName      string                 `json:"providerName,omitempty"`
	ProviderPhone     string                 `json:"providerPhone,omitempty"`
	ProviderID        string                 `json:"providerId,omitempty"`
	VehicleType       string                 `json:"vehicleType"`
	ServiceBidType    string                 `json:"serviceBidType"`
	ServiceType       string                 `json:"serviceType"`
	Problems          []string               `json:"problems"`
	BookingDate       time.Time              `json:"bookingDate"`
	Zone              string                 `json:"zone"`
	Status            string                 `json:"status"`
	Amount            float64                `json:"amount"`
	PaymentStatus     string                 `json:"paymentStatus"`
	UserRating        *RatingInfo            `json:"userRating,omitempty"`
	ProviderRating    *RatingInfo            `json:"providerRating,omitempty"`
	Complaint         *ComplaintInfo         `json:"complaint,omitempty"`
	ServiceRequest    *domain.ServiceRequest `json:"serviceRequest,omitempty"`
	Transaction       *domain.Transaction    `json:"transaction,omitempty"`
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
		booking := BookingResponse{
			ID:             svc.ID,
			BookingID:      fmt.Sprintf("BK%d", svc.InternalID),
			ProviderID:     svc.ProviderID,
			Status:         s.mapStatus(svc.Status),
			Amount:         svc.FinalPrice,
			PaymentStatus:  s.mapPaymentStatus(svc.PaymentStatus),
			ServiceType:    svc.ServiceType,
			BookingDate:    svc.CreatedAt,
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

		if sr, err := s.repo.FindServiceRequestByID(ctx, svc.ServiceRequestID); err == nil {
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

func (s *AdminBookingService) GetBookingByID(ctx context.Context, bookingID string) (*DetailedBookingResponse, error) {
	svc, err := s.repo.FindAcceptedServiceByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	booking := &DetailedBookingResponse{
		ID:            svc.ID,
		BookingID:     fmt.Sprintf("BK%d", svc.InternalID),
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

	if sr, err := s.repo.FindServiceRequestByID(ctx, svc.ServiceRequestID); err == nil {
		booking.VehicleType = sr.VehicleType
		booking.ServiceBidType = sr.ServiceBidType
		booking.Problems = sr.Problems
		booking.ServiceRequest = sr
	}

	if svc.OrderID != "" {
		if txn, err := s.repo.FindTransactionByOrderID(ctx, svc.OrderID); err == nil {
			booking.Transaction = txn
		}
	}

	ratings, _ := s.repo.FindRatingsByServiceIDs(ctx, []string{svc.ID})
	for _, rating := range ratings {
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

	if svc.ComplaintUserID != "" || svc.ComplaintProviderID != "" {
		complaintID := svc.ComplaintUserID
		if complaintID == "" {
			complaintID = svc.ComplaintProviderID
		}
		if complaint, err := s.repo.FindComplaintByID(ctx, complaintID); err == nil {
			booking.Complaint = &ComplaintInfo{
				ID:        complaint.ID,
				RaisedBy:  complaint.RaisedBy,
				Problem:   complaint.Problem,
				Photos:    complaint.Photos,
				Status:    complaint.Status,
				CreatedAt: complaint.CreatedAt,
			}
		}
	}

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