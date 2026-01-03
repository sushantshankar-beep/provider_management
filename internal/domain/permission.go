package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Permission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Category    string             `bson:"category" json:"category"`
	Label       string             `bson:"label" json:"label"`
	Permissions []PermissionItem   `bson:"permissions" json:"permissions"`
	Order       int                `bson:"order" json:"order"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type PermissionItem struct {
	Key   string `bson:"key" json:"key"`
	Label string `bson:"label" json:"label"`
}