package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
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
	kycRepo  *repository.ProviderKYCRepository
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
	k *repository.ProviderKYCRepository,
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
		kycRepo: k,
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


func (s *ProviderAdminService) GetAllProviders(
	ctx context.Context,
	filters dto.ProviderFilters,
	pagination dto.ProviderPagination,
	zoneFilter bson.M,
) (*dto.ProviderListResponse, error) {

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 20
	}
	pagination.Skip = (pagination.Page - 1) * pagination.Limit

	search := strings.TrimSpace(filters.Search)

	var conditions []bson.M

	if len(zoneFilter) > 0 {
		conditions = append(conditions, zoneFilter)
	}

	if search != "" {
		conditions = append(conditions, bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": search, "$options": "i"}},
				{"phone": bson.M{"$regex": search}},
				{"providerCode": bson.M{"$regex": search, "$options": "i"}},
				{"city": bson.M{"$regex": search, "$options": "i"}},
				{"address": bson.M{"$regex": search, "$options": "i"}},
				{"vehicleType": bson.M{"$regex": search, "$options": "i"}},
			},
		})
	}

	switch strings.ToLower(filters.AccountStatus) {
	case "active":
		conditions = append(conditions, bson.M{
			"isActive": domain.AccountStatusActive,
		})
	case "inactive":
		conditions = append(conditions, bson.M{
			"isActive": bson.M{
				"$in": []string{
					domain.AccountStatusSuspended,
					domain.AccountStatusBlacklisted,
					domain.AccountStatusDeactivated,
				},
			},
		})
	case "suspended":
		conditions = append(conditions, bson.M{
			"isActive": domain.AccountStatusSuspended,
		})
	case "blacklisted":
		conditions = append(conditions, bson.M{
			"isActive": domain.AccountStatusBlacklisted,
		})
	}

	if filters.VehicleType != "" {
		conditions = append(conditions, bson.M{"vehicleType": bson.M{"$elemMatch": bson.M{"$regex": filters.VehicleType, "$options": "i"}}})
	}

	if filters.KYCStatus != "" {

		kycStatus := strings.ToUpper(filters.KYCStatus)
	
		kycs, err := s.kycRepo.Find(ctx, bson.M{
			"status": kycStatus,
		})
	
		if err != nil {
			return nil, fmt.Errorf("failed to filter by kyc status: %w", err)
		}
	
		var kycIDs []primitive.ObjectID
		for _, k := range kycs {
			kycIDs = append(kycIDs, k.ID)
		}
	
		if len(kycIDs) > 0 {
			conditions = append(conditions, bson.M{
				"kycId": bson.M{"$in": kycIDs},
			})
		} else {
			conditions = append(conditions, bson.M{
				"_id": primitive.NilObjectID,
			})
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

	if filters.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			conditions = append(conditions, bson.M{
				"createdAt": bson.M{
					"$gte": primitive.NewDateTimeFromTime(t),
					"$lte": primitive.NewDateTimeFromTime(t.Add(24*time.Hour - time.Second)),
				},
			})
		}
	}

	query := bson.M{}
	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	providers, total, err := s.providers.FindAll(
		ctx,
		query,
		pagination.Skip,
		pagination.Limit,
		pagination.Sort,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %w", err)
	}

	var kycIDs []primitive.ObjectID
	for _, p := range providers {
		if p.KYCID != primitive.NilObjectID {
			kycIDs = append(kycIDs, p.KYCID)
		}
	}

	kycMap := make(map[primitive.ObjectID]domain.ProviderKYC)
	if len(kycIDs) > 0 {
		kycs, _ := s.kycRepo.FindByIDs(ctx, kycIDs)
		for _, k := range kycs {
			kycMap[k.ID] = k
		}
	}

	formatted := make([]dto.ProviderAllResponse, 0, len(providers))

	for _, p := range providers {

		totalJobs, completedJobs, _ := s.services.GetServiceStats(ctx, p.ID.Hex())

		kycStatus := "Not Submitted"
		if k, ok := kycMap[p.KYCID]; ok {
			switch k.Status {
			case domain.KYC_APPROVED:
				kycStatus = "Verified"
			case domain.KYC_REJECTED:
				kycStatus = "Rejected"
			case domain.KYC_PENDING:
				kycStatus = "Pending"
			}
		}

		vehicle := "N/A"
		if len(p.VehicleType) > 0 {
			vehicle = strings.Join(p.VehicleType, ", ")
		}

		formatted = append(formatted, dto.ProviderAllResponse{
			ID:            p.ID.Hex(),
			ProviderID:    p.ProviderCode,
			Name:          defaultStr(p.Name, "N/A"),
			Mobile:        p.Phone,
			Email:         defaultStr(p.Email, "N/A"),
			KYC:           kycStatus,
			Account:       string(p.IsActive),
			Vehicle:       vehicle,
			Zone:          p.City,
			DOJ:           formatDate(p.CreatedAt),
			ProfileURL:    p.ProfileURL,
			IsServiceOn:   p.IsServiceOn,
			IsActive:      p.IsActive,
			TotalJobs:     totalJobs,
			CompletedJobs: completedJobs,
		})
	}

	activeCount, _ := s.providers.Count(ctx, bson.M{
		"isActive": domain.AccountStatusActive,
	})

	inactiveCount, _ := s.providers.Count(ctx, bson.M{
		"isActive": bson.M{
			"$in": []string{
				domain.AccountStatusSuspended,
				domain.AccountStatusBlacklisted,
				domain.AccountStatusDeactivated,
			},
		},
	})

	pendingKycCount, _ := s.kycRepo.Count(ctx, bson.M{
		"status": domain.KYC_PENDING,
	})

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

	var (
		kycStatus   = "Not Submitted"
		kycDocs     []domain.KYCDocument
		bankDetails *domain.ProviderBankDetails
		approvedAt  string
	)

	if provider.KYCID != primitive.NilObjectID {

		kyc, err := s.kycRepo.FindByID(ctx, provider.KYCID)
		if err != nil {
			fmt.Printf("Error fetching KYC: %v\n", err)
		} else {

			switch kyc.Status {
			case domain.KYC_APPROVED:
				kycStatus = "Verified"
				approvedAt = formatDateDetailed(kyc.ApprovedAt)
			case domain.KYC_REJECTED:
				kycStatus = "Rejected" 
			case domain.KYC_PENDING:
				kycStatus = "Pending"
			}

			kycDocs = kyc.Documents
			bankDetails = &kyc.Bank
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

	return &dto.ProviderDetailResponse{
		ID:                   provider.ID.Hex(),
		ProviderID:           provider.ProviderCode,
		Name:                 defaultStr(provider.Name, "N/A"),
		Phone:                provider.Phone,
		Email:                defaultStr(provider.Email, "N/A"),
		AlternateContact:     provider.AlternateContact,
		ProfileURL:           provider.ProfileURL,
		Address:              provider.Address,
		PermanentAddress:     provider.PermanentAddress,
		City:                 provider.City,
		Account:              accountStatus,
		KYCID:                provider.KYCID,
		KYCStatus:            kycStatus,
		KYCDocuments:         kycDocs,
		BankDetails:          bankDetails,
		VehicleType:          provider.VehicleType,
		VehicleNumber:        provider.VehicleNumber,
		ProviderBrands:       provider.ProviderBrands,
		ProviderServices:     provider.ProviderServices,
		CompanyName:          provider.CompanyName,
		Description:          provider.Description,
		Zone:                 defaultStr(provider.City, defaultStr(provider.Address, "N/A")),
		DOJ:                  formatDateDetailed(provider.CreatedAt),
		IsActive:             string(provider.IsActive),
		TotalJobs:            totalJobs,
		CompletedJobs:        completedJobs,
		CommissionPercentage: provider.CommissionPercentage,
		Notes:                provider.Notes,
		ApprovedAt:           approvedAt,
		AgreementSubmittedAt:  provider.AgreementSubmittedAt,
		IsAgreementSubmitted: provider.IsAgreementSubmitted,
	}, nil
}

func (s *ProviderAdminService) UpdateProviderStatus(ctx context.Context, id, status string) (*domain.Provider, error) {

	validStatuses := map[string]bool{
		domain.StatusActive:   true,
		domain.StatusPending:  true,
		domain.StatusRejected: true,
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

func (s *ProviderAdminService) VerifyDocument( ctx context.Context, kycID, documentID, action string ) (*domain.ProviderKYC, error) {

	if action != "approve" && action != "reject" {
		return nil, fmt.Errorf("invalid action (approve/reject required)")
	}

	verificationStatus := domain.VERIFICATION_APPROVED
	if action == "reject" {
		verificationStatus = domain.VERIFICATION_REJECTED
	}

	kyc, err := s.kycRepo.UpdateDocumentVerification(ctx, kycID, documentID, verificationStatus)
	if err != nil {
		return nil, err
	}

	allApproved := true
	hasRejected := false

	for _, doc := range kyc.Documents {
		if doc.Verified == domain.VERIFICATION_REJECTED {
			hasRejected = true
			break
		}
		if doc.Verified != domain.VERIFICATION_APPROVED {
			allApproved = false
		}
	}

	var newStatus domain.KYCStatus
	if hasRejected {
		newStatus = domain.KYC_REJECTED
	} else if allApproved {
		newStatus = domain.KYC_APPROVED
	} else {
		newStatus = domain.KYC_PENDING
	}

	if kyc.Status != newStatus {
		kyc, err = s.kycRepo.UpdateKYCStatus(ctx, kycID, newStatus)
		if err != nil {
			return nil, err
		}
	}

	return kyc, nil
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
	if t.IsZero() {
		return "N/A"
	}
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
func (s *ProviderAdminService) CreateProvider(ctx context.Context, req dto.CreateProviderRequest, documents []domain.KYCDocument,createdBy primitive.ObjectID, role string) (*domain.Provider, error) {

	existingProvider, err := s.providers.FindByPhone(ctx, req.Phone)
	if err == nil && existingProvider != nil {
		return nil, errors.New("phone number already exists")
	}

	now := time.Now()
	providerID := time.Now().UnixNano() / 1000000

	provider := &domain.Provider{
		ID:               primitive.NewObjectID(),
		ProviderCode:     "PRVH" + strconv.FormatInt(providerID, 10),
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
		Description:      req.Description,
		ProfileURL:       req.ProfileURL,
		IsActive:         domain.AccountStatusActive,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.providers.Create(ctx, provider); err != nil {
		return nil, err
	}

	kyc := &domain.ProviderKYC{
		ID:         primitive.NewObjectID(),
		ProviderID: provider.ID,
		Documents:  documents,
		Status:     domain.KYC_PENDING,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	
	if req.AccountHolderName != "" {
		kyc.Bank = domain.ProviderBankDetails{
			AccountHolderName: req.AccountHolderName,
			AccountNumber:     req.AccountNumber,
			IFSC:              req.IfscCode,
			BranchName:        req.BranchName,
			UPIID:             req.Upi,
			GSTNumber:         req.GSTNumber,
		}
	}

	if err := s.kycRepo.Create(ctx, kyc); err != nil {
		return nil, err
	}

	if err := s.providers.UpdateKYCID(ctx, provider.ID, kyc.ID); err != nil {
		return nil, err
	}

	return provider, nil
}


func (s *ProviderAdminService) UpdateProvider(
    ctx context.Context,
    id string,
    req dto.UpdateProviderRequest,
    documents []domain.KYCDocument,
    updatedBy primitive.ObjectID,
) (*domain.Provider, error) {

    provider, err := s.providers.FindByID(ctx, id)
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
    if len(req.ProviderBrands) > 0 {
        update["providerBrands"] = req.ProviderBrands
    }
    if len(req.ProviderServices) > 0 {
        update["providerServices"] = req.ProviderServices
    }

    updatedProvider, err := s.providers.Update(ctx, id, update)
    if err != nil {
        return nil, err
    }

    if len(documents) > 0 || req.AccountHolderName != "" {

        var existingKYC *domain.ProviderKYC
        existingKYC, err = s.kycRepo.FindByProviderID(ctx, provider.ID)
        if err != nil && err != mongo.ErrNoDocuments {
            return nil, err
        }

        kycUpdate := bson.M{
            "updatedAt": time.Now(),
            "status":    domain.KYC_PENDING,
        }

        if len(documents) > 0 {

            var existingDocs []domain.KYCDocument
            if existingKYC != nil {
                existingDocs = existingKYC.Documents
            }

            mergedDocs := mergeKYCDocuments(existingDocs, documents)
            kycUpdate["documents"] = mergedDocs
        }

        if req.AccountHolderName != "" {
            kycUpdate["bank"] = domain.ProviderBankDetails{
                AccountHolderName: req.AccountHolderName,
                AccountNumber:     req.AccountNumber,
                IFSC:              req.IfscCode,
                BranchName:        req.BranchName,
                UPIID:             req.Upi,
                GSTNumber:         req.GSTNumber,
            }
        }

        if err := s.kycRepo.UpdateByProviderID(ctx, provider.ID, kycUpdate); err != nil {
            return nil, err
        }
    }

    return updatedProvider, nil
}

func mergeKYCDocuments(existing, incoming []domain.KYCDocument) []domain.KYCDocument {
    docMap := make(map[domain.DocumentType]domain.KYCDocument)

    for _, d := range existing {
        docMap[d.Type] = d
    }

    for _, d := range incoming {
        docMap[d.Type] = d
    }

    result := make([]domain.KYCDocument, 0, len(docMap))
    for _, v := range docMap {
        result = append(result, v)
    }

    return result
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

		formatted[i] =dto.ProviderAllResponse{
			ID:            p.ID.Hex(),
			ProviderID:    p.ProviderCode,
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
			IsActive:      string(p.IsActive),
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
			bookingID = service.ServiceNumber
			serviceTime := service.CreatedAt.Add(5*time.Hour + 30*time.Minute)
			bookingDate = serviceTime.Format("2006-01-02 15:04:05")
			paymentStatus = string(settlement.SettlementStatus)
           
			serviceReq, err := s.serviceRequestRepo.FindByID(ctx, service.ID.Hex())
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