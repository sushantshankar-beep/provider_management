package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
    "strconv"
	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAdminService struct {
	users    *repository.UserRepo
	vehicles *repository.SavedVehiclesRepo
	bookings *repository.AcceptedServiceRepo
	amc      *repository.AMCPurchaseRepo
}

func NewUserAdminService(
	u *repository.UserRepo,
	v *repository.SavedVehiclesRepo,
	b *repository.AcceptedServiceRepo,
	a *repository.AMCPurchaseRepo,
) *UserAdminService {
	return &UserAdminService{
		users:    u,
		vehicles: v,
		bookings: b,
		amc:      a,
	}
}

type UserAdminListResponse struct {
	Users      []UserAdminResponse `json:"users"`
	Pagination Pagination          `json:"pagination"`
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
	ID            int64  `json:"id"`
	MongoID       string `json:"_id"`
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


func (s *UserAdminService) GetAllUsers(
	ctx context.Context,
	search, status, zone,amcStatus,platformUsed,vehicleType,startDate string,
	page, limit int64,
) (*UserAdminListResponse, error) {
	query := bson.M{}

	if search != "" {
		orConditions := []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
		}
		query["$or"] = orConditions
	}

	if status != "" {
		if status == "active" {
			query["isActive"] = domain.AccountStatusActive
		} else if status == "inactive" {
			query["isActive"] = bson.M{"$nin": []interface{}{domain.AccountStatusActive, "true"}}
		}
	}

	if zone != "" {
		query["selectedCityName"] = zone
	}

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			query["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(t),
				"$lte": primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}

	skip := (page - 1) * limit
	users, total, err := s.users.FindAll(ctx, query, skip, limit)

	if err != nil {
		return nil, err
	}

	stats, err := s.getUserStatistics(ctx, 14)
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		stats = &UserStatsResponse{}
	}

	if len(users) == 0 {
		return &UserAdminListResponse{
			Users: []UserAdminResponse{},
			Pagination: Pagination{
				CurrentPage: int(page),
				TotalPages:  0,
				TotalUsers:  0,
				HasNext:     false,
				HasPrev:     false,
			},
			Stats: stats,
		}, nil
	}

	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	vehicleMap, err := s.getVehicleCounts(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	bookingMap, err := s.getBookingCounts(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	amcMap, err := s.getAMCStatus(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]UserAdminResponse, 0, len(users))
	for _, u := range users {
		vehicleInfo := vehicleMap[u.ID]
		userAMCStatus := "No"
		if amc, exists := amcMap[u.ID]; exists {
			if amc.PlanEndDate.After(time.Now()) {
				userAMCStatus = "Active"
			} else {
				userAMCStatus = "Expired"
			}
		}

		userStatus := u.IsActive
		if u.IsActive == "true" {
			userStatus = "active"
		}

		userVehicleType := defaultStr(vehicleInfo.Type, "N/A")
		userPlatformUsed := "Android"

		if amcStatus != "" && strings.ToLower(amcStatus) != strings.ToLower(userAMCStatus) {
			continue
		}

		if platformUsed != "" && strings.ToLower(platformUsed) != strings.ToLower(userPlatformUsed) {
			continue
		}

		if vehicleType != "" && strings.ToLower(vehicleType) != strings.ToLower(userVehicleType) {
			continue
		}

		result = append(result, UserAdminResponse{
			ID:            u.InternalID,
			MongoID:       u.ID,
			Name:          defaultStr(u.Name, "N/A"),
			Phone:         u.Phone,
			Email:         defaultStr(u.Email, "—"),
			UserID:        fmt.Sprintf("VW%06d", u.InternalID),
			Zone:          defaultStr(u.SelectedCityName, "N/A"),
			Status:        userStatus,
			CreatedDate:   u.CreatedAt.Format("2006-01-02"),
			PlatformUsed:  userPlatformUsed,
			VehicleType:   userVehicleType,
			VehicleCount:  vehicleInfo.Count,
			TotalBookings: bookingMap[u.ID],
			AMCStatus:     userAMCStatus,
			ProfileURL:    u.ProfileURL,
			Address:       u.Address,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &UserAdminListResponse{
		Users: result,
		Pagination: Pagination{
			CurrentPage: int(page),
			TotalPages:  totalPages,
			TotalUsers:  total,
			HasNext:     page < int64(totalPages),
			HasPrev:     page > 1,
		},
		Stats: stats,
	}, nil
}

func (s *UserAdminService) GetUserByID(
	ctx context.Context,
	userID string,
) (*UserDetailResponse, error) {
	query := bson.M{}

	if strings.HasPrefix(userID, "VW") {
		var id int64
		fmt.Sscanf(userID, "VW%d", &id)
		query["id"] = id
	} else {
		query["_id"] = userID
	}

	u, err := s.users.FindOne(ctx, query)
	if err != nil {
		return nil, err
	}

	vehicles, err := s.vehicles.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	totalBookings, err := s.bookings.CountByUserID(ctx, u.ID, bson.M{"paymentStatus": "paid"})
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

	vehicleInfos := make([]VehicleInfo, len(vehicles))
	for i, v := range vehicles {
		vehicleInfos[i] = VehicleInfo{
			ID:            v.ID.Hex(),
			VehicleNumber: v.VehicleNumber,
			Brand:         v.Brand,
			Model:         v.Model,
			Year:          v.Year,
			FuelType:      v.FuelType,
			VehicleType:   v.VehicleType,
		}
	}

	status := u.IsActive
	if u.IsActive == "true" {
		status = "active"
	}

	return &UserDetailResponse{
		UserAdminResponse: UserAdminResponse{
			ID:            u.InternalID,
			MongoID:       u.ID,
			Name:          defaultStr(u.Name, "N/A"),
			Phone:         u.Phone,
			Email:         defaultStr(u.Email, "—"),
			UserID:        fmt.Sprintf("VW%06d", u.InternalID),
			Zone:          defaultStr(u.SelectedCityName, "N/A"),
			Status:        status,
			CreatedDate:   u.CreatedAt.Format(time.RFC3339),
			VehicleCount:  len(vehicles),
			TotalBookings: int(totalBookings),
			Notes:  u.Notes,
			ProfileURL:    u.ProfileURL,
			Address:       u.Address,
		},
		TotalExpenses: totalExpenses,
		Vehicles:      vehicleInfos,
		AMCInfo:       *amcInfo,
		PreferredLang: "English",
	}, nil
}

func (s *UserAdminService) UpdateUserStatus(
	ctx context.Context,
	userID, status string,
) (*UserAdminStatusResponse, error) {
	query := bson.M{}

	if strings.HasPrefix(userID, "VW") {
		var id int64
		fmt.Sscanf(userID, "VW%d", &id)
		query["id"] = id
	} else {
		query["_id"] = userID
	}

	u, err := s.users.UpdateStatus(ctx, query, status)
	if err != nil {
		return nil, err
	}

	updatedStatus := u.IsActive
	if u.IsActive == "true" {
		updatedStatus = "active"
	}

	return &UserAdminStatusResponse{
		UserID:  fmt.Sprintf("VW%06d", u.InternalID),
		Status:  updatedStatus,
		Message: "Status updated successfully",
	}, nil
}

func (s *UserAdminService) getUserStatistics(ctx context.Context, days int) (*UserStatsResponse, error) {
	stats := &UserStatsResponse{}

	userStats, err := s.users.GetStatistics(ctx, days)
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		return stats, nil
	}

	stats.TotalRegisteredUsers = userStats.TotalUsers
	stats.TotalActiveUsers = userStats.ActiveUsers
	stats.TotalInactiveUsers = userStats.InactiveUsers

	activeAMCUsers, err := s.getActiveAMCUsersCount(ctx)
	if err != nil {
		log.Printf("Error counting active AMC users: %v", err)
		stats.ActiveAMCUsers = 0
	} else {
		stats.ActiveAMCUsers = activeAMCUsers
	}

	usersWithBookings, err := s.getUsersWithBookingsCount(ctx)
	if err != nil {
		log.Printf("Error counting users with bookings: %v", err)
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

func (s *UserAdminService) getDetailedAMCInfo(ctx context.Context, userID string) (*AMCInfo, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &AMCInfo{
			AMCStatus:          "No",
			CurrentPlanName:    "—",
			BoundVehicleNumber: "—",
			PaymentStatus:      "—",
		}, nil
	}

	amc, err := s.amc.FindActiveByUserID(ctx, userObjID)
	if err != nil || amc == nil {
		return &AMCInfo{
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

	return &AMCInfo{
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

func (s *UserAdminService) AddNote(
	ctx context.Context,
	userID string,
	req AddNoteRequest,
) error {

	if req.Content == "" {
		return fmt.Errorf("note content is required")
	}
	if req.AddedBy == "" {
		return fmt.Errorf("addedBy is required")
	}

	cleanID := strings.TrimPrefix(userID, "VW")
	internalID, err := strconv.ParseInt(cleanID, 10, 64)
	log.Println(internalID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	user, err := s.users.FindByInternalID(ctx, internalID)
	log.Println("cskabcjab",user)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	note := domain.UserNote{
		ID:        primitive.NewObjectID().Hex(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	log.Println("jdsbhjsb",note)
	log.Println("hdvw",user.ID)
	return s.users.AddUserNote(ctx, user.InternalID, note)
}
