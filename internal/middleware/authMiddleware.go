package middleware

import (
	"context"
	
	"net/http"
	"os"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
type AuthMiddleware struct {
    adminRepo *repository.AdminRepository
}

func NewAuthMiddleware(adminRepo *repository.AdminRepository) *AuthMiddleware {
    return &AuthMiddleware{
        adminRepo: adminRepo,
    }
}

func (m *AuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		// ✅ Allow CORS preflight
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(
			strings.TrimPrefix(authHeader, "Bearer "),
		)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		idHex, ok := claims["_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		adminID, err := primitive.ObjectIDFromHex(idHex)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		admin, err := m.adminRepo.FindByID(context.Background(), adminID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		if admin.Status == domain.StatusDeactive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Account is deactivated.",
				"code": "ACCOUNT_DEACTIVATED",
			})
			c.Abort()
			return
		}
		

		// allowedForAllAdmins := []string{
		// 	"/admin/panel/logout",
		// 	"/admin/panel/logout-all",
		// 	"/admin/panel/profile",
		// 	"/admin/panel/change-password",
		// }

		// path := c.FullPath()
		// isAllowedForAll := false
		// for _, allowedPath := range allowedForAllAdmins {
		// 	if path == allowedPath {
		// 		isAllowedForAll = true
		// 		break
		// 	}
		// }

	
		// if !isAllowedForAll && c.Request.Method != http.MethodGet && admin.Role != domain.RoleSuperAdmin {
		// 	c.JSON(http.StatusForbidden, gin.H{"message": "Only superAdmin can perform this action"})
		// 	c.Abort()
		// 	return
		// }

		c.Set("admin", admin)
		c.Set("token", tokenString)
		c.Next()
	}
}



func GetAdminFromContext(c *gin.Context) *domain.Admin {
	admin, exists := c.Get("admin")
	if !exists {
		return nil
	}
	return admin.(*domain.Admin)
}

func GetTokenFromContext(c *gin.Context) string {
	token, exists := c.Get("token")
	if !exists {
		return ""
	}
	return token.(string)
}