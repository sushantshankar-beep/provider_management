package domain

type ActivationTeamMember struct {
    ActivationPersonName    string `bson:"activationPersonName" json:"activationPersonName"`
    AssignZone              string `bson:"assignZone" json:"assignZone"`
    TotalActivatedProviders int64  `bson:"totalActivatedProviders" json:"totalActivatedProviders"`
}
