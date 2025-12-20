package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminID     primitive.ObjectID `bson:"admin_id" json:"admin_id"`
	AdminName   string             `bson:"admin_name" json:"admin_name"`
	AdminEmail  string             `bson:"admin_email" json:"admin_email"`
	Action      string             `bson:"action" json:"action"`
	EntityType  string             `bson:"entity_type" json:"entity_type"`
	EntityID    string             `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	Method      string             `bson:"method" json:"method"`
	Endpoint    string             `bson:"endpoint" json:"endpoint"`
	IPAddress   string             `bson:"ip_address" json:"ip_address"`
	UserAgent   string             `bson:"user_agent" json:"user_agent"`
	Details     string             `bson:"details,omitempty" json:"details,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}


type ActivityLogResponse struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	AdminName   string            `bson:"admin_name" json:"admin_name"`
	EntityType string             `bson:"entity_type" json:"entity_type"`
	EntityID   string             `bson:"entity_id" json:"entity_id"`
	Action     string             `bson:"action" json:"action"`
	AdminID    primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}