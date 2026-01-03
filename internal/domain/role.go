package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --------------------
// Role Status
// --------------------
type RoleStatus string

const (
	RoleActive   RoleStatus = "active"
	RoleInactive RoleStatus = "inactive"
)

// --------------------
// Role Type
// --------------------
const (
	RoleTypeAdmin    = "admin"
	RoleTypeSubAdmin = "subAdmin"
)

// --------------------
// Zone Scope
// --------------------
// const (
// 	ZoneScopeAll      = "all"
// 	ZoneScopeAssigned = "assigned"
// )

// --------------------
// Role Model
// --------------------
type Role struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	Name        string              `bson:"name" json:"name"`
	RoleType    string              `bson:"roleType" json:"roleType"`
	Status      RoleStatus          `bson:"status" json:"status"`
	ZoneScope    []string            `bson:"zoneScope" json:"zoneScope"`
	Description string              `bson:"description,omitempty" json:"description,omitempty"`
	Permissions map[string][]string `bson:"permissions" json:"permissions"`
	CreatedBy   primitive.ObjectID  `bson:"createdBy" json:"createdBy"`
	CreatedAt   time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time           `bson:"updatedAt" json:"updatedAt"`
}
