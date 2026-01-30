package domain

import "go.mongodb.org/mongo-driver/bson/primitive"
type ActivationTeamMember struct {
	PersonID        primitive.ObjectID `json:"personId" bson:"_id"`
	PersonName      string             `json:"personName" bson:"personName"`
	AssignZone      string             `json:"assignZone" bson:"assignZone"`
	TotalProviders  int64              `json:"totalProviders" bson:"totalProviders"`
}
