package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/domain"
)

type ZoneFilterMiddleware struct{}

func NewZoneFilterMiddleware() *ZoneFilterMiddleware {
	return &ZoneFilterMiddleware{}
}

func (m *ZoneFilterMiddleware) ApplyZoneFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := GetAdminFromContext(c)
		if admin == nil {
			c.Next()
			return
		}

		if admin.Role == domain.RoleSuperAdmin {
			c.Set("zoneFilter", bson.M{})
			c.Next()
			return
		}

		if len(admin.ServiceZones) == 0 {
			c.Set("zoneFilter", bson.M{})
			c.Next()
			return
		}

		zoneFilter := bson.M{
			"serviceZone": bson.M{
				"$in": admin.ServiceZones,
			},
		}

		c.Set("zoneFilter", zoneFilter)
		c.Next()
	}
}

func GetZoneFilter(c *gin.Context) bson.M {
	filter, exists := c.Get("zoneFilter")
	if !exists {
		return bson.M{}
	}
	return filter.(bson.M)
}

func GetAdminZones(c *gin.Context) []string {
	admin := GetAdminFromContext(c)
	if admin == nil || admin.Role == domain.RoleSuperAdmin {
		return []string{}
	}
	return admin.ServiceZones
}

func CanAccessAllZones(c *gin.Context) bool {
	admin := GetAdminFromContext(c)
	if admin == nil {
		return false
	}
	return admin.Role == domain.RoleSuperAdmin || len(admin.ServiceZones) == 0
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
	if len(zones) == 0 {
		return ids
	}

	return ids
}