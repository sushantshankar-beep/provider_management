package domain

type ZoneStats struct {
	ZoneName       string `bson:"zoneName" json:"zoneName"`
	ActivationTeam int    `bson:"activationTeam" json:"activationTeam"`
	TotalProviders int64  `bson:"totalProviders" json:"totalProviders"`
}
