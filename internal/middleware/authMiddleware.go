package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"provider_management/internal/domain"
	"provider_management/internal/repository"

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

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
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
			log.Println("Invalid JWT token:", err)
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
		if err != nil || admin == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Account is deleted or does not exist",
				"code":    "ACCOUNT_DELETED",
			})
			c.Abort()
			return
		}

		iatFloat, ok := claims["iat"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		tokenIssuedAt := time.Unix(int64(iatFloat), 0)

		if admin.PasswordChangedAt != nil &&
			tokenIssuedAt.Before(*admin.PasswordChangedAt) {

			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Password changed. Please login again.",
				"code":    "PASSWORD_CHANGED",
			})
			c.Abort()
			return
		}

		if admin.Status == domain.StatusDeactive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Account is deactivated.",
				"code":    "ACCOUNT_DEACTIVATED",
			})
			c.Abort()
			return
		}

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
