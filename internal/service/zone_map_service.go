package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
)

type ZoneMapService struct {
	providers *repository.ProviderRepo
	admins    *repository.AdminRepository
	roles     *repository.RoleRepository
	services  *repository.AcceptedServiceRepo
}

func NewZoneMapService(
	providers *repository.ProviderRepo,
	admins *repository.AdminRepository,
	roles *repository.RoleRepository,
	services *repository.AcceptedServiceRepo,
) *ZoneMapService {
	return &ZoneMapService{
		providers: providers,
		admins:    admins,
		roles:     roles,
		services:  services,
	}
}

type ZoneStatsResponse struct {
	ZoneName           string `json:"zoneName"`
	ActivationTeam     int    `json:"activationTeam"`
	TotalProviders     int64  `json:"totalProviders"`
}

type ActivationTeamMember struct {
	ActivationPersonName      string `json:"activationPersonName"`
	AssignZone                string `json:"assignZone"`
	TotalActivatedProviders   int64  `json:"totalActivatedProviders"`
}

type AdminWithRole struct {
	Admin domain.Admin
	Role  domain.Role
}

func (s *ZoneMapService) ValidateAdminAccess(ctx context.Context, adminID string) (*AdminWithRole, error) {
	objID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return nil, errors.New("invalid admin ID")
	}

	admin, err := s.admins.FindByID(ctx, objID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if admin.Status != "active" {
		return nil, errors.New("admin account is not active")
	}

	role, err := s.roles.FindByID(ctx, admin.RoleID)
	if err != nil {
		return nil, errors.New("role not found")
	}

	return &AdminWithRole{
		Admin: *admin,
		Role:  *role,
	}, nil
}

func (s *ZoneMapService) GetZoneStats(
	ctx context.Context,
	adminID string,
) ([]domain.ZoneStats, error) {

	// 1️⃣ Validate admin access
	adminWithRole, err := s.ValidateAdminAccess(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if adminWithRole.Role.RoleType != "admin" {
		return nil, errors.New("only admin can access zone statistics")
	}

	// 2️⃣ Zone scope filtering
	matchStage := bson.M{}
	if len(adminWithRole.Role.ZoneScope) > 0 {
		matchStage["city"] = bson.M{
			"$in": adminWithRole.Role.ZoneScope,
		}
	}

	pipeline := []bson.M{
		{
			"$match": matchStage,
		},
		{
			"$group": bson.M{
				"_id":            "$city",
				"totalProviders": bson.M{"$sum": 1},
				"activationTeam": bson.M{"$addToSet": "$createdBy"},
			},
		},
		{
			"$project": bson.M{
				"zoneName":       "$_id",
				"totalProviders": 1,
				"activationTeam": bson.M{"$size": "$activationTeam"},
				"_id":            0,
			},
		},
	}

	return s.providers.AggregateZoneStats(ctx, pipeline)
}

func (s *ZoneMapService) GetActivationTeam(
    ctx context.Context,
    adminID string,
    zoneName string,
) ([]domain.ActivationTeamMember, error) {

    adminWithRole, err := s.ValidateAdminAccess(ctx, adminID)
    if err != nil {
        return nil, err
    }

    if adminWithRole.Role.RoleType != "admin" {
        return nil, errors.New("only admin can access activation team")
    }

    if len(adminWithRole.Role.ZoneScope) > 0 &&
        !contains(adminWithRole.Role.ZoneScope, zoneName) {
        return nil, errors.New("access denied for this zone")
    }

    pipeline := []bson.M{
        {"$match": bson.M{"city": zoneName}},
        {"$group": bson.M{
            "_id":            "$createdBy",
            "totalActivated": bson.M{"$sum": 1},
        }},
        {"$project": bson.M{
            "activationPersonName":    "$_id",
            "assignZone":              zoneName,
            "totalActivatedProviders": "$totalActivated",
            "_id":                     0,
        }},
    }

    return s.providers.AggregateActivationTeam(ctx, pipeline)
}


func (s *ZoneMapService) GetProvidersByActivator(
	ctx context.Context,
	adminID string,
	zoneName string,
	activatorName string,
	pageStr string,
	limitStr string,
	sort string,
) (*ProviderListResponse, error) {
	adminWithRole, err := s.ValidateAdminAccess(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if adminWithRole.Role.RoleType != "admin" {
		return nil, errors.New("only admin can access this view")
	}

	if len(adminWithRole.Role.ZoneScope) > 0 && !contains(adminWithRole.Role.ZoneScope, zoneName) {
		return nil, errors.New("access denied for this zone")
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

	query := bson.M{
		"createdBy": activatorName,
		"city":      zoneName,
	}

	if len(adminWithRole.Role.ZoneScope) > 0 {
		query["city"] = bson.M{"$in": adminWithRole.Role.ZoneScope}
	}

	providers, total, err := s.providers.FindAll(ctx, query, skip, int64(limit), sort)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %v", err)
	}

	formattedProviders := make([]ProviderResponse, len(providers))
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

func (s *ZoneMapService) GetMyProviders(
	ctx context.Context,
	adminID string,
	pageStr string,
	limitStr string,
	sort string,
) (*ProviderListResponse, error) {
	adminWithRole, err := s.ValidateAdminAccess(ctx, adminID)
	if err != nil {
		return nil, err
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

	query := bson.M{"createdBy": adminWithRole.Admin.Name}

	if len(adminWithRole.Role.ZoneScope) > 0 {
		query["city"] = bson.M{"$in": adminWithRole.Role.ZoneScope}
	}

	providers, total, err := s.providers.FindAll(ctx, query, skip, int64(limit), sort)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %v", err)
	}

	formattedProviders := make([]ProviderResponse, len(providers))
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}