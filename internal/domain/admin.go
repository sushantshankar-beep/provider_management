package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AdminStatusActive   = "active"
	StatusDeactive = "deactive"
	RoleSubAdmin   = "subAdmin"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superAdmin"
	PowerLevelSubAdmin   = 1
	PowerLevelAdmin      = 2
	PowerLevelSuperAdmin = 3
)


var AllowedRoles = map[string]bool{
	RoleSubAdmin:   true,
	RoleAdmin:      true,
	RoleSuperAdmin: true,
}

func IsValidRole(role string) bool {
	_, ok := AllowedRoles[role]
	return ok
}

type Admin struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	Name          string               `bson:"name" json:"name"`
	AdminID       int64                `bson:"id" json:"id"`
	Email         string               `bson:"email" json:"email"`
	Phone         string               `bson:"phone" json:"phone"`
	Password      string               `bson:"password" json:"password,omitempty"`
	ProfileURL    string               `bson:"profileUrl" json:"profileUrl"`
	RoleID        primitive.ObjectID   `bson:"roleId,omitempty" json:"roleId,omitempty"`
	Role          string               `bson:"role" json:"role"`
	RoleName      string               `bson:"roleName" json:"roleName"`
	PowerLevel    int                  `bson:"powerLevel" json:"powerLevel"`
	ServiceZones  []string             `bson:"serviceZones" json:"serviceZones"`
	AccessModules []primitive.ObjectID `bson:"accessModules" json:"accessModules"`
	Status        string               `bson:"status" json:"status"`
	CreatedBy     primitive.ObjectID   `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	Token         string  `bson:"token,omitempty" json:"token,omitempty"`
	CreatedAt     time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time            `bson:"updatedAt" json:"updatedAt"`
}

type AdminResponse struct {
	ID            primitive.ObjectID   `json:"_id"`
	Name          string               `json:"name"`
	Email         string               `json:"email"`
	Phone         string               `json:"phone"`
	Role          string               `json:"role"`
	PowerLevel    int                  `json:"powerLevel"`
	ProfileURL    string               `json:"profile"`
	ServiceZones  []string             `json:"serviceZones"`
	AccessModules []primitive.ObjectID `json:"accessModules"`
}

func GetPowerLevel(role string) int {
	switch role {
	case RoleSubAdmin:
		return PowerLevelSubAdmin
	case RoleAdmin:
		return PowerLevelAdmin
	case RoleSuperAdmin:
		return PowerLevelSuperAdmin
	default:
		return 0
	}
}