package service

import (
	"context"
	"errors"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"time"
)

type ZoneService struct {
	zones repository.ZoneRepo
}

func NewZoneService(z *repository.ZoneRepo) *ZoneService {
	return &ZoneService{
		zones: *z,
	}
}

type ZoneStats struct {
	TotalZones       int64 `json:"totalZones"`
	ActiveZones      int64 `json:"activeZones"`
	DeactivateZones  int64 `json:"deactivateZones"`
}

func (s *ZoneService) CreateZone(ctx context.Context, zoneName, stateName string, isActive *bool) (*domain.Zone, error) {
	existing, _ := s.zones.FindByName(ctx, zoneName)
	if existing != nil {
		return nil, errors.New("zone already exists")
	}

	active := true
	if isActive != nil {
		active = *isActive
	}

	zone := &domain.Zone{
		ZoneName:  zoneName,
		StateName: stateName,
		IsActive:  active,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.zones.Create(ctx, zone)
	if err != nil {
		return nil, err
	}

	return zone, nil
}

func (s *ZoneService) ListZones(ctx context.Context, pageStr, limitStr, search, isActiveStr,state, createdAtStr, updatedAtStr string) ([]domain.Zone, ZoneStats, int64, error) {
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	var isActivePtr *bool
	if isActiveStr != "" {
		val := isActiveStr == "true"
		isActivePtr = &val
	}

	var createdAtPtr, updatedAtPtr *time.Time

	if createdAtStr != "" {
		if t, err := time.Parse("02/01/2006", createdAtStr); err == nil {
			createdAtPtr = &t
		}
	}

	if updatedAtStr != "" {
		if t, err := time.Parse("02/01/2006", updatedAtStr); err == nil {
			updatedAtPtr = &t
		}
	}


	zones, total, err := s.zones.FindWithFilter(ctx, skip, limit, search, isActivePtr, state ,createdAtPtr,
		updatedAtPtr,)
	if err != nil {
		return nil, ZoneStats{}, 0, err
	}

	activeCount, _ := s.zones.CountActive(ctx, search, true)
	inactiveCount, _ := s.zones.CountActive(ctx, search, false)

	stats := ZoneStats{
		TotalZones:      total,
		ActiveZones:     activeCount,
		DeactivateZones: inactiveCount,
	}

	return zones, stats, total, nil
}

func (s *ZoneService) GetActiveZones(ctx context.Context, state string) ([]map[string]interface{}, error) {
	zones, err := s.zones.FindActive(ctx, state)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(zones))
	for i, z := range zones {
		result[i] = map[string]interface{}{
			"_id":      z.ID,
			"zoneName": z.ZoneName,
		}
	}

	return result, nil
}

func (s *ZoneService) UpdateZone(ctx context.Context, id string, zoneName, stateName *string, isActive *bool) (*domain.Zone, error) {
	zone, err := s.zones.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("zone not found")
	}

	if zoneName != nil {
		zone.ZoneName = *zoneName
	}
	if stateName != nil {
		zone.StateName = *stateName
	}
	if isActive != nil {
		zone.IsActive = *isActive
	}

	zone.UpdatedAt = time.Now()

	err = s.zones.Update(ctx, id, zone)
	if err != nil {
		return nil, err
	}

	return zone, nil
}

func (s *ZoneService) ToggleZoneStatus(ctx context.Context, id string) (*domain.Zone, error) {
	zone, err := s.zones.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("zone not found")
	}

	zone.IsActive = !zone.IsActive
	zone.UpdatedAt = time.Now()

	err = s.zones.Update(ctx, id, zone)
	if err != nil {
		return nil, err
	}

	return zone, nil
}

func (s *ZoneService) DeleteZone(ctx context.Context, id string) error {
	err := s.zones.Delete(ctx, id)
	if err != nil {
		return errors.New("zone not found")
	}

	return nil
}

func (s *ZoneService) GetActiveStates(ctx context.Context) ([]map[string]interface{}, error) {
	states, err := s.zones.FindActiveStates(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(states))
	for i, state := range states {
		result[i] = map[string]interface{}{
			"zoneState": state,
		}
	}

	return result, nil
}

