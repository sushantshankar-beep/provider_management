package middleware

import (
	"context"
	"log"
	"provider_management/internal/domain"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ZoneFilterMiddleware struct {
	db *mongo.Database
}

func NewZoneFilterMiddleware(db *mongo.Database) *ZoneFilterMiddleware {
	return &ZoneFilterMiddleware{db: db}
}

func (m *ZoneFilterMiddleware) ApplyZoneFilter(filterType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Zone Filter Middleware Started - FilterType: %s", filterType)
		
		admin := GetAdminFromContext(c)
		if admin == nil {
			log.Println("No admin found in context")
			c.Set("zoneFilter", bson.M{})
			c.Next()
			return
		}

		log.Printf("Admin Role: %s, RoleID: %v", admin.Role, admin.RoleID)

		if admin.Role == domain.RoleSuperAdmin {
			log.Println("SuperAdmin - No zone filter applied")
			c.Set("zoneFilter", bson.M{})
			c.Set("adminZones", map[string][]string{
				"zoneName":  []string{},
				"zoneScope": []string{},
			})
			c.Next()
			return
		}

		roleCollection := m.db.Collection("roles")
		var role struct {
			ZoneName  []string `bson:"zoneName"`
			ZoneScope []string `bson:"zoneScope"`
		}
		
		err := roleCollection.FindOne(c.Request.Context(), bson.M{"_id": admin.RoleID}).Decode(&role)
		if err != nil {
			log.Printf("Error fetching role: %v", err)
			c.Set("zoneFilter", bson.M{})
			c.Set("adminZones", map[string][]string{
				"zoneName":  []string{},
				"zoneScope": []string{},
			})
			c.Next()
			return
		}

		log.Printf("Role ZoneName: %v, ZoneScope: %v", role.ZoneName, role.ZoneScope)

		if len(role.ZoneName) == 0 && len(role.ZoneScope) == 0 {
			log.Println("No zones defined for role - No filter applied")
			c.Set("zoneFilter", bson.M{})
			c.Set("adminZones", map[string][]string{
				"zoneName":  []string{},
				"zoneScope": []string{},
			})
			c.Next()
			return
		}

		zoneFilter := bson.M{}

		switch filterType {
		case "zoneName":
			if len(role.ZoneName) > 0 {
				zoneFilter = bson.M{
					"state": bson.M{"$in": role.ZoneName},
				}
				log.Printf("Applied ZoneName Filter: %+v", zoneFilter)
			}
		case "zoneScope":
			if len(role.ZoneScope) > 0 {
				zoneFilter = bson.M{
					"city": bson.M{"$in": role.ZoneScope},
				}
				log.Printf("Applied ZoneScope Filter: %+v", zoneFilter)
			}
		case "both":
			orConditions := []bson.M{}
			log.Println(role.ZoneName)
			if len(role.ZoneName) > 0 {
				orConditions = append(orConditions, bson.M{
					"state": bson.M{"$in": role.ZoneName},
				})
			}
			
			
			if len(role.ZoneScope) > 0 {
				orConditions = append(orConditions, bson.M{
					"city": bson.M{"$in": role.ZoneScope},
				})
			}

			if len(orConditions) > 0 {
				if len(orConditions) == 1 {
					zoneFilter = orConditions[0]
				} else {
					zoneFilter = bson.M{"$or": orConditions}
				}
			}
			log.Printf("Applied Both Zones Filter: %+v", zoneFilter)
		default:
			log.Printf("Invalid filterType: %s, applying both by default", filterType)
			orConditions := []bson.M{}
			
			if len(role.ZoneName) > 0 {
				orConditions = append(orConditions, bson.M{
					"state": bson.M{"$in": role.ZoneName},
				})
			}
			
			if len(role.ZoneScope) > 0 {
				orConditions = append(orConditions, bson.M{
					"city": bson.M{"$in": role.ZoneScope},
				})
			}

			if len(orConditions) > 0 {
				if len(orConditions) == 1 {
					zoneFilter = orConditions[0]
				} else {
					zoneFilter = bson.M{"$or": orConditions}
				}
			}
		}
		
		c.Set("zoneFilter", zoneFilter)
		c.Set("zoneFilterType", filterType)
		c.Set("adminZones", map[string][]string{
			"zoneName":  role.ZoneName,
			"zoneScope": role.ZoneScope,
		})
		
		c.Next()
	}
}

func GetZoneFilterType(c *gin.Context) string {
	filterType, exists := c.Get("zoneFilterType")
	if !exists {
		return "both"
	}
	return filterType.(string)
}

func GetZoneFilter(c *gin.Context) bson.M {
	filter, exists := c.Get("zoneFilter")
	if !exists {
		return bson.M{}
	}
	return filter.(bson.M)
}

func GetAdminZones(c *gin.Context) map[string][]string {
	zones, exists := c.Get("adminZones")
	if !exists {
		return map[string][]string{
			"zoneName":  []string{},
			"zoneScope": []string{},
		}
	}
	return zones.(map[string][]string)
}

func CanAccessAllZones(c *gin.Context) bool {
	admin := GetAdminFromContext(c)
	if admin == nil {
		return false
	}
	return admin.Role == domain.RoleSuperAdmin
}

func BuildZoneQuery(c *gin.Context, baseQuery bson.M) bson.M {
	zoneFilter := GetZoneFilter(c)
	
	if len(zoneFilter) == 0 {
		return baseQuery
	}

	if len(baseQuery) == 0 {
		return zoneFilter
	}

	return bson.M{
		"$and": []bson.M{
			baseQuery,
			zoneFilter,
		},
	}
}

func FilterByZones(ctx context.Context, c *gin.Context, ids []primitive.ObjectID) []primitive.ObjectID {
	if CanAccessAllZones(c) {
		return ids
	}

	zones := GetAdminZones(c)
	if len(zones["zoneName"]) == 0 && len(zones["zoneScope"]) == 0 {
		return ids
	}

	return ids
}