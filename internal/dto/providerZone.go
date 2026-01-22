package dto

type ProviderZonePagination struct {
	Page  int64
	Limit int64
	Sort  string
	Skip  int64
}

type ProviderZoneMetaPagination struct {
    CurrentPage int64  `json:"current_page"`
    TotalPages  int64  `json:"total_pages"`
    Total       int64 `json:"total"`
	TotalUsers  int64 `json:"total_users"`
    Limit       int64   `json:"limit"`
    HasNext     bool  `json:"has_next,omitempty"`
    HasPrev     bool  `json:"has_prev,omitempty"`
}

type ProviderZoneStatsResponse struct {
	Zones                   []ProviderZoneStats `json:"zones"`
	TotalZones              int                `json:"totalZones"`
	TotalProviders          int                `json:"totalProviders"`
	TotalActivationMembers  int                `json:"totalActivationMembers"`
	NewlyActivatedProviders int                `json:"newlyActivatedProviders"`
}

type ProviderZoneStats struct {
	ZoneName                 string `bson:"zoneName" json:"zoneName"`
	TotalProviders           int  `bson:"totalProviders" json:"totalProviders"`
	TotalActivationMembers   int  `bson:"totalActivationMembers" json:"totalActivationMembers"`
}
