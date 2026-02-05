package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"math"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strings"
	"time"
	"unicode"
)

type AdminBookingService struct {
	repo            *repository.AdminBookingRepo
	Invoicerepo     *repository.InvoiceRepo
	transactionRepo *repository.TransactionRepo
	settlementRepo  *repository.SettlementHistoryRepository
	ratingRepo  *repository.RatingRepo
}

func NewAdminBookingService(repo *repository.AdminBookingRepo, Invoicerepo *repository.InvoiceRepo, transactionRepo *repository.TransactionRepo, settlementRepo *repository.SettlementHistoryRepository,ratingRepo  *repository.RatingRepo) *AdminBookingService {
	return &AdminBookingService{
		repo:            repo,
		Invoicerepo:     Invoicerepo,
		transactionRepo: transactionRepo,
		settlementRepo:  settlementRepo,
		ratingRepo: ratingRepo,
	}

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

func (s *AdminBookingService) GetAllBookings(ctx context.Context, filters dto.BookingFilters, pagination dto.BookingPagination, zoneFilter bson.M) (*dto.BookingListResponse, error) {

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10
	}
	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	filter := bson.M{}

	if filters.Status != "" {
		if filters.Status == "in_progress" {
			filter["status"] = bson.M{
				"$in": []string{
					StatusStarted,
					StatusReached,
					StatusOTPVerified,
					StatusInProgress,
				},
			}
		} else {
			filter["status"] = mapStatusLabelToDB(filters.Status)
		}
	}

	if filters.PaymentStatus != "" {
		filter["paymentStatus"] = filters.PaymentStatus
	}

	if filters.ServiceType != "" {
		filter["serviceType"] = filters.ServiceType
	}

	if filters.VehicleType != "" {
		filter["vehicleType"] = filters.VehicleType
	}

	if filters.UserID != "" {
		if err := s.addUserFilter(ctx, filters.UserID, filter); err != nil {
			return s.buildEmptyResponse(filters, pagination)
		}
	}

	if filters.ProviderID != "" {
		if err := s.addProviderFilter(ctx, filters.ProviderID, filter); err != nil {
			return s.buildEmptyResponse(filters, pagination)
		}
	}

	if filters.BookingID != "" {
		s.addBookingIDFilter(filters.BookingID, filter)
	}

	if filters.Search != "" {
		if err := s.addSearchFilter(ctx, filters.Search, filter); err != nil {
			return s.buildEmptyResponse(filters, pagination)
		}
	}

	s.addDateFilters(filters, filter)

	if len(zoneFilter) > 0 {
		allowedZones := s.extractAllowedZones(zoneFilter)
		if len(allowedZones) > 0 {
			if err := s.addZoneFilter(ctx, allowedZones, filter); err != nil {
				return s.buildEmptyResponse(filters, pagination)
			}
		}
	}

	services, total, err := s.repo.FindAcceptedServices(ctx, filter, pagination.Skip, pagination.Limit, pagination.Sort)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookings: %w", err)
	}

	if len(services) == 0 {
		return s.buildEmptyResponse(filters, pagination)
	}

	bookings, err := s.buildBookingResponses(ctx, services)
	if err != nil {
		return nil, err
	}

	// Get stats
	stats, err := s.GetBookingStats(ctx, map[string]string{
		"status":        filters.Status,
		"paymentStatus": filters.PaymentStatus,
		"serviceType":   filters.ServiceType,
		"vehicleType":   filters.VehicleType,
		"userId":        filters.UserID,
		"providerId":    filters.ProviderID,
		"bookingId":     filters.BookingID,
		"startDate":     filters.StartDate,
		"endDate":       filters.EndDate,
		"search":        filters.Search,
	})
	if err != nil {
		stats = &dto.BookingStats{}
	}

	totalPages := int64(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &dto.BookingListResponse{
		Bookings: bookings,
		Stats: dto.BookingStats{
			TotalBookings:      stats.TotalBookings,
			InProgressBookings: stats.InProgressBookings,
			PendingBookings:    stats.PendingBookings,
			CompletedBookings:  stats.CompletedBookings,
			CancelledBookings:  stats.CancelledBookings,
			TotalRevenue:       stats.TotalRevenue,
		},
		Pagination: dto.BookingPaginationMeta{
			CurrentPage: pagination.Page,
			TotalPages:  totalPages,
			Total:       total,
			Limit:       pagination.Limit,
			HasNext:     pagination.Page < totalPages,
			HasPrev:     pagination.Page > 1,
		},
	}, nil
}

func (s *AdminBookingService) addUserFilter(ctx context.Context, userID string, filter bson.M) error {

	if !primitive.IsValidObjectID(userID) {
		return fmt.Errorf("invalid user object id")
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter["user"] = objID
	return nil
}

func (s *AdminBookingService) addProviderFilter(ctx context.Context, providerID string, filter bson.M) error {

	if !primitive.IsValidObjectID(providerID) {
		return fmt.Errorf("invalid user object id")
	}

	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return err
	}

	filter["user"] = objID
	return nil
}

func (s *AdminBookingService) addBookingIDFilter(serviceNumber string, filter bson.M) {
	if serviceNumber == "" {
		return
	}
	filter["serviceNumber"] = serviceNumber
}


func (s *AdminBookingService) addSearchFilter(ctx context.Context, search string, filter bson.M) error {
	searchConditions := []bson.M{}

	if strings.HasPrefix(search, "VHBK") {
		serviceNumber := strings.TrimSpace(search)

		searchConditions = append(searchConditions, bson.M{
			"serviceNumber": serviceNumber,
		})
	}

	if users, err := s.repo.FindUsersBySearch(ctx, search); err == nil && len(users) > 0 {
		userIDs := s.convertToObjectIDs(users)
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
		return nil
	}

	return fmt.Errorf("no results found for search")
}

func (s *AdminBookingService) addDateFilters(filters dto.BookingFilters, filter bson.M) {
	if filters.StartDate != "" {
		if sd, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			sd = time.Date(sd.Year(), sd.Month(), sd.Day(), 0, 0, 0, 0, time.UTC)
			filter["startedAt"] = bson.M{"$gte": sd}
		}
	}
	if filters.EndDate != "" {
		if ed, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			ed = time.Date(ed.Year(), ed.Month(), ed.Day(), 23, 59, 59, 999999999, time.UTC)
			filter["completedAt"] = bson.M{"$lte": ed}
		}
	}
}

func (s *AdminBookingService) addZoneFilter(ctx context.Context, allowedZones []string, filter bson.M) error {
	matchingSRs, err := s.repo.FindServiceRequestsByZones(ctx, allowedZones)
	if err != nil {
		return err
	}

	if len(matchingSRs) == 0 {
		return fmt.Errorf("no service requests found in allowed zones")
	}

	filter["serviceRequest"] = bson.M{"$in": matchingSRs}
	return nil
}

func (s *AdminBookingService) buildBookingResponses(ctx context.Context, services []domain.AcceptedService) ([]dto.BookingResponse, error) {

	userIDs, providerIDs, serviceRequestIDs, serviceIDs := s.collectIDs(services)

	userMap, providerMap, serviceRequestMap, ratingsMap, err := s.fetchRelatedData(ctx, userIDs, providerIDs, serviceRequestIDs, serviceIDs)
	if err != nil {
		return nil, err
	}

	bookings := make([]dto.BookingResponse, 0, len(services))

	for _, svc := range services {
		booking := s.mapServiceToBookingResponse(svc, userMap, providerMap, serviceRequestMap, ratingsMap)
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

func (s *AdminBookingService) collectIDs(services []domain.AcceptedService) ([]string, []string, []string, []string) {
	userIDSet := make(map[string]bool)
	providerIDSet := make(map[string]bool)
	serviceRequestIDSet := make(map[string]bool)
	serviceIDs := make([]string, 0, len(services))

	for _, svc := range services {
		userIDSet[svc.User.Hex()] = true
		providerIDSet[svc.Provider.Hex()] = true
		serviceRequestIDSet[svc.ServiceNumber] = true
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

	return userIDs, providerIDs, serviceRequestIDs, serviceIDs
}

func (s *AdminBookingService) fetchRelatedData(ctx context.Context, userIDs, providerIDs, serviceRequestIDs, serviceIDs []string) (
	map[string]*domain.User,
	map[string]*domain.Provider,
	map[string]*domain.ServiceRequest,
	map[string][]*domain.Rating,
	error,
) {
	userMap := make(map[string]*domain.User)
	if len(userIDs) > 0 {
		users, _ := s.repo.FindUsersByIDs(ctx, userIDs)
		for i := range users {
			userMap[users[i].ID] = &users[i]
		}
	}

	providerMap := make(map[string]*domain.Provider)
	if len(providerIDs) > 0 {
		providers, _ := s.repo.FindProvidersByIDs(ctx, providerIDs)
		for i := range providers {
			providerMap[providers[i].ID.Hex()] = &providers[i]
		}
	}

	serviceRequestMap := make(map[string]*domain.ServiceRequest)
	if len(serviceRequestIDs) > 0 {
		serviceRequests, _ := s.repo.FindServiceRequestsByIDs(ctx, serviceRequestIDs)
		for i := range serviceRequests {
			serviceRequestMap[serviceRequests[i].ID.Hex()] = &serviceRequests[i]
		}
	}

	ratingsMap := make(map[string][]*domain.Rating)
	if len(serviceIDs) > 0 {
	
		bookingObjectIDs := make([]primitive.ObjectID, 0, len(serviceIDs))
		for _, id := range serviceIDs {
			if oid, err := primitive.ObjectIDFromHex(id); err == nil {
				bookingObjectIDs = append(bookingObjectIDs, oid)
			}
		}
	
		ratings, _ := s.ratingRepo.FindRatingsByBookingIDs(ctx, bookingObjectIDs)
	
		for _, rating := range ratings {
			bookingID := rating.BookingID.Hex()
			ratingsMap[bookingID] =
				append(ratingsMap[bookingID], rating)
		}
	}
	
	return userMap, providerMap, serviceRequestMap, ratingsMap, nil
}

func (s *AdminBookingService) mapServiceToBookingResponse(
	svc domain.AcceptedService,
	userMap map[string]*domain.User,
	providerMap map[string]*domain.Provider,
	serviceRequestMap map[string]*domain.ServiceRequest,
	ratingsMap map[string][]*domain.Rating,
) dto.BookingResponse {
	booking := dto.BookingResponse{
		ID:            svc.ID.Hex(),
		BookingID:     svc.ServiceNumber,
		ProviderID:    svc.Provider.Hex(),
		Status:        domain.ServiceStatus(s.mapStatus(svc.Status)),
		Amount:        svc.FinalPrice,
		PaymentStatus: s.mapPaymentStatus(svc.PaymentStatus),
		VehicleType:   svc.VehicleType,
		ServiceType:   svc.ServiceType,
		BookingDate:   svc.CreatedAt,
		CompletedAt:   svc.Timestamps.CompletedAt,
	}

	if user, ok := userMap[svc.User.Hex()]; ok {
		booking.UserID = user.UserCode
		booking.CustomerName = user.Name
		booking.Phone = user.Phone
		booking.Email = user.Email
	}

	if provider, ok := providerMap[svc.Provider.Hex()]; ok {
		booking.ProviderName = provider.Name
		booking.ProviderPhone = provider.Phone
		booking.Zone = provider.City
	}

	if sr, ok := serviceRequestMap[svc.ServiceRequest.Hex()]; ok {
		booking.VehicleType = sr.VehicleType
		booking.ServiceBidType = sr.ServiceBidType
	}

	if bookingRatings, ok := ratingsMap[svc.ID.Hex()]; ok {
		for _, rating := range bookingRatings {
			ratingInfo := &dto.RatingInfo{
				Stars:             rating.Stars,
				Review:            rating.Review,
				RecommendToFriend: rating.Recommended,
			}
		
			if rating.RatingType == "user" {
				booking.UserRating = ratingInfo
			}
		
			if rating.RatingType == "provider" {
				booking.ProviderRating = ratingInfo
			}
		}
	}

	return booking
}

func (s *AdminBookingService) buildEmptyResponse(filters dto.BookingFilters, pagination dto.BookingPagination) (*dto.BookingListResponse, error) {

	stats, _ := s.GetBookingStats(context.Background(), map[string]string{
		"status":        filters.Status,
		"paymentStatus": filters.PaymentStatus,
		"serviceType":   filters.ServiceType,
		"vehicleType":   filters.VehicleType,
		"userId":        filters.UserID,
		"providerId":    filters.ProviderID,
		"bookingId":     filters.BookingID,
		"startDate":     filters.StartDate,
		"endDate":       filters.EndDate,
		"search":        filters.Search,
	})

	totalPages := int64(0)
	if pagination.Limit > 0 {
		totalPages = int64(math.Ceil(float64(0) / float64(pagination.Limit)))
	}

	return &dto.BookingListResponse{
		Bookings: []dto.BookingResponse{},
		Stats: dto.BookingStats{
			TotalBookings:      stats.TotalBookings,
			InProgressBookings: stats.InProgressBookings,
			PendingBookings:    stats.PendingBookings,
			CompletedBookings:  stats.CompletedBookings,
			CancelledBookings:  stats.CancelledBookings,
			TotalRevenue:       stats.TotalRevenue,
		},
		Pagination: dto.BookingPaginationMeta{
			CurrentPage: pagination.Page,
			TotalPages:  totalPages,
			Total:       0,
			Limit:       pagination.Limit,
			HasNext:     pagination.Page < totalPages,
			HasPrev:     pagination.Page > 1,
		},
	}, nil
}

func (s *AdminBookingService) GetBookingStats(ctx context.Context, params map[string]string) (*dto.BookingStats, error) {
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

	completedServices, _, err := s.repo.FindAcceptedServices(ctx, completedFilter, 0, 0, "")
	if err != nil {
		return nil, err
	}

	var serviceIDs []string
	for _, svc := range completedServices {
		serviceIDs = append(serviceIDs, svc.ID.Hex())
	}

	revenue, err := s.transactionRepo.SumAmountByServiceIDs(ctx, serviceIDs)
	if err != nil {
		return nil, err
	}

	return &dto.BookingStats{
		TotalBookings:      total,
		InProgressBookings: inProgress,
		PendingBookings:    pending,
		CompletedBookings:  completed,
		CancelledBookings:  cancelled,
		TotalRevenue:       revenue,
	}, nil
}

func (s *AdminBookingService) GetBookingByID(ctx context.Context, bookingID string) (*dto.DetailedBookingResponse, error) {

	filter := bson.M{"serviceNumber": bookingID}

	allServices, _, err := s.repo.FindAcceptedServices(ctx, filter, 0, 100, "-createdAt")

	if err != nil || len(allServices) == 0 {
		return nil, fmt.Errorf("booking not found")
	}

	svc := allServices[0]

	status := s.mapStatus(svc.Status)

	booking := &dto.DetailedBookingResponse{
		ID:            svc.ID.Hex(),
		BookingID:     svc.ServiceNumber,
		ProviderID:    svc.Provider.Hex(),
		Status:        domain.ServiceStatus(status),
		Amount:        svc.FinalPrice,
		PaymentStatus: svc.PaymentStatus,
		ServiceType:   svc.ServiceType,
		BookingDate:   svc.CreatedAt,
		VehicleType:   svc.VehicleType,
		VehicleNumber: svc.VehicleNumber,
		Brand:         svc.Brand,
		Model:         svc.Model,
		Year:          svc.ModelYear,
		FuelType:      svc.FuelType,
		Problems:      svc.Issues,
		// Description:      sr.Description,
		// Location:         sr.Address,
		SettlementStatus: dto.SettlementStatus(svc.SettlementStatus),
		Notes:            svc.Notes,
	}

	if user, err := s.repo.FindUserByID(ctx, svc.User.Hex()); err == nil {
		booking.UserID = user.UserCode
		booking.CustomerName = user.Name
		booking.Phone = user.Phone
		booking.Email = user.Email
		booking.OTP = user.ServiceOTP
	}

	if provider, err := s.repo.FindProviderByID(ctx, svc.Provider.Hex()); err == nil {
		booking.ProviderName = provider.Name
		booking.ProviderPhone = provider.Phone
		booking.MechanicType = provider.VehicleType
		booking.Zone = provider.City
		booking.RatingProvider = provider.Rating
	}

	basePrice := svc.BasePrice
	if basePrice == 0 {
		basePrice = svc.FinalPrice
	}
	finalPrice := svc.FinalPrice
	gstAmount := finalPrice * 0.18
	subtotal := finalPrice + gstAmount
	discount := 0.0
	if basePrice > finalPrice {
		discount = basePrice - finalPrice
	}

	booking.PaymentDetails = &dto.PaymentDetailsInfo{
		ServiceCharge: basePrice,
		Discount:      discount,
		Subtotal:      utils.RoundTo2(subtotal),
		GST:           gstAmount,
		Total:         subtotal,
	}

	timestamps := domain.ServiceTimestamps{}
	createdAt := svc.CreatedAt
	timestamps.CreatedAt = &createdAt

	if svc.Timestamps != nil {
		if svc.Timestamps.CreatedAt != nil {
			timestamps.CreatedAt = svc.Timestamps.CreatedAt
		}
		timestamps.StartedAt = svc.Timestamps.StartedAt
		timestamps.ReachedAt = svc.Timestamps.ReachedAt
		timestamps.InProgressAt = svc.Timestamps.InProgressAt
		timestamps.OtpVerified = svc.Timestamps.OtpVerified
		timestamps.CompletedAt = svc.Timestamps.CompletedAt
		timestamps.CancelledAt = svc.Timestamps.CancelledAt
	}

	booking.Timestamps = timestamps

	var transaction *domain.Transaction
	if svc.ID.Hex() != "" {
		txn, err := s.repo.FindTransactionByServiceID(ctx, svc.ID.Hex())

		if err == nil && txn != nil {
			transaction = txn
			// paymentStatus := "failure"
			// if txn.TxnResponse != nil {
			// 	if respMap, ok := txn.TxnResponse.(map[string]interface{}); ok {
			// 		if txnStatus, ok := respMap["status"].(string); ok {
			// 			paymentStatus = txnStatus
			// 		}
			// 	}
			// }
			booking.TransactionDetails = &dto.TransactionDetailsInfo{
				TransactionID: txn.TxnID,
				Amount:        txn.Amount,
				Status:        txn.Status,
				// PaymentStatus: paymentStatus,
				PaymentMethod: txn.Method,
			}
		}
	}

	ratings, _ := s.ratingRepo.FindRatingsByBookingID(ctx, svc.ID)
	var userRatings []int
	var providerRatings []int
	
	for _, rating := range ratings {
		ratingInfo := &dto.RatingInfo{
			Stars:             rating.Stars,
			Review:            rating.Review,
			RecommendToFriend: rating.Recommended,
			CreatedAt:         rating.CreatedAt,
			RatedBy:           rating.RaterID.Hex(),
			RatedTo:           rating.RateeID.Hex(),
		}
	
		if rating.RatingType == "user" {
			booking.UserRating = ratingInfo
			userRatings = append(userRatings, rating.Stars)
		}
	
		if rating.RatingType == "provider" {
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

	booking.Rating = &dto.RatingsSummary{
		User:     userAvg,
		Provider: providerAvg,
	}

	if svc.ComplaintUserID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintUserID); err == nil {

			booking.UserComplaint = &dto.ComplaintInfo{
				ID:          complaint.ID,
				ComplaintID: complaint.ComplaintNumber,
				RaisedBy:    complaint.RaisedBy,
				Status:      complaint.Status,
				CreatedAt:   complaint.CreatedAt,
				UserComplaint: domain.ComplaintSide{
					Problem:  complaint.UserComplaint.Problem,
					Photos:   complaint.UserComplaint.Photos,
					RaisedAt: complaint.UserComplaint.RaisedAt,
				},
			}
		}
	}

	if svc.ComplaintProviderID != "" {
		if complaint, err := s.repo.FindComplaintByID(ctx, svc.ComplaintProviderID); err == nil {
			booking.ProviderComplaint = &dto.ComplaintInfo{
				ID:          complaint.ID,
				ComplaintID: complaint.ComplaintNumber,
				RaisedBy:    complaint.RaisedBy,
				Status:      complaint.Status,
				CreatedAt:   complaint.CreatedAt,
				ProviderComplaint: domain.ComplaintSide{
					Problem:  complaint.ProviderComplaint.Problem,
					Photos:   complaint.ProviderComplaint.Photos,
					RaisedAt: complaint.ProviderComplaint.RaisedAt,
				},
			}
		}
	}

	if transaction != nil && transaction.Status == "paid" {

		userPaid := transaction.Amount

		baseAmount := userPaid / 1.18
		gstAmount := userPaid - baseAmount
		commissionAmount := baseAmount * 0.20

		booking.ProviderEarnings = &dto.ProviderEarningsInfo{
			UserPaid: utils.RoundTo2(userPaid),
			Commission: dto.CommissionInfo{
				Percentage: 20,
				Amount:     utils.RoundTo2(commissionAmount),
			},
			GST: dto.GSTInfo{
				Percentage: 18,
				Amount:     utils.RoundTo2(gstAmount),
			},
			PaymentMode:  transaction.PaymentSource,
			ExpectedTime: "24 Hours",
		}
	}

	if settlement, err := s.settlementRepo.FindLatestByServiceID(ctx, svc.ID); err == nil && settlement != nil {

		booking.ProviderEarnings = &dto.ProviderEarningsInfo{
			UserPaid: utils.RoundTo2(settlement.OriginalAmount),

			Commission: dto.CommissionInfo{
				Percentage: settlement.CommissionPercent,
				Amount:     utils.RoundTo2(settlement.CommissionAmount),
			},

			GST: dto.GSTInfo{
				Percentage: settlement.GSTPercent,
				Amount:     utils.RoundTo2(settlement.GSTAmount),
			},

			NetPayout:          utils.RoundTo2(settlement.NetAmount),
			PayoutStatus:       string(settlement.SettlementStatus),
			PaymentMode:        transaction.PaymentSource,
			ExpectedPayoutDate: &settlement.CreatedAt,
			ExpectedTime:       "24 Hours",
		}
	}

	var allBookings []dto.BookingSummary
	for _, service := range allServices {
		allBookings = append(allBookings, dto.BookingSummary{
			ID:          service.ID.Hex(),
			Status:      service.Status,
			CreatedAt:   service.CreatedAt,
			CancelledBy: service.CancelledBy,
		})
	}
	booking.AllBookings = allBookings

	return booking, nil
}

func (s *AdminBookingService) CancelBooking(ctx context.Context, bookingID string) (*dto.DetailedBookingResponse, error) {

	filter := bson.M{"serviceNumber": bookingID}
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

	if svc.Provider.Hex() != "" {
		err = s.repo.UpdateProviderIsAssigned(ctx, svc.Provider.Hex(), false)
		if err != nil {
			log.Printf("Failed to update provider status: %v", err)
		}
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) MarkBookingCompleted(ctx context.Context, bookingID string) (*dto.DetailedBookingResponse, error) {

	filter := bson.M{"serviceNumber": bookingID}
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

	if svc.Provider.Hex() != "" {
		err = s.repo.UpdateProviderIsAssigned(ctx, svc.Provider.Hex(), false)
		if err != nil {
			log.Printf("Failed to update provider status: %v", err)
		}
	}

	return s.GetBookingByID(ctx, bookingID)
}

func (s *AdminBookingService) mapStatus(status domain.ServiceStatus) string {
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
		return string(status)
	}
}

func (s *AdminBookingService) mapPaymentStatus(status string) string {
	if status == "success" || status == "paid" {
		return "paid"
	}
	return status
}

func (s *AdminBookingService) AddNote(ctx context.Context, bookingID string, req dto.AddNoteRequest) error {

	if req.Content == "" {
		return fmt.Errorf("note content is required")
	}
	if req.AddedBy == "" {
		return fmt.Errorf("addedBy is required")
	}

	svcs, _, err := s.repo.FindAcceptedServices(
		ctx,
		bson.M{"serviceNumber": bookingID},
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

func extractZone(address string) string {
	parts := strings.Split(address, ",")
	for i := len(parts) - 2; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}
		partNoNumbers := strings.Map(func(r rune) rune {
			if unicode.IsDigit(r) {
				return -1
			}
			return r
		}, part)

		partNoNumbers = strings.TrimSpace(partNoNumbers)
		if partNoNumbers != "" {
			return partNoNumbers
		}
	}
	return "N/A"
}

func (s *AdminBookingService) extractAllowedZones(zoneFilter bson.M) []string {
	zones := []string{}

	if stateFilter, ok := zoneFilter["state"]; ok {
		if inMap, ok := stateFilter.(bson.M); ok {
			if inArray, ok := inMap["$in"]; ok {
				if arr, ok := inArray.([]string); ok {
					zones = append(zones, arr...)
				}
			}
		}
	}

	if cityFilter, ok := zoneFilter["city"]; ok {
		if inMap, ok := cityFilter.(bson.M); ok {
			if inArray, ok := inMap["$in"]; ok {
				if arr, ok := inArray.([]string); ok {
					zones = append(zones, arr...)
				}
			}
		}
	}

	if orFilter, ok := zoneFilter["$or"]; ok {
		if orArray, ok := orFilter.([]bson.M); ok {
			for idx, condition := range orArray {
				log.Printf("  Processing $or condition %d: %+v", idx, condition)
				extractedFromOr := s.extractAllowedZones(condition)
				zones = append(zones, extractedFromOr...)
			}
		}
	}

	return zones
}

func (s *AdminBookingService) convertToObjectIDs(users []domain.User) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, 0, len(users))
	for _, user := range users {
		objID, _ := primitive.ObjectIDFromHex(user.ID)
		ids = append(ids, objID)
	}
	return ids
}

func (s *AdminBookingService) GetInvoice(ctx context.Context, bookingID string) (*domain.Invoice, error) {
	return s.Invoicerepo.GetByID(ctx, bookingID)
}
