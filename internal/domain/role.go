// domain/role.go
type Role struct {
	ID          primitive.ObjectID            `bson:"_id,omitempty" json:"_id"`
	Name        string                        `bson:"name" json:"name"`
	RoleType    string                        `bson:"roleType" json:"roleType"` // admin | subAdmin
	Status      RoleStatus                    `bson:"status" json:"status"`
	ZoneScope   string                        `bson:"zoneScope" json:"zoneScope"` // all | assigned
	Description string                        `bson:"description,omitempty" json:"description,omitempty"`
	Permissions map[string][]string           `bson:"permissions" json:"permissions"`
	CreatedBy   primitive.ObjectID            `bson:"createdBy"`
	CreatedAt   time.Time                     `bson:"createdAt"`
	UpdatedAt   time.Time                     `bson:"updatedAt"`
}
