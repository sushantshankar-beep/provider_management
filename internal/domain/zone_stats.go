package domain

type ZoneStats struct {
	ZoneName                 string `bson:"zoneName" json:"zoneName"`
	TotalProviders           int  `bson:"totalProviders" json:"totalProviders"`
	TotalActivationMembers   int  `bson:"totalActivationMembers" json:"totalActivationMembers"`
}