package service

import (
	"time"
	"fmt"
	"math"
	"strconv"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"provider_management/internal/dto"
	"provider_management/internal/utils"
	"provider_management/internal/domain"
	"provider_management/internal/constants"
	"provider_management/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProviderAdminService struct {
	providers      *repository.ProviderRepo
	services       *repository.AcceptedServiceRepo
	admin          *repository.AdminRepository
	zone           *repository.ZoneRepo
	role           *repository.RoleRepository
	settlementRepo *repository.ProviderSettlementRepo
	settlementHistoryRepo *repository.SettlementHistoryRepository
	serviceRequestRepo *repository.ServiceRequestRepo
}

func NewProviderAdminService(
	p *repository.ProviderRepo,
	s *repository.AcceptedServiceRepo,
	a *repository.AdminRepository,
	z *repository.ZoneRepo,
	r *repository.RoleRepository,
	t *repository.ProviderSettlementRepo,
	h *repository.SettlementHistoryRepository,
	u *repository.ServiceRequestRepo,
) *ProviderAdminService {
	return &ProviderAdminService{
		providers:      p,
		services:       s,
		admin:          a,
		zone:           z,
		role:           r,
		settlementRepo: t,
		settlementHistoryRepo:  h,
		serviceRequestRepo: u,
	}
}

type JobHistoryItem struct {
	BookingID      string  `json:"booking_id"`
	ServiceType    string  `json:"service_type"`
	BookingDate    string  `json:"booking_date"`
	AmountEarned   float64 `json:"amount_earned"`
	IsAMC          string  `json:"is_amc"`
	PaymentStatus  string  `json:"payment_status"`
	SettlementDate string  `json:"settlement_date"`
}


func (s *ProviderAdminService) GetAllProviders( ctx context.Context, filters dto.ProviderFilters, pagination dto.ProviderPagination, zoneFilter bson.M) (*dto.ProviderListResponse, error) {

	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.Limit < 1 {
		pagination.Limit = 20
	}

	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	var conditions []bson.M

	if len(zoneFilter) > 0 {
		conditions = append(conditions, zoneFilter)
	}

	inactiveStatuses := []string{
		string(domain.AccountStatusSuspended),
		string(domain.AccountStatusBlacklisted),
		string(domain.AccountStatusDeactivated),
	}

	activeStatus := []string{
		string(domain.AccountStatusActive),
	}

	if filters.Filter == "inactive" {
		conditions = append(conditions, bson.M{"isActive": bson.M{"$in": inactiveStatuses}})
	}

	if filters.Filter == "active" {
		conditions = append(conditions, bson.M{"isActive": bson.M{"$in": activeStatus}})
	}

	if filters.Filter != "" {
		switch filters.Filter {
		case "Pending":
			conditions = append(conditions, bson.M{"status": domain.StatusPending})
		case "verified":
			conditions = append(conditions, bson.M{"status": domain.StatusActive})
		case "rejected":
			conditions = append(conditions, bson.M{"status": domain.StatusRejected})
		}
	}

	if filters.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			conditions = append(conditions, bson.M{
				"createdAt": bson.M{
					"$gte": primitive.NewDateTimeFromTime(t),
					"$lte": primitive.NewDateTimeFromTime(endOfDay),
				},
			})
		}
	}

	if filters.Search != "" {
		searchConditions := []bson.M{
			{"name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"phone": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"city": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"address": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"vehicleType": bson.M{"$elemMatch": bson.M{"$regex": filters.Search, "$options": "i"}}},
		}

		if strings.HasPrefix(strings.ToUpper(filters.Search), "PRO") {
			idSearch := strings.TrimPrefix(strings.ToUpper(filters.Search), "PRO")
			searchConditions = append(searchConditions, bson.M{
				"$or": []bson.M{
					{"id": bson.M{"$regex": idSearch, "$options": "i"}},
					{"_id": bson.M{"$regex": idSearch, "$options": "i"}},
				},
			})
		} else {
			searchConditions = append(searchConditions, bson.M{
				"$or": []bson.M{
					{"id": bson.M{"$regex": filters.Search, "$options": "i"}},
					{"_id": bson.M{"$regex": filters.Search, "$options": "i"}},
				},
			})
		}

		conditions = append(conditions, bson.M{"$or": searchConditions})
	}

	if filters.Name != "" {
		conditions = append(conditions, bson.M{"name": bson.M{"$regex": filters.Name, "$options": "i"}})
	}
	if filters.Mobile != "" {
		conditions = append(conditions, bson.M{"phone": bson.M{"$regex": filters.Mobile, "$options": "i"}})
	}
	if filters.ProviderID != "" {
		idSearch := strings.TrimPrefix(strings.ToUpper(filters.ProviderID), "PRO")
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{"id": bson.M{"$regex": idSearch, "$options": "i"}},
				{"_id": bson.M{"$regex": idSearch, "$options": "i"}},
			},
		})
	}
	if filters.Zone != "" {
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{"city": bson.M{"$regex": filters.Zone, "$options": "i"}},
				{"address": bson.M{"$regex": filters.Zone, "$options": "i"}},
			},
		})
	}
	if filters.VehicleType != "" {
		conditions = append(conditions, bson.M{"vehicleType": bson.M{"$elemMatch": bson.M{"$regex": filters.VehicleType, "$options": "i"}}})
	}

	if filters.Status != "" {
		conditions = append(conditions, bson.M{"status": filters.Status})
	}

	if filters.KYCStatus != "" {
		kycStatusLower := strings.ToLower(filters.KYCStatus)
		switch kycStatusLower {
		case "active", "verified":
			conditions = append(conditions, bson.M{"status": domain.StatusActive})
		case "pending":
			conditions = append(conditions, bson.M{"status": domain.StatusPending})
		case "rejected":
			conditions = append(conditions, bson.M{"status": domain.StatusRejected})
		}
	}

	if filters.AccountStatus != "" {
		accountStatusLower := strings.ToLower(filters.AccountStatus)
		switch accountStatusLower {
		case "active":
			conditions = append(conditions, bson.M{
				"$or": []bson.M{
					{"isActive": domain.AccountStatusActive},
					{"isActive": true},
				},
			})
		case "suspended":
			conditions = append(conditions, bson.M{"isActive": domain.AccountStatusSuspended})
		case "blacklisted":
			conditions = append(conditions, bson.M{"isActive": domain.AccountStatusBlacklisted})
		}
	}

	query := bson.M{}
	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	providers, total, err := s.providers.FindAll(ctx, query, pagination.Skip, pagination.Limit, pagination.Sort)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %v", err)
	}

	countQuery := query
	if andConditions, ok := countQuery["$and"].([]bson.M); ok && len(andConditions) == 1 {
		for k, v := range andConditions[0] {
			countQuery[k] = v
		}
		delete(countQuery, "$and")
	}

	activeCount, _ := s.providers.CountByStatus(ctx, countQuery, "isActive", domain.AccountStatusActive)
	inactiveCount, _ := s.providers.CountByMultipleStatuses(ctx, countQuery, "isActive", inactiveStatuses)
	pendingKycCount, _ := s.providers.CountByStatus(ctx, countQuery, "status", domain.StatusPending)

	formattedProviders := make([]dto.ProviderAllResponse, len(providers))
	for i, p := range providers {
		totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, p.ID.Hex())
		if err != nil {
		    fmt.Printf("Error getting service stats for provider %s: %v", p.ID, err)
			totalJobs = 0
			completedJobs = 0
		}

		kyc := "Not Submitted"
		if len(p.IdentityProof) > 0 {
			switch p.IdentityProof[0].Verified {
			case domain.VerificationApproved:
				kyc = "Verified"
			case domain.VerificationRejected:
				kyc = "Rejected"
			default:
				kyc = "Pending"
			}
		}

		account := "Active"
		switch p.IsActive {
		case domain.AccountStatusSuspended:
			account = "Suspended"
		case domain.AccountStatusBlacklisted:
			account = "Blacklisted"
		default:
			account = "Active"
		}

		vehicle := "N/A"
		if len(p.VehicleType) > 0 {
			vehicle = strings.Join(p.VehicleType, ", ")
		}

		providerIDStr := fmt.Sprintf("PRO%d", p.InternalID)
		if p.InternalID == 0 {
			if len(p.ID) >= 6 {
				providerIDStr = fmt.Sprintf("PRO%s", p.ID[len(p.ID)-6:])
			} else {
				providerIDStr = fmt.Sprintf("PRO%s", p.ID)
			}
		}

		formattedProviders[i] = dto.ProviderAllResponse{
			ID:            p.ID.Hex(),
			ProviderID:    providerIDStr,
			Name:          defaultStr(p.Name, "N/A"),
			Mobile:        p.Phone,
			Email:         defaultStr(p.Email, "N/A"),
			KYC:           kyc,
			Account:       account,
			Vehicle:       vehicle,
			Zone:          defaultStr(p.City, defaultStr(p.Address, "N/A")),
			DOJ:           formatDate(p.CreatedAt),
			ProfileURL:    p.ProfileURL,
			IsServiceOn:   p.IsServiceOn,
			IdentityProof: p.IdentityProof,
			AddressProof:  p.AddressProof,
			CancelCheque:  p.CancelCheque,
			IsActive:      string(p.IsActive),
			Status:        p.Status,
			TotalJobs:     totalJobs,
			CompletedJobs: completedJobs,
		}
	}

	var totalPages int64 = 1
    if total > 0 && pagination.Limit > 0 {
	    totalPages = int64(math.Ceil(float64(total) / float64(pagination.Limit)))
    }


	return &dto.ProviderListResponse{
		Providers: formattedProviders,
		Counts: dto.ProviderCounts{
			Total:      total,
			Active:     activeCount,
			Inactive:   inactiveCount,
			PendingKYC: pendingKycCount,
		},
		Pagination: dto.ProviderMetaPagination{
			CurrentPage: pagination.Page,
			TotalPages:  totalPages,
			Total:       total,
			Limit:       pagination.Limit,
			HasNext:     pagination.Page < totalPages,
			HasPrev:     pagination.Page > 1,
		},
	}, nil
}

func (s *ProviderAdminService) GetProviderByID(ctx context.Context, id string) (*dto.ProviderDetailResponse, error) {
	provider, err := s.providers.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, provider.ID.Hex())

	if err != nil {
		fmt.Printf("Error getting service stats: %v", err)
		totalJobs = 0
		completedJobs = 0
	}

	recentServices, err := s.services.FindByProviderID(ctx, provider.ID.Hex(), 10)

	if err != nil {
		fmt.Printf("Error fetching recent services: %v", err)
		recentServices = []domain.Service{}
	}

	kycStatus := "Not Submitted"
	if len(provider.IdentityProof) > 0 {
		if provider.IdentityProof[0].Verified == domain.VerificationApproved {
			kycStatus = "Verified"
		} else if provider.IdentityProof[0].Verified == domain.VerificationRejected {
			kycStatus = "Rejected"
		} else {
			kycStatus = "Pending"
		}
	}

	accountStatus := "Active"
	if provider.IsActive == domain.AccountStatusSuspended {
		accountStatus = "Suspended"
	} else if provider.IsActive == domain.AccountStatusBlacklisted {
		accountStatus = "Blacklisted"
	} else if provider.IsActive != "active" && provider.IsActive != "true" {
		accountStatus = "Inactive"
	}

	providerIDStr := fmt.Sprintf("PRO%d", provider.InternalID)
	if provider.InternalID == 0 {
		if len(provider.ID) >= 6 {
			providerIDStr = fmt.Sprintf("PRO%s", provider.ID[len(provider.ID)-6:])
		} else {
			providerIDStr = fmt.Sprintf("PRO%s", provider.ID)
		}
	}

	brandNames := constants.GetBrandNamesByIDs(provider.ProviderBrands)
	serviceNames := constants.GetServiceNamesByIDs(provider.ProviderServices)

	return &dto.ProviderDetailResponse{
		ID:                   provider.ID.Hex(),
		ProviderID:           providerIDStr,
		Name:                 defaultStr(provider.Name, "N/A"),
		Phone:                provider.Phone,
		Email:                defaultStr(provider.Email, "N/A"),
		AlternateContact:     provider.AlternateContact,
		ProfileURL:           provider.ProfileURL,
		Address:              provider.Address,
		PermanentAddress:     provider.PermanentAddress,
		City:                 provider.City,
		Status:               provider.Status,
		Account:              accountStatus,
		KYC:                  kycStatus,
		VehicleType:          provider.VehicleType,
		VehicleNumber:        provider.VehicleNumber,
		ProviderBrands:       brandNames,
		ProviderServices:     serviceNames,
		GSTNumber:            provider.GSTNumber,
		CompanyName:          provider.CompanyName,
		Description:          provider.Description,
		Zone:                 defaultStr(provider.City, defaultStr(provider.Address, "N/A")),
		DOJ:                  formatDateDetailed(provider.CreatedAt),
		IdentityProof:        provider.IdentityProof,
		AddressProof:         provider.AddressProof,
		CancelCheque:         provider.CancelCheque,
		BankDetails:          provider.BankDetails,
		IsServiceOn:          provider.IsServiceOn,
		IsActive:             string(provider.IsActive),
		IsSocketConnected:    provider.IsSocketConnected,
		Location:             &provider.Location,
		TotalJobs:            totalJobs,
		CompletedJobs:        completedJobs,
		RecentServices:       recentServices,
		CommissionPercentage: provider.CommissionPercentage,
		Notes:                provider.Notes,
		ApprovedAt:           formatDateDetailed(provider.ApprovedAt),
	}, nil
}

func (s *ProviderAdminService) UpdateProviderStatus(ctx context.Context, id, status string) (*domain.Provider, error) {

	validStatuses := map[string]bool{
		domain.StatusActive:   true,
		domain.StatusPending:  true,
		domain.StatusRejected: true,
		"deactive":            true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status")
	}

	return s.providers.UpdateStatus(ctx, id, status)
}

func (s *ProviderAdminService) UpdateProviderKYC(ctx context.Context, id, kycStatus string) (*domain.Provider, error) {

	statusMap := map[string]string{
		"active":   domain.StatusActive,
		"verified": domain.StatusActive,
		"pending":  domain.StatusPending,
		"rejected": domain.StatusRejected,
	}

	statusLower := strings.ToLower(kycStatus)

	mappedStatus, exists := statusMap[statusLower]
	if !exists {
		return nil, fmt.Errorf("invalid KYC status")
	}

	return s.providers.UpdateKYCStatus(ctx, id, mappedStatus)
}

func (s *ProviderAdminService) VerifyDocument( ctx context.Context, id, documentType, documentID, action string ) (*domain.Provider, error) {

	if action != "approve" && action != "reject" {
		return nil, fmt.Errorf("invalid action (approve/reject required)")
	}

	verificationStatus := domain.VerificationApproved

	if action == "reject" {
		verificationStatus = domain.VerificationRejected
	}

	provider, err := s.providers.UpdateDocumentVerification(ctx, id, documentType, documentID, verificationStatus)

	if err != nil {
		return nil, err
	}

	allApproved := true
	hasRejected := false

	for _, proof := range provider.IdentityProof {
		if proof.Verified == domain.VerificationRejected {
			hasRejected = true
			break
		}
		if proof.Verified != domain.VerificationApproved {
			allApproved = false
		}
	}

	if !hasRejected {
		for _, proof := range provider.AddressProof {
			if proof.Verified == domain.VerificationRejected {
				hasRejected = true
				break
			}
			if proof.Verified != domain.VerificationApproved {
				allApproved = false
			}
		}
	}

	if !hasRejected && provider.CancelCheque != nil {
		if provider.CancelCheque.Verified == domain.VerificationRejected {
			hasRejected = true
		} else if provider.CancelCheque.Verified != domain.VerificationApproved {
			allApproved = false
		}
	}

	var newStatus string
	if hasRejected {
		newStatus = domain.StatusRejected
	} else if allApproved {
		newStatus = domain.StatusActive
	} else {
		newStatus = domain.StatusPending
	}

	return s.providers.UpdateKYCStatus(ctx, id, newStatus)
}

func (s *ProviderAdminService) UpdateProviderAccountAction(ctx context.Context, id, action string) (*domain.Provider, error) {

	statusMap := map[string]string{
		"activate":  domain.AccountStatusActive,
		"suspend":   domain.AccountStatusSuspended,
		"blacklist": domain.AccountStatusBlacklisted,
	}

	mappedStatus, exists := statusMap[action]
	if !exists {
		return nil, fmt.Errorf("invalid action. Use: activate, suspend, or blacklist")
	}

	return s.providers.UpdateAccountStatus(ctx, id, mappedStatus)
}

func (s *ProviderAdminService) UpdateProviderCommission(ctx context.Context, id string, commissionPercentage float64) (*domain.Provider, error) {

	if commissionPercentage < 0 || commissionPercentage > 100 {
		return nil, fmt.Errorf("commission percentage must be between 0 and 100")
	}

	return s.providers.UpdateCommission(ctx, id, commissionPercentage)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatDateDetailed(t time.Time) string {
	return t.Format("Jan 2, 2006")
}

func (s *ProviderAdminService) GetDocumentURL( ctx context.Context, providerID, documentType, documentID string ) (*dto.DocumentResponse, error) {

	provider, err := s.providers.FindByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	switch documentType {

	case "identity":
		if len(provider.IdentityProof) == 0 {
			return nil, fmt.Errorf("no identity proof documents found")
		}

		if documentID != "" {
			for _, proof := range provider.IdentityProof {
				if proof.ID.Hex() == documentID {
					return &dto.DocumentResponse{
						DocumentID:   proof.ID.Hex(),
						DocumentType: "identity",
						Type:         proof.Type,
						File:         proof.File,
						Verified:     proof.Verified,
					}, nil
				}
			}
			return nil, fmt.Errorf("identity document with ID %s not found", documentID)
		}

		proof := provider.IdentityProof[0]
		return &dto.DocumentResponse{
			DocumentID:   proof.ID.Hex(),
			DocumentType: "identity",
			Type:         proof.Type,
			File:         proof.File,
			Verified:     proof.Verified,
		}, nil

	case "address":
		if len(provider.AddressProof) == 0 {
			return nil, fmt.Errorf("no address proof documents found")
		}

		if documentID != "" {
			for _, proof := range provider.AddressProof {
				if proof.ID.Hex() == documentID {
					return &dto.DocumentResponse{
						DocumentID:   proof.ID.Hex(),
						DocumentType: "address",
						Type:         proof.Type,
						File:         proof.File,
						Verified:     proof.Verified,
					}, nil
				}
			}
			return nil, fmt.Errorf("address document with ID %s not found", documentID)
		}

		proof := provider.AddressProof[0]
		return &dto.DocumentResponse{
			DocumentID:   proof.ID.Hex(),
			DocumentType: "address",
			Type:         proof.Type,
			File:         proof.File,
			Verified:     proof.Verified,
		}, nil

	case "cancel_cheque":
		if provider.CancelCheque == nil {
			return nil, fmt.Errorf("no cancel cheque document found")
		}

		return &dto.DocumentResponse{
			DocumentType: "cancel_cheque",
			File:         provider.CancelCheque.File,
			Verified:     provider.CancelCheque.Verified,
		}, nil

	default:
		return nil, fmt.Errorf("invalid document type. Use: identity, address, or cancel_cheque")
	}
}

func (s *ProviderAdminService) AddNote( ctx context.Context, providerID string, req dto.AddNoteRequest ) error {

	if req.Content == "" {
		return fmt.Errorf("note content is required")
	}
	if req.AddedBy == "" {
		return fmt.Errorf("addedBy is required")
	}

	objectID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return fmt.Errorf("invalid provider id")
	}

	note := domain.ProviderNote{
		ID:        primitive.NewObjectID(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	return s.providers.AddProviderNote(ctx, objectID, note)
}
func (s *ProviderAdminService) CreateProvider(ctx context.Context, req dto.CreateProviderRequest, createdBy primitive.ObjectID, role string) (*domain.Provider, error) {

	providerID := time.Now().UnixNano() / 1000000

	provider := &domain.Provider{
		ID:               primitive.NewObjectID(),
		InternalID:       int64(providerID),
		Name:             req.Name,
		CompanyName:      req.CompanyName,
		Phone:            req.Phone,
		Email:            req.Email,
		AlternateContact: req.AlternateContact,
		City:             req.City,
		Address:          req.Address,
		PermanentAddress: req.PermanentAddress,
		VehicleType:      req.VehicleType,
		VehicleNumber:    req.VehicleNumber,
		ProviderBrands:   req.ProviderBrands,
		ProviderServices: req.ProviderServices,
		GSTNumber:        req.GSTNumber,
		Description:      req.Description,
		ProfileURL:       req.ProfileURL,
		Status:           domain.StatusPending,
		IsActive:         domain.AccountStatusActive,
		IsServiceOn:      false,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	provider.Location = domain.GeoPoint{
		Type:        "Point",
		Coordinates: []float64{0, 0},
	}

	if req.AccountHolderName != "" || req.AccountNumber != "" {
		provider.BankDetails = &domain.BankDetails{
			AccountHolderName: req.AccountHolderName,
			AccountNumber:     req.AccountNumber,
			IfscCode:          req.IfscCode,
			BranchName:        req.BranchName,
			Upi:               req.Upi,
		}
	}

	provider.IdentityProof = []domain.Proof{}
	provider.AddressProof = []domain.Proof{}

	if len(req.IdentityProofs) > 0 {
		provider.IdentityProof = req.IdentityProofs
	}

	if len(req.AddressProofs) > 0 {
		provider.AddressProof = req.AddressProofs
	}

	if req.CancelCheque != nil {
		provider.CancelCheque = req.CancelCheque
	}

	err := s.providers.Create(ctx, provider)

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %v", err)
	}

	return provider, nil
}

func (s *ProviderAdminService) UpdateProvider( ctx context.Context, id string, req dto.UpdateProviderRequest, updatedBy primitive.ObjectID,
) (*domain.Provider, error) {

	_, err := s.providers.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("provider not found")
	}

	update := bson.M{
		"updatedAt": time.Now(),
		"updatedBy": updatedBy,
	}

	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.CompanyName != "" {
		update["companyName"] = req.CompanyName
	}
	if req.Phone != "" {
		update["phone"] = req.Phone
	}
	if req.Email != "" {
		update["email"] = req.Email
	}
	if req.City != "" {
		update["city"] = req.City
	}
	if req.Address != "" {
		update["address"] = req.Address
	}
	if req.ProfileURL != "" {
		update["profileUrl"] = req.ProfileURL
	}
	if len(req.VehicleType) > 0 {
		update["vehicleType"] = req.VehicleType
	}
	if req.GSTNumber != "" {
		update["GSTNumber"] = req.GSTNumber
	}

	if len(req.ProviderBrands) > 0 {
		update["providerBrands"] = req.ProviderBrands
	}

	if len(req.ProviderServices) > 0 {
		update["providerServices"] = req.ProviderServices
	}

	if req.AccountHolderName != "" || req.AccountNumber != "" {
		update["bankDetails"] = &domain.BankDetails{
			AccountHolderName: req.AccountHolderName,
			AccountNumber:     req.AccountNumber,
			IfscCode:          req.IfscCode,
			BranchName:        req.BranchName,
			Upi:               req.Upi,
		}
	}

	if len(req.IdentityProofs) > 0 {
		update["identityProof"] = req.IdentityProofs
	}

	if len(req.AddressProofs) > 0 {
		update["addressProof"] = req.AddressProofs
	}

	if req.CancelCheque != nil {
		update["cancelCheque"] = req.CancelCheque
	}

	return s.providers.Update(ctx, id, update)
}

func (s *ProviderAdminService) GetZoneStats(ctx context.Context, adminZones map[string][]string, adminID primitive.ObjectID) (*dto.ProviderZoneStatsResponse, error) {
	zoneNames := adminZones["zoneName"]

	if len(zoneNames) == 0 {
		return &dto.ProviderZoneStatsResponse{
			Zones:                  []dto.ProviderZoneStats{},
			TotalZones:             0,
			TotalProviders:         0,
			TotalActivationMembers: 0,
		}, nil
	}

	results := make([]dto.ProviderZoneStats, 0, len(zoneNames))

	totalProvidersAcrossZones := 0
	totalActivationMembersAcrossZones := 0

	for _, zoneName := range zoneNames {
		subAdminIDs, err := s.getSubAdminsByZone(ctx, zoneName)
		if err != nil {
			results = append(results, dto.ProviderZoneStats{
				ZoneName:               zoneName,
				TotalProviders:         0,
				TotalActivationMembers: 0,
			})
			continue
		}

		totalActivationMembers := len(subAdminIDs)
		totalActivationMembersAcrossZones += totalActivationMembers

		if totalActivationMembers == 0 {
			results = append(results, dto.ProviderZoneStats{
				ZoneName:               zoneName,
				TotalProviders:         0,
				TotalActivationMembers: 0,
			})
			continue
		}

		totalProviders, err := s.countProvidersByCreators(ctx, subAdminIDs)
		if err != nil {
			results = append(results, dto.ProviderZoneStats{
				ZoneName:               zoneName,
				TotalProviders:         0,
				TotalActivationMembers: totalActivationMembers,
			})
			continue
		}

		totalProvidersAcrossZones += totalProviders

		results = append(results, dto.ProviderZoneStats{
			ZoneName:               zoneName,
			TotalProviders:         totalProviders,
			TotalActivationMembers: totalActivationMembers,
		})
	}

	return &dto.ProviderZoneStatsResponse{
		Zones:                   results,
		TotalZones:              len(results),
		TotalProviders:          totalProvidersAcrossZones,
		NewlyActivatedProviders: totalProvidersAcrossZones,
		TotalActivationMembers:  totalActivationMembersAcrossZones,
	}, nil
}

func (s *ProviderAdminService) getSubAdminsByZone(ctx context.Context, zoneName string) ([]primitive.ObjectID, error) {
	roleIDs, err := s.role.FindRoleIDsByZoneName(ctx, zoneName)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []primitive.ObjectID{}, nil
	}

	adminIDs, err := s.admin.FindAdminIDsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	subAdminIDs := make([]primitive.ObjectID, 0, len(adminIDs))

	for _, adminID := range adminIDs {

		admin, err := s.admin.FindByID(ctx, adminID)
		if err != nil {
			continue
		}

		if admin.Role == domain.RoleTypeSubAdmin {
			subAdminIDs = append(subAdminIDs, adminID)
		}
	}

	return subAdminIDs, nil
}

func (s *ProviderAdminService) countProvidersByCreators(ctx context.Context, adminIDs []primitive.ObjectID) (int, error) {
	return s.providers.CountByCreators(ctx, adminIDs)
}

func (s *ProviderAdminService) GetZoneActivationTeam( ctx context.Context, zoneName string, adminZones map[string][]string ) (*dto.ProviderActivationTeamResponse, error) {

	roleIDs, err := s.role.FindRoleIDsByZoneName(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %v", err)
	}

	if len(roleIDs) == 0 {
		return &dto.ProviderActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []dto.ProviderActivationTeamMember{},
		}, nil
	}

	adminIDs, err := s.admin.FindAdminIDsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find admins: %v", err)
	}

	if len(adminIDs) == 0 {
		return &dto.ProviderActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []dto.ProviderActivationTeamMember{},
		}, nil
	}

	subAdminIDs := make([]primitive.ObjectID, 0)
	for _, adminID := range adminIDs {
		admin, err := s.admin.FindByID(ctx, adminID)
		if err != nil {
			continue
		}
		if admin.Role == domain.RoleTypeSubAdmin {
			subAdminIDs = append(subAdminIDs, adminID)
		}
	}

	if len(subAdminIDs) == 0 {
		return &dto.ProviderActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []dto.ProviderActivationTeamMember{},
		}, nil
	}

	totalProviders, err := s.providers.CountByCreatedBy(ctx, subAdminIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to count providers: %v", err)
	}

	team, err := s.getActivationTeamDetails(ctx, subAdminIDs, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get activation team details: %v", err)
	}

	return &dto.ProviderActivationTeamResponse{
		ZoneName:        zoneName,
		TotalProviders:  int64(totalProviders),
		TotalActivators: int64(len(subAdminIDs)),
		Team:            team,
	}, nil
}

func (s *ProviderAdminService) getActivationTeamDetails( ctx context.Context, adminIDs []primitive.ObjectID, zoneName string, ) ([]dto.ProviderActivationTeamMember, error) {

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":  bson.M{"$in": adminIDs},
				"role": domain.RoleTypeSubAdmin,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "providerschemas",
				"localField":   "_id",
				"foreignField": "createdBy",
				"as":           "providers",
			},
		},
		{
			"$addFields": bson.M{
				"totalProviders": bson.M{"$size": "$providers"},
			},
		},
		{
			"$project": bson.M{
				"_id":            1,
				"personName":     "$name",
				"assignZone":     zoneName,
				"totalProviders": 1,
			},
		},
		{
			"$sort": bson.M{"totalProviders": -1},
		},
	}

	team, err := s.admin.AggregateActivationTeam(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (s *ProviderAdminService) GetActivationPersonProviders( ctx context.Context, personID string, zoneName string, filters dto.ProviderFilters, pagination dto.ProviderPagination, adminZones map[string][]string,
) (*dto.ProviderListResponse, error) {

	adminObjectID, err := primitive.ObjectIDFromHex(personID)
	if err != nil {
		return nil, fmt.Errorf("invalid activation person id")
	}

	if zoneName != "" {
		admin, err := s.admin.FindByID(ctx, adminObjectID)
		if err != nil {
			return nil, fmt.Errorf("admin not found")
		}

		role, err := s.role.FindByID(ctx, admin.RoleID)
		if err != nil {
			return nil, fmt.Errorf("role not found")
		}

		hasZone := false
		for _, zn := range role.ZoneName {
			if zn == zoneName {
				hasZone = true
				break
			}
		}

		if !hasZone {
			return nil, fmt.Errorf("admin does not have access to zone %s", zoneName)
		}
	}

	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.Limit < 1 {
		pagination.Limit = 20
	}

	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	conditions := []bson.M{
		{"createdBy": adminObjectID},
	}

	if filters.Search != "" {
		searchConditions := []bson.M{
			{"name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"phone": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filters.Search, "$options": "i"}},
		}

		conditions = append(conditions, bson.M{"$or": searchConditions})
	}

	query := bson.M{"$and": conditions}

	providers, total, err := s.providers.FindAll(ctx, query, pagination.Skip, pagination.Limit, pagination.Sort)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers")
	}

	inactiveStatuses := []string{
		string(domain.AccountStatusSuspended),
		string(domain.AccountStatusBlacklisted),
		string(domain.AccountStatusDeactivated),
	}

	activeCount, _ := s.providers.CountByStatus(ctx, query, "isActive", domain.AccountStatusActive)
	inactiveCount, _ := s.providers.CountByMultipleStatuses(ctx, query, "isActive", inactiveStatuses)
	pendingKycCount, _ := s.providers.CountByStatus(ctx, query, "status", domain.StatusPending)

	formatted := make([]dto.ProviderAllResponse, len(providers))
	for i, p := range providers {
		totalJobs, completedJobs, _ := s.services.GetServiceStats(ctx, p.ID.Hex())

		kyc := "Not Submitted"
		if len(p.IdentityProof) > 0 {
			switch p.IdentityProof[0].Verified {
			case domain.VerificationApproved:
				kyc = "Verified"
			case domain.VerificationRejected:
				kyc = "Rejected"
			default:
				kyc = "Pending"
			}
		}

		account := "Active"
		switch p.IsActive {
		case domain.AccountStatusSuspended:
			account = "Suspended"
		case domain.AccountStatusBlacklisted:
			account = "Blacklisted"
		}

		vehicle := "N/A"
		if len(p.VehicleType) > 0 {
			vehicle = strings.Join(p.VehicleType, ", ")
		}

		providerIDStr := fmt.Sprintf("PRO%d", p.InternalID)
		if p.InternalID == 0 {
			if len(p.ID) >= 6 {
				providerIDStr = fmt.Sprintf("PRO%s", p.ID[len(p.ID)-6:])
			} else {
				providerIDStr = fmt.Sprintf("PRO%s", p.ID)
			}
		}

		formatted[i] =dto.ProviderAllResponse{
			ID:            p.ID.Hex(),
			ProviderID:    providerIDStr,
			Name:          defaultStr(p.Name, "N/A"),
			Mobile:        p.Phone,
			Email:         defaultStr(p.Email, "N/A"),
			Zone:          p.City,
			KYC:           kyc,
			Account:       account,
			Vehicle:       vehicle,
			DOJ:           formatDate(p.CreatedAt),
			ProfileURL:    p.ProfileURL,
			IsServiceOn:   p.IsServiceOn,
			IdentityProof: p.IdentityProof,
			AddressProof:  p.AddressProof,
			CancelCheque:  p.CancelCheque,
			IsActive:      string(p.IsActive),
			Status:        p.Status,
			TotalJobs:     totalJobs,
			CompletedJobs: completedJobs,
		}
	}

	totalPages := int64(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &dto.ProviderListResponse{
		Providers: formatted,
		Counts: dto.ProviderCounts{
			Total:      total,
			Active:     activeCount,
			Inactive:   inactiveCount,
			PendingKYC: pendingKycCount,
		},
		Pagination: dto.ProviderMetaPagination{
			CurrentPage: pagination.Page,
			TotalPages:  totalPages,
			Total:       total,
			Limit:       pagination.Limit,
			HasNext:     pagination.Page < totalPages,
			HasPrev:     pagination.Page > 1,
		},
	}, nil
}

type ProviderEarningResponse struct {
	BookingID        string  `json:"booking_id"`
	ServiceType      string  `json:"service_type"`
	BookingDate      string  `json:"booking_date"`
	AmountEarned     float64 `json:"amount_earned"`
	AMCvsRegular     string  `json:"amc_vs_regular"`
	PaymentStatus    string  `json:"payment_status"`
	SettlementDate   string  `json:"settlement_date,omitempty"`
	DeductionAmount  float64 `json:"deduction_amount,omitempty"`
	SettlementAmount float64 `json:"settlement_amount"`
}

type ProviderEarningsSummary struct {
	TotalEarnings      float64 `json:"total_earnings"`
	TotalAmountSettled float64 `json:"total_amount_settled"`
	PendingAmount      float64 `json:"pending_amount"`
	CompletedJobs      int64   `json:"completed_jobs"`
}

func (s *ProviderAdminService) GetProviderEarnings(
	ctx context.Context,
	providerID string,
	page int,
	limit int,
	sortField string,
	sortOrder string,
	search string,
) ([]ProviderEarningResponse, *ProviderEarningsSummary, int64, error) {

	providerObjID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid provider id")
	}

	filter := bson.M{
		"providerId": providerObjID,
	}

	skip := (page - 1) * limit
	order := -1
	if strings.ToLower(sortOrder) == "asc" {
		order = 1
	}

	settlements, total, err := s.settlementHistoryRepo.GetSettlementRecords(
		ctx,
		filter,
		skip,
		limit,
		sortField,
		order,
	)

	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get settlement records: %v", err)
	}

	var responses []ProviderEarningResponse
	var totalEarnings, totalSettled, pendingAmount float64
	completedCount := int64(0)

	for _, settlement := range settlements {
		serviceType := "Regular"
		bookingID := ""
		indianTime := settlement.CreatedAt.Add(5*time.Hour + 30*time.Minute)
		bookingDate := indianTime.Format("2006-01-02 15:04:05")
		paymentStatus := "Pending"

		service, err := s.services.FindByObjectIDs(ctx, settlement.ServiceID)

		if err == nil && service != nil {
			bookingID = "BK" + strconv.FormatInt(service.InternalID, 10)
			serviceTime := service.CreatedAt.Add(5*time.Hour + 30*time.Minute)
			bookingDate = serviceTime.Format("2006-01-02 15:04:05")
			paymentStatus = string(settlement.SettlementStatus)
           
			serviceReq, err := s.serviceRequestRepo.FindByID(ctx, service.ServiceRequestID.Hex())
			if err == nil && serviceReq != nil && len(serviceReq.Problems) > 0 {
				serviceType = serviceReq.Problems[0] 
			}
		}

		deductionAmount := 0.0
		if settlement.HasDeduction {
			deductionAmount = settlement.DeductionAmount
		}

		settlementAmount := settlement.NetAmount - deductionAmount
		settlementDateStr := ""

		if settlement.SettlementStatus == "settled" {
			totalSettled += settlementAmount
			completedCount++
			if settlement.SettledAt != nil {
				settledTime := settlement.SettledAt.Add(5*time.Hour + 30*time.Minute)
				settlementDateStr = settledTime.Format("2006-01-02 15:04:05")
			}
		} else {
			pendingAmount += settlementAmount
		}

		totalEarnings += settlement.NetAmount

		resp := ProviderEarningResponse{
			BookingID:        bookingID,
			ServiceType:      serviceType,
			BookingDate:      bookingDate,
			AmountEarned:     settlement.NetAmount,
			AMCvsRegular:     "Regular booking",
			PaymentStatus:    paymentStatus,
			SettlementDate:   settlementDateStr,
			DeductionAmount:  deductionAmount,
			SettlementAmount: settlementAmount,
		}

		responses = append(responses, resp)
	}

	summary := &ProviderEarningsSummary{
		TotalEarnings:      utils.RoundTo2(totalEarnings),
		TotalAmountSettled: utils.RoundTo2(totalSettled),
		PendingAmount:      utils.RoundTo2(pendingAmount),
		CompletedJobs:      completedCount,
	}

	return responses, summary, total, nil
}