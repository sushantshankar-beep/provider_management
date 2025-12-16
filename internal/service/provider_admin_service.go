package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"provider_management/internal/constants"
	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
)

type ProviderAdminService struct {
	providers *repository.ProviderRepo
	services  *repository.AcceptedServiceRepo
}

func NewProviderAdminService(
	p *repository.ProviderRepo,
	s *repository.AcceptedServiceRepo,
) *ProviderAdminService {
	return &ProviderAdminService{
		providers: p,
		services:  s,
	}
}

type ProviderListResponse struct {
	Providers  []ProviderResponse `json:"providers"`
	Counts     ProviderCounts     `json:"counts"`
	Pagination Pagination         `json:"pagination"`
}

type ProviderResponse struct {
	ID               string   `json:"id"`
	ProviderID       string   `json:"provider_id"`
	Name             string   `json:"name"`
	Mobile           string   `json:"mobile"`
	Email            string   `json:"email"`
	KYC              string   `json:"kyc"`
	Account          string   `json:"account"`
	Vehicle          string   `json:"vehicle"`
	Zone             string   `json:"zone"`
	DOJ              string   `json:"doj"`
	ProfileURL       string   `json:"profile_url"`
	IsServiceOn      bool     `json:"is_service_on"`
	IsActive         string   `json:"is_active"`
	Status           string   `json:"status"`
	TotalJobs        int64    `json:"total_jobs"`
	CompletedJobs    int64    `json:"completed_jobs"`
}

type ProviderCounts struct {
	Total       int64 `json:"total"`
	Active      int64 `json:"active"`
	Suspended   int64 `json:"suspended"`
	Blacklisted int64 `json:"blacklisted"`
	PendingKYC  int64 `json:"pending_kyc"`
	ActiveKYC   int64 `json:"active_kyc"`
	RejectedKYC int64 `json:"rejected_kyc"`
}

type ProviderDetailResponse struct {
	ID                   string               `json:"id"`
	ProviderID           string               `json:"provider_id"`
	Name                 string               `json:"name"`
	Phone                string               `json:"phone"`
	Email                string               `json:"email"`
	AlternateContact     string               `json:"alternate_contact"`
	ProfileURL           string               `json:"profile_url"`
	Address              string               `json:"address"`
	PermanentAddress     string               `json:"permanent_address"`
	City                 string               `json:"city"`
	Status               string               `json:"status"`
	Account              string               `json:"account"`
	KYC                  string               `json:"kyc"`
	VehicleType          []string             `json:"vehicle_type"`
	VehicleNumber        string               `json:"vehicle_number"`
	ProviderBrands       []string             `json:"provider_brands"`
	ProviderServices     []string             `json:"provider_services"`
	GSTNumber            string               `json:"gst_number"`
	CompanyName          string               `json:"company_name"`
	Description          string               `json:"description"`
	Zone                 string               `json:"zone"`
	DOJ                  string               `json:"doj"`
	IdentityProof        []domain.Proof       `json:"identity_proof"`
	AddressProof         []domain.Proof       `json:"address_proof"`
	CancelCheque         *domain.CancelCheque `json:"cancel_cheque"`
	BankDetails          *domain.BankDetails  `json:"bank_details"`
	IsServiceOn          bool                 `json:"is_service_on"`
	IsActive             string               `json:"is_active"`
	IsSocketConnected    bool                 `json:"is_socket_connected"`
	Location             *domain.GeoPoint     `json:"location"`
	TotalJobs            int64                `json:"total_jobs"`
	CompletedJobs        int64                `json:"completed_jobs"`
	RecentServices       []domain.Service     `json:"recent_services"`
	CommissionPercentage float64              `json:"commission_percentage,omitempty"`
}

func (s *ProviderAdminService) GetAllProviders(
	ctx context.Context,
	pageStr, limitStr, sort, search, status, name, mobile,
	providerID, kycStatus, accountStatus, vehicleType, zone string,
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

	query := bson.M{}
	var conditions []bson.M

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

	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	log.Printf("Final query: %+v", query)
	log.Printf("Pagination: page=%d, limit=%d, skip=%d, sort=%s", page, limit, skip, sort)

	providers, total, err := s.providers.FindAll(ctx, query, skip, int64(limit), sort)
	if err != nil {
		log.Printf("Error in FindAll: %v", err)
		return nil, fmt.Errorf("failed to fetch providers: %v", err)
	}

	log.Printf("Found %d providers out of total %d", len(providers), total)

	countQuery := query
	if andConditions, ok := countQuery["$and"].([]bson.M); ok && len(andConditions) == 1 {
		for k, v := range andConditions[0] {
			countQuery[k] = v
		}
		delete(countQuery, "$and")
	}

	activeCount, _ := s.providers.CountByStatus(ctx, countQuery, "isActive", domain.AccountStatusActive)
	suspendedCount, _ := s.providers.CountByStatus(ctx, countQuery, "isActive", domain.AccountStatusSuspended)
	blacklistedCount, _ := s.providers.CountByStatus(ctx, countQuery, "isActive", domain.AccountStatusBlacklisted)
	pendingKycCount, _ := s.providers.CountByStatus(ctx, countQuery, "status", domain.StatusPending)
	activeKycCount, _ := s.providers.CountByStatus(ctx, countQuery, "status", domain.StatusActive)
	rejectedKycCount, _ := s.providers.CountByStatus(ctx, countQuery, "status", domain.StatusRejected)

	formattedProviders := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, p.ID)
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
			ID:               p.ID,
			ProviderID:       providerIDStr,
			Name:             defaultStr(p.Name, "N/A"),
			Mobile:           p.Phone,
			Email:            defaultStr(p.Email, "N/A"),
			KYC:              kyc,
			Account:          account,
			Vehicle:          vehicle,
			Zone:             defaultStr(p.City, defaultStr(p.Address, "N/A")),
			DOJ:              formatDate(p.CreatedAt),
			ProfileURL:       p.ProfileURL,
			IsServiceOn:      p.IsServiceOn,
			IsActive:         string(p.IsActive),
			Status:           p.Status,
			TotalJobs:        totalJobs,
			CompletedJobs:    completedJobs,
		}
	}

	totalPages := 1
	if total > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &ProviderListResponse{
		Providers: formattedProviders,
		Counts: ProviderCounts{
			Total:       total,
			Active:      activeCount,
			Suspended:   suspendedCount,
			Blacklisted: blacklistedCount,
			PendingKYC:  pendingKycCount,
			ActiveKYC:   activeKycCount,
			RejectedKYC: rejectedKycCount,
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

	totalJobs, completedJobs, err := s.services.GetServiceStats(ctx, provider.ID)
	if err != nil {
		log.Printf("Error getting service stats: %v", err)
		totalJobs = 0
		completedJobs = 0
	}

	recentServices, err := s.services.FindByProviderID(ctx, provider.ID, 10)
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
		ID:                   provider.ID,
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
