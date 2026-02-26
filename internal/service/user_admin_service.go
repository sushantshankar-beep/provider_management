package service

import (
	"context"
	"fmt"
	"math"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAdminService struct {
	users    *repository.UserRepo
	vehicles *repository.SavedVehiclesRepo
	bookings *repository.AcceptedServiceRepo
	amc      *repository.AMCPurchaseRepo
	vehicleRepo *repository.VehiclesRepo
}

func NewUserAdminService(
	u *repository.UserRepo,
	v *repository.SavedVehiclesRepo,
	b *repository.AcceptedServiceRepo,
	a *repository.AMCPurchaseRepo,
	vehicleRepo *repository.VehiclesRepo,
) *UserAdminService {
	return &UserAdminService{
		users:    u,
		vehicles: v,
		bookings: b,
		amc:      a,
		vehicleRepo: vehicleRepo,
	}
}

func (s *UserAdminService) GetAllUsers(
	ctx context.Context,
	filter dto.UserFilters,
	pagination dto.UserPagination,
) (*dto.UserAdminListResponse, error) {

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 20
	}
	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	query := bson.M{}

	if filter.Search != "" {
		query["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"phone": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"userCode": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	if filter.Status != "" {
		switch filter.Status {
		case "active":
			query["isActive"] = domain.AccountStatusActive
		case "inactive":
			query["isActive"] = bson.M{
				"$in": []string{
					domain.AccountStatusBlacklisted,
					domain.AccountStatusDeactivated,
				},
			}
		case "blacklisted":
			query["isActive"] = domain.AccountStatusBlacklisted
		case "deactivated":
			query["isActive"] = domain.AccountStatusDeactivated
		}
	}

	if filter.Zone != "" {
		query["selectedCityName"] = filter.Zone
	}

	if filter.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			query["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(t),
				"$lte": primitive.NewDateTimeFromTime(end),
			}
		}
	}

	users, total, err := s.users.FindAll(ctx, query, pagination.Skip, pagination.Limit)
	if err != nil {
		return nil, err
	}

	stats, _ := s.getUserStatistics(ctx, 14)

	if len(users) == 0 {
		return &dto.UserAdminListResponse{
			Users: []dto.UserAdminResponse{},
			Pagination: dto.UserMetaPagination{
				CurrentPage: pagination.Page,
				TotalPages:  0,
				TotalUsers:  0,
			},
			Stats: stats,
		}, nil
	}

	vehicleIDSet := make(map[primitive.ObjectID]struct{})

	for _, u := range users {
		if u.PrimaryVehicleID != nil {
			vehicleIDSet[*u.PrimaryVehicleID] = struct{}{}
		}
		for _, vid := range u.FallbackVehicleIDs {
			vehicleIDSet[vid] = struct{}{}
		}
	}

	vehicleIDs := make([]primitive.ObjectID, 0, len(vehicleIDSet))
	for id := range vehicleIDSet {
		vehicleIDs = append(vehicleIDs, id)
	}

	vehicles, err := s.vehicleRepo.FindByIDs(ctx, vehicleIDs)
	if err != nil {
		return nil, err
	}

	vehicleMap := make(map[primitive.ObjectID]domain.Vehicle)
	for _, v := range vehicles {
		vehicleMap[v.ID] = v
	}

	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	vehicleCountMap, _ := s.getVehicleCounts(ctx, userIDs)
	bookingMap, _ := s.getBookingCounts(ctx, userIDs)
	amcMap, _ := s.getAMCStatus(ctx, userIDs)

	result := make([]dto.UserAdminResponse, 0, len(users))

	for _, u := range users {

		userVehicleType := "N/A"

		if u.PrimaryVehicleID != nil {
			if v, ok := vehicleMap[*u.PrimaryVehicleID]; ok {
				userVehicleType = v.VehicleType
			}
		} else {
			for _, vid := range u.FallbackVehicleIDs {
				if v, ok := vehicleMap[vid]; ok {
					userVehicleType = v.VehicleType
					break
				}
			}
		}

		userAMCStatus := "No"
		if amc, ok := amcMap[u.ID]; ok {
			if amc.PlanEndDate.After(time.Now()) {
				userAMCStatus = "Active"
			} else {
				userAMCStatus = "Expired"
			}
		}

		status := u.IsActive
		if u.IsActive == "true" {
			status = "active"
		}

		result = append(result, dto.UserAdminResponse{
			ID:            u.ID,
			Name:          defaultStr(u.Name, "N/A"),
			Phone:         u.Phone,
			Email:         defaultStr(u.Email, "—"),
			UserID:        u.UserCode,
			Zone:          defaultStr(u.SelectedCityName, "N/A"),
			Status:        status,
			CreatedDate:   u.CreatedAt.Format("2006-01-02"),
			PlatformUsed:  "Android",
			VehicleType:   userVehicleType,
			VehicleCount:  vehicleCountMap[u.ID].Count,
			TotalBookings: bookingMap[u.ID],
			AMCStatus:     userAMCStatus,
			ProfileURL:    u.ImageUrl,
			Address:       u.Address,
		})
	}

	totalPages := int64(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &dto.UserAdminListResponse{
		Users: result,
		Pagination: dto.UserMetaPagination{
			CurrentPage: pagination.Page,
			TotalPages:  totalPages,
			TotalUsers:  total,
			HasNext:     pagination.Page < totalPages,
			HasPrev:     pagination.Page > 1,
		},
		Stats: stats,
	}, nil
}


func (s *UserAdminService) GetUserByID(ctx context.Context, userCode string) (*dto.UserDetailResponse, error) {

	query := bson.M{
		"userCode": userCode,
	}

	u, err := s.users.FindOne(ctx, query)
	if err != nil {
		return nil, err
	}

	var vehicleIDs []primitive.ObjectID

	if u.PrimaryVehicleID != nil {
		vehicleIDs = append(vehicleIDs, *u.PrimaryVehicleID)
	}

	if len(u.FallbackVehicleIDs) > 0 {
		vehicleIDs = append(vehicleIDs, u.FallbackVehicleIDs...)
	}

	vehicles, err := s.vehicleRepo.FindByIDs(ctx, vehicleIDs)

	if err != nil {
		return nil, err
	}

	totalBookings, err := s.bookings.CountByUserID(
		ctx,
		u.ID,
		bson.M{"paymentStatus": "paid"},
	)
	if err != nil {
		return nil, err
	}

	totalExpenses, err := s.bookings.SumExpensesByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	amcInfo, err := s.getDetailedAMCInfo(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	vehicleInfos := make([]dto.VehicleInfo, 0, len(vehicles))
	for _, v := range vehicles {
		vehicleInfos = append(vehicleInfos, dto.VehicleInfo{
			ID:            v.ID.Hex(),
			VehicleNumber: v.VehicleNumber,
			Brand:         v.Brand,
			Model:         v.Model,
			Year:          v.ModelYear,
			VehicleType:   v.VehicleType,
		})
	}

	status := u.IsActive
	if u.IsActive == "true" {
		status = "active"
	}

	return &dto.UserDetailResponse{
		UserAdminResponse: dto.UserAdminResponse{
			ID:            u.ID,
			Name:          defaultStr(u.Name, "N/A"),
			Phone:         u.Phone,
			Email:         defaultStr(u.Email, "—"),
			UserID:        u.UserCode,
			Zone:          defaultStr(u.SelectedCityName, "N/A"),
			Status:        status,
			CreatedDate:   u.CreatedAt.Format(time.RFC3339),
			VehicleCount:  len(vehicles),
			TotalBookings: int(totalBookings),
			Notes:         u.Notes,
			ProfileURL:    u.ImageUrl,
			Address:       u.Address,
			ISActive:      u.IsActive,
		},
		TotalExpenses: totalExpenses,
		Vehicles:      vehicleInfos,
		AMCInfo:       *amcInfo,
		PreferredLang: "English",
	}, nil
}

func (s *UserAdminService) UpdateUserStatus( ctx context.Context, userID, status string ) (*dto.UserAdminStatusResponse, error) {
	query := bson.M{
		"userCode": userID,
	}

	u, err := s.users.UpdateStatus(ctx, query, status)
	if err != nil {
		return nil, err
	}

	updatedStatus := u.IsActive

	return &dto.UserAdminStatusResponse{
		UserID:  u.UserCode,
		Status:  updatedStatus,
		Message: "Status updated successfully",
	}, nil
}

func (s *UserAdminService) getUserStatistics(ctx context.Context, days int) (*dto.UserStatsResponse, error) {

	stats := &dto.UserStatsResponse{}

	userStats, err := s.users.GetStatistics(ctx, days)
	if err != nil {
		fmt.Printf("Error getting user stats: %v", err)
		return stats, nil
	}

	stats.TotalRegisteredUsers = userStats.TotalUsers
	stats.TotalActiveUsers = userStats.ActiveUsers
	stats.TotalInactiveUsers = userStats.InactiveUsers

	activeAMCUsers, err := s.getActiveAMCUsersCount(ctx)
	if err != nil {
		fmt.Printf("Error counting active AMC users: %v", err)
		stats.ActiveAMCUsers = 0
	} else {
		stats.ActiveAMCUsers = activeAMCUsers
	}

	usersWithBookings, err := s.getUsersWithBookingsCount(ctx)
	if err != nil {
		fmt.Printf("Error counting users with bookings: %v", err)
		stats.UsersWithBookings = 0
	} else {
		stats.UsersWithBookings = usersWithBookings
	}

	return stats, nil
}

func (s *UserAdminService) getActiveAMCUsersCount(ctx context.Context) (int64, error) {
	activeAMCs, err := s.amc.FindActiveAMCs(ctx)
	if err != nil {
		return 0, err
	}

	userIDs := make(map[primitive.ObjectID]bool)
	for _, amc := range activeAMCs {
		userIDs[amc.UserID] = true
	}

	return int64(len(userIDs)), nil
}

func (s *UserAdminService) getUsersWithBookingsCount(ctx context.Context) (int64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"paymentStatus": "paid",
			},
		},
		{
			"$group": bson.M{
				"_id": "$user",
			},
		},
		{
			"$count": "totalUsers",
		},
	}

	cursor, err := s.bookings.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalUsers int64 `bson:"totalUsers"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.TotalUsers, nil
	}

	return 0, nil
}

func (s *UserAdminService) getVehicleCounts(ctx context.Context, userIDs []string) (map[string]struct {
	Count int
	Type  string
}, error) {
	objIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, id := range userIDs {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, oid)
		}
	}

	pipeline := []bson.M{
		{"$match": bson.M{"userId": bson.M{"$in": objIDs}}},
		{"$group": bson.M{
			"_id":   "$userId",
			"count": bson.M{"$sum": 1},
			"types": bson.M{"$addToSet": "$vehicleType"},
		}},
	}

	cursor, err := s.vehicles.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]struct {
		Count int
		Type  string
	})

	for cursor.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Count int                `bson:"count"`
			Types []string           `bson:"types"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		vehicleType := "N/A"
		if len(doc.Types) > 0 {
			vehicleType = doc.Types[0]
		}
		result[doc.ID.Hex()] = struct {
			Count int
			Type  string
		}{Count: doc.Count, Type: vehicleType}
	}

	return result, nil
}

func (s *UserAdminService) getBookingCounts(ctx context.Context, userIDs []string) (map[string]int, error) {
	objIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, id := range userIDs {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, oid)
		}
	}

	pipeline := []bson.M{
		{"$match": bson.M{"user": bson.M{"$in": objIDs}}},
		{"$group": bson.M{
			"_id":   "$user",
			"total": bson.M{"$sum": 1},
		}},
	}

	cursor, err := s.bookings.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Total int                `bson:"total"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		result[doc.ID.Hex()] = doc.Total
	}

	return result, nil
}

func (s *UserAdminService) getAMCStatus(ctx context.Context, userIDs []string) (map[string]*domain.AMCPurchase, error) {
	objIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, id := range userIDs {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, oid)
		}
	}

	amcList, err := s.amc.FindActiveByUserIDs(ctx, objIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*domain.AMCPurchase)
	for i := range amcList {
		result[amcList[i].UserID.Hex()] = &amcList[i]
	}

	return result, nil
}

func (s *UserAdminService) getDetailedAMCInfo(ctx context.Context, userID string) (*dto.AMCInfo, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &dto.AMCInfo{
			AMCStatus:          "No",
			CurrentPlanName:    "—",
			BoundVehicleNumber: "—",
			PaymentStatus:      "—",
		}, nil
	}

	amc, err := s.amc.FindActiveByUserID(ctx, userObjID)
	if err != nil || amc == nil {
		return &dto.AMCInfo{
			AMCStatus:          "No",
			CurrentPlanName:    "—",
			BoundVehicleNumber: "—",
			PaymentStatus:      "—",
		}, nil
	}

	// servicesUsed, err := s.bookings.CountCompletedInDateRange(ctx, userObjID, amc.PlanStartDate, amc.PlanEndDate)
	// if err != nil {
	// 	servicesUsed = 0
	// }

	amcStatus := "Expired"
	if amc.PlanEndDate.After(time.Now()) {
		amcStatus = "Active"
	}

	// totalServices := len(amc.PlanServicesIncluded)
	// servicesRemaining := totalServices - int(servicesUsed)
	// if servicesRemaining < 0 {
	// 	servicesRemaining = 0
	// }

	return &dto.AMCInfo{
		AMCID:        amc.ID.Hex(),
		AMCStatus:          amcStatus,
		CurrentPlanName:    defaultStr(amc.PlanName, "—"),
		StartDate:          amc.PlanStartDate,
		EndDate:            amc.PlanEndDate,
		ActivationDate:     amc.CreatedAt,
		BoundVehicleNumber: amc.Vehicle.VehicleNumber,
		// TotalServices:      totalServices,
		// ServicesUsed:       int(servicesUsed),
		// ServicesRemaining:  servicesRemaining,
		PaymentStatus:      amc.PaymentStatus,
	}, nil
}

func (s *UserAdminService) AddNote( ctx context.Context, userID string, req dto.AddNoteRequest ) error {

	if req.Content == "" {
		return fmt.Errorf("note content is required")
	}
	if req.AddedBy == "" {
		return fmt.Errorf("addedBy is required")
	}

	query := bson.M{
		"userCode": userID,
	}

	user, err := s.users.FindOne(ctx, query)

	if err != nil {
		return fmt.Errorf("user not found")
	}

	note := domain.UserNote{
		ID:        primitive.NewObjectID().Hex(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	return s.users.AddUserNote(ctx, user.UserCode, note)
}


func (s *UserAdminService) ListOfferUsageByUser(
	ctx context.Context,
	userID string,
	page, limit int64,
) ([]dto.OfferUsageItem, int64, int64, error) {

	skip := (page - 1) * limit

	services, total, err := s.bookings.GetOfferUsageByUser(
		ctx,
		userID,
		skip,
		limit,
	)
	if err != nil {
		return nil, 0, 0, err
	}

	result := make([]dto.OfferUsageItem, 0, len(services))

	for _, svc := range services {

		item := dto.OfferUsageItem{
			ServiceID:     svc.ID.Hex(),
			ServiceNumber: svc.ServiceNumber,
			UserID:        svc.User.Hex(),
			TotalDiscount: svc.TotalDiscount,
			AmountPaid:    svc.AmountPaidByUser,
			CreatedAt:     svc.CreatedAt.Format(time.RFC3339),
		}

		if svc.AppliedPromo != nil {
			item.Promo = &dto.OfferPromoInfo{
				Code:   svc.AppliedPromo.Code,
				Amount: svc.AppliedPromo.DiscountAmt,
			}
		}

		if svc.AppliedDiscount != nil {
			item.Discount = &dto.OfferDiscountInfo{
				Code:   svc.AppliedDiscount.Code,
				Amount: svc.AppliedDiscount.DiscountAmt,
			}
		}

		result = append(result, item)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return result, total, totalPages, nil
}