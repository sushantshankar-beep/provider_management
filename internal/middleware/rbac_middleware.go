package middleware

import (
	"net/http"

	"provider_management/internal/domain"
	"provider_management/internal/repository"

	"github.com/gin-gonic/gin"
)

type RBACMiddleware struct {
	roleRepo *repository.RoleRepository
}

func NewRBACMiddleware(roleRepo *repository.RoleRepository) *RBACMiddleware {
	return &RBACMiddleware{roleRepo: roleRepo}
}

func (r *RBACMiddleware) Check(module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {

		admin := GetAdminFromContext(c)
		if admin == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}
   
		if admin.Role == domain.RoleSuperAdmin {
			c.Next()
			return
		}

		role, err := r.roleRepo.FindByName(c, admin.RoleName)

		if err != nil || role.Status != domain.RoleActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Role inactive or not found",
			})
			return
		}
        
		actions := role.Permissions[module]
		for _, a := range actions {
			if a == action {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Permission denied",
		})
	}
}
