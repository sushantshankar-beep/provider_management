package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Zone struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ZoneName  string             `bson:"zoneName" json:"zoneName"`
	StateName string             `bson:"stateName" json:"stateName"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}