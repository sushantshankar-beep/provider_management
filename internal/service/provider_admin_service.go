package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/constants"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"time"
)

type ProviderAdminService struct {
	providers *repository.ProviderRepo
	services  *repository.AcceptedServiceRepo
	admin     *repository.AdminRepository
	zone      *repository.ZoneRepo
	role      *repository.RoleRepository
}

func NewProviderAdminService(
	p *repository.ProviderRepo,
	s *repository.AcceptedServiceRepo,
	a *repository.AdminRepository,
	z *repository.ZoneRepo,
	r *repository.RoleRepository,
) *ProviderAdminService {
	return &ProviderAdminService{
		providers: p,
		services:  s,
		admin:     a,
		zone:      z,
		role:      r,
	}
}

type ProviderListResponse struct {
	Providers  []ProviderResponse `json:"providers"`
	Counts     ProviderCounts     `json:"counts"`
	Pagination Pagination         `json:"pagination"`
}

type ProviderResponse struct {
	ID            string               `json:"id"`
	ProviderID    string               `json:"provider_id"`
	Name          string               `json:"name"`
	Mobile        string               `json:"mobile"`
	Email         string               `json:"email"`
	KYC           string               `json:"kyc"`
	Account       string               `json:"account"`
	Vehicle       string               `json:"vehicle"`
	Zone          string               `json:"zone"`
	DOJ           string               `json:"doj"`
	ProfileURL    string               `json:"profile_url"`
	IsServiceOn   bool                 `json:"is_service_on"`
	IsActive      string               `json:"is_active"`
	IdentityProof []domain.Proof       `json:"identity_proof"`
	AddressProof  []domain.Proof       `json:"address_proof"`
	CancelCheque  *domain.CancelCheque `json:"cancel_cheque"`
	Status        string               `json:"status"`
	TotalJobs     int64                `json:"total_jobs"`
	CompletedJobs int64                `json:"completed_jobs"`
}

type ProviderCounts struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Inactive   int64 `json:"inactive"`
	PendingKYC int64 `json:"pending_kyc"`
}

type ProviderDetailResponse struct {
	ID                   string                `json:"id"`
	ProviderID           string                `json:"provider_id"`
	Name                 string                `json:"name"`
	Phone                string                `json:"phone"`
	Email                string                `json:"email"`
	AlternateContact     string                `json:"alternate_contact"`
	ProfileURL           string                `json:"profile_url"`
	Address              string                `json:"address"`
	PermanentAddress     string                `json:"permanent_address"`
	City                 string                `json:"city"`
	Status               string                `json:"status"`
	Account              string                `json:"account"`
	KYC                  string                `json:"kyc"`
	VehicleType          []string              `json:"vehicle_type"`
	VehicleNumber        string                `json:"vehicle_number"`
	ProviderBrands       []string              `json:"provider_brands"`
	ProviderServices     []string              `json:"provider_services"`
	GSTNumber            string                `json:"gst_number"`
	CompanyName          string                `json:"company_name"`
	Description          string                `json:"description"`
	Zone                 string                `json:"zone"`
	DOJ                  string                `json:"doj"`
	IdentityProof        []domain.Proof        `json:"identity_proof"`
	AddressProof         []domain.Proof        `json:"address_proof"`
	CancelCheque         *domain.CancelCheque  `json:"cancel_cheque"`
	BankDetails          *domain.BankDetails   `json:"bank_details"`
	IsServiceOn          bool                  `json:"is_service_on"`
	IsActive             string                `json:"is_active"`
	IsSocketConnected    bool                  `json:"is_socket_connected"`
	Location             *domain.GeoPoint      `json:"location"`
	TotalJobs            int64                 `json:"total_jobs"`
	CompletedJobs        int64                 `json:"completed_jobs"`
	RecentServices       []domain.Service      `json:"recent_services"`
	CommissionPercentage float64               `json:"commission_percentage,omitempty"`
	Notes                []domain.ProviderNote `json:"notes,omitempty"`
	ApprovedAt           string                `json:"approved_at"`
}

type DocumentResponse struct {
	DocumentID   string `json:"document_id,omitempty"`
	DocumentType string `json:"document_type"`
	Type         string `json:"type,omitempty"`
	File         string `json:"file"`
	Verified     string `json:"verified"`
}

type ZoneStatsResponse struct {
	Zones []domain.ZoneStats `json:"zones"`
	TotalZones               int                `json:"totalZones"`
    TotalProviders           int                `json:"totalProviders"`
    TotalActivationMembers   int                `json:"totalActivationMembers"`
	NewlyActivatedProviders  int                `json:"newlyActivatedProviders"`
}

type ActivationTeamResponse struct {
	ZoneName        string                        `json:"zoneName"`
	TotalProviders  int64                         `json:"totalProviders"`
	TotalActivators int64                         `json:"totalActivators"`
	Team            []domain.ActivationTeamMember `json:"team"`
}

func (s *ProviderAdminService) GetAllProviders(
	ctx context.Context,
	pageStr, limitStr, sort, search, status, name, mobile,
	providerID, kycStatus, accountStatus, vehicleType, zone, startDate, filter string,
	zoneFilter bson.M,
) (*ProviderListResponse, error) {

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	skip := int64((page - 1) * limit)

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

	if filter == "inactive" {
		conditions = append(conditions, bson.M{"isActive": bson.M{"$in": inactiveStatuses}})
	}

	if filter == "active" {
		conditions = append(conditions, bson.M{"isActive": bson.M{"$in": activeStatus}})
	}

	if filter != "" {
		switch filter {
		case "Pending":
			conditions = append(conditions, bson.M{"status": domain.StatusPending})
		case "verified":
			conditions = append(conditions, bson.M{"status": domain.StatusActive})
		case "rejected":
			conditions = append(conditions, bson.M{"status": domain.StatusRejected})
		}
	}

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			conditions = append(conditions, bson.M{
				"createdAt": bson.M{
					"$gte": primitive.NewDateTimeFromTime(t),
					"$lte": primitive.NewDateTimeFromTime(endOfDay),
				},
			})
		}
	}

	if search != "" {
		searchConditions := []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"city": bson.M{"$regex": search, "$options": "i"}},
			{"address": bson.M{"$regex": search, "$options": "i"}},
			{"vehicleType": bson.M{"$elemMatch": bson.M{"$regex": search, "$options": "i"}}},
		}

		if strings.HasPrefix(strings.ToUpper(search), "PRO") {
			idSearch := strings.TrimPrefix(strings.ToUpper(search), "PRO")
			searchConditions = append(searchConditions, bson.M{
				"$or": []bson.M{
					{"id": bson.M{"$regex": idSearch, "$options": "i"}},
					{"_id": bson.M{"$regex": idSearch, "$options": "i"}},
				},
			})
		} else {
			searchConditions = append(searchConditions, bson.M{
				"$or": []bson.M{
					{"id": bson.M{"$regex": search, "$options": "i"}},
					{"_id": bson.M{"$regex": search, "$options": "i"}},
				},
			})
		}

		conditions = append(conditions, bson.M{"$or": searchConditions})
	}

	if name != "" {
		conditions = append(conditions, bson.M{"name": bson.M{"$regex": name, "$options": "i"}})
	}
	if mobile != "" {
		conditions = append(conditions, bson.M{"phone": bson.M{"$regex": mobile, "$options": "i"}})
	}
	if providerID != "" {
		idSearch := strings.TrimPrefix(strings.ToUpper(providerID), "PRO")
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{"id": bson.M{"$regex": idSearch, "$options": "i"}},
				{"_id": bson.M{"$regex": idSearch, "$options": "i"}},
			},
		})
	}
	if zone != "" {
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{"city": bson.M{"$regex": zone, "$options": "i"}},
				{"address": bson.M{"$regex": zone, "$options": "i"}},
			},
		})
	}
	if vehicleType != "" {
		conditions = append(conditions, bson.M{"vehicleType": bson.M{"$elemMatch": bson.M{"$regex": vehicleType, "$options": "i"}}})
	}

	if status != "" {
		conditions = append(conditions, bson.M{"status": status})
	}

	if kycStatus != "" {
		kycStatusLower := strings.ToLower(kycStatus)
		switch kycStatusLower {
		case "active", "verified":
			conditions = append(conditions, bson.M{"status": domain.StatusActive})
		case "pending":
			conditions = append(conditions, bson.M{"status": domain.StatusPending})
		case "rejected":
			conditions = append(conditions, bson.M{"status": domain.StatusRejected})
		}
	}

	if accountStatus != "" {
		accountStatusLower := strings.ToLower(accountStatus)
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

	log.Println("Final Query:", query)

	providers, total, err := s.providers.FindAll(ctx, query, skip, int64(limit), sort)
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

	formattedProviders := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, p.ID.Hex())
		if err != nil {
			log.Printf("Error getting service stats for provider %s: %v", p.ID, err)
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

		formattedProviders[i] = ProviderResponse{
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

	totalPages := 1
	if total > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &ProviderListResponse{
		Providers: formattedProviders,
		Counts: ProviderCounts{
			Total:      total,
			Active:     activeCount,
			Inactive:   inactiveCount,
			PendingKYC: pendingKycCount,
		},
		Pagination: Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			Total:       total,
			Limit:       limit,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}

func (s *ProviderAdminService) GetProviderByID(ctx context.Context, id string) (*ProviderDetailResponse, error) {
	provider, err := s.providers.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, provider.ID.Hex())
	if err != nil {
		log.Printf("Error getting service stats: %v", err)
		totalJobs = 0
		completedJobs = 0
	}

	recentServices, err := s.services.FindByProviderID(ctx, provider.ID.Hex(), 10)
	if err != nil {
		log.Printf("Error fetching recent services: %v", err)
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

	return &ProviderDetailResponse{
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

func (s *ProviderAdminService) VerifyDocument(
	ctx context.Context,
	id, documentType, documentID, action string,
) (*domain.Provider, error) {

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

func (s *ProviderAdminService) GetDocumentURL(
	ctx context.Context,
	providerID, documentType, documentID string,
) (*DocumentResponse, error) {
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
					return &DocumentResponse{
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
		return &DocumentResponse{
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
					return &DocumentResponse{
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
		return &DocumentResponse{
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

		return &DocumentResponse{
			DocumentType: "cancel_cheque",
			File:         provider.CancelCheque.File,
			Verified:     provider.CancelCheque.Verified,
		}, nil

	default:
		return nil, fmt.Errorf("invalid document type. Use: identity, address, or cancel_cheque")
	}
}

func (s *ProviderAdminService) AddNote(
	ctx context.Context,
	providerID string,
	req AddNoteRequest,
) error {
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

func (s *ProviderAdminService) UpdateProvider(
	ctx context.Context,
	id string,
	req dto.UpdateProviderRequest,
	updatedBy primitive.ObjectID,
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

func (s *ProviderAdminService) GetZoneStats(ctx context.Context, adminZones map[string][]string, adminID primitive.ObjectID) (*ZoneStatsResponse, error) {
    zoneNames := adminZones["zoneName"]

    if len(zoneNames) == 0 {
        return &ZoneStatsResponse{
            Zones: []domain.ZoneStats{},
            TotalZones: 0,
            TotalProviders: 0,
            TotalActivationMembers: 0,
        }, nil
    }

    results := make([]domain.ZoneStats, 0, len(zoneNames))
    
    totalProvidersAcrossZones := 0
    totalActivationMembersAcrossZones := 0

    for _, zoneName := range zoneNames {
        subAdminIDs, err := s.getSubAdminsByZone(ctx, zoneName)
        if err != nil {
            results = append(results, domain.ZoneStats{
                ZoneName:               zoneName,
                TotalProviders:         0,
                TotalActivationMembers: 0,
            })
            continue
        }

        totalActivationMembers := len(subAdminIDs)
        totalActivationMembersAcrossZones += totalActivationMembers

        if totalActivationMembers == 0 {
            results = append(results, domain.ZoneStats{
                ZoneName:               zoneName,
                TotalProviders:         0,
                TotalActivationMembers: 0,
            })
            continue
        }

        totalProviders, err := s.countProvidersByCreators(ctx, subAdminIDs)
        if err != nil {
            results = append(results, domain.ZoneStats{
                ZoneName:               zoneName,
                TotalProviders:         0,
                TotalActivationMembers: totalActivationMembers,
            })
            continue
        }

        totalProvidersAcrossZones += totalProviders

        results = append(results, domain.ZoneStats{
            ZoneName:               zoneName,
            TotalProviders:         totalProviders,
            TotalActivationMembers: totalActivationMembers,
        })
    }

    return &ZoneStatsResponse{
        Zones:                    results,
        TotalZones:               len(results),
        TotalProviders:           totalProvidersAcrossZones,
		NewlyActivatedProviders:   totalProvidersAcrossZones,
        TotalActivationMembers:   totalActivationMembersAcrossZones,
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

func (s *ProviderAdminService) GetZoneActivationTeam(
	ctx context.Context,
	zoneName string,
	adminZones map[string][]string,
) (*ActivationTeamResponse, error) {

	roleIDs, err := s.role.FindRoleIDsByZoneName(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %v", err)
	}

	if len(roleIDs) == 0 {
		return &ActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []domain.ActivationTeamMember{},
		}, nil
	}

	adminIDs, err := s.admin.FindAdminIDsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find admins: %v", err)
	}

	if len(adminIDs) == 0 {
		return &ActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []domain.ActivationTeamMember{},
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
		return &ActivationTeamResponse{
			ZoneName:        zoneName,
			TotalProviders:  0,
			TotalActivators: 0,
			Team:            []domain.ActivationTeamMember{},
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

	return &ActivationTeamResponse{
		ZoneName:        zoneName,
		TotalProviders:  int64(totalProviders),
		TotalActivators: int64(len(subAdminIDs)),
		Team:            team,
	}, nil
}

func (s *ProviderAdminService) getActivationTeamDetails(
	ctx context.Context,
	adminIDs []primitive.ObjectID,
	zoneName string,
) ([]domain.ActivationTeamMember, error) {

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

func (s *ProviderAdminService) GetActivationPersonProviders(
	ctx context.Context,
	personID string,
	zoneName string,
	pageStr, limitStr, sort, search string,
	adminZones map[string][]string,
) (*ProviderListResponse, error) {

	
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

	
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	skip := int64((page - 1) * limit)

	conditions := []bson.M{
		{"createdBy": adminObjectID},
	}

	if search != "" {
		searchConditions := []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		}

		conditions = append(conditions, bson.M{"$or": searchConditions})
	}

	query := bson.M{"$and": conditions}

	providers, total, err := s.providers.FindAll(
		ctx,
		query,
		skip,
		int64(limit),
		sort,
	)
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

	formatted := make([]ProviderResponse, len(providers))
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

		formatted[i] = ProviderResponse{
			ID:            p.ID.Hex(),
			ProviderID:    providerIDStr,
			Name:          defaultStr(p.Name, "N/A"),
			Mobile:        p.Phone,
			Email:         defaultStr(p.Email, "N/A"),
			Zone: p.City,
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

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ProviderListResponse{
		Providers: formatted,
		Counts: ProviderCounts{
			Total:      total,
			Active:     activeCount,
			Inactive:   inactiveCount,
			PendingKYC: pendingKycCount,
		},
		Pagination: Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			Total:       total,
			Limit:       limit,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}
