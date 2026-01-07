package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	
	"io"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ActivityLogMiddleware struct {
	repo *repository.ActivityLogRepository
}

func NewActivityLogMiddleware(repo *repository.ActivityLogRepository) *ActivityLogMiddleware {
	return &ActivityLogMiddleware{repo: repo}
}

func (m *ActivityLogMiddleware) LogActivity() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		adminInterface, exists := c.Get("admin")
		if !exists {
			c.Next()
			return
		}

		admin, ok := adminInterface.(*domain.Admin)
		if !ok || admin == nil {
			c.Next()
			return
		}

		adminID := admin.ID
		adminName := admin.Name
		adminEmail := admin.Email


		path := c.Request.URL.Path
		method := c.Request.Method

		if shouldSkipLogging(path, method) {
			c.Next()
			return
		}

		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			action, entityType, entityID := parseAction(path, method)
			details := extractDetails(bodyBytes, path, method)

			logEntry := domain.ActivityLog{
				AdminID:    adminID,
				AdminName:  adminName,
				AdminEmail: adminEmail,
				Action:     action,
				EntityType: entityType,
				EntityID:   entityID,
				Method:     method,
				Endpoint:   path,
				IPAddress:  c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
				Details:    details,
			}
			

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := m.repo.Create(ctx, &logEntry); err != nil {
				log.Printf("Failed to create activity log: %v", err)
			}
		}
	}
}

func shouldSkipLogging(path, method string) bool {
	skipPaths := []string{
		"/admin/panel/profile",
		"/admin/panel/stats",
		"/admin/activity-logs",
		"/admin/bookings/stats",
		"/admin/complaints/stats",
		"/admin/services/stats",
	}

	for _, skip := range skipPaths {
		if strings.Contains(path, skip) {
			return true
		}
	}

	return false
}

func parseAction(path, method string) (action, entityType, entityID string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case strings.Contains(path, "/users"):
		entityType = "user"
		if strings.Contains(path, "/status") {
			action = "Update User Status"
			for i, part := range parts {
				if part == "users" && i+1 < len(parts) {
					entityID = parts[i+1]
					break
				}
			}
		} else if method == "POST" && strings.Contains(path, "/create") {
			action = "Create User"
		} else if strings.Contains(path, "/notes") {
			action = "Add Booking Note"
		}else if method == "PUT" {
			action = "Update User"
		} else if method == "DELETE" {
			action = "Delete User"
		}

	case strings.Contains(path, "/providers"):
		entityType = "provider"
		if strings.Contains(path, "/status") {
			action = "Update Provider Status"
		} else if strings.Contains(path, "/kyc") {
			action = "Update provider Kyc"
		} else if strings.Contains(path, "/verify-document") {
			action = "Verify Provider Document"
		} else if strings.Contains(path, "/account-action") {
			action = "Update Provider Account Action"
		} else if strings.Contains(path, "/commission") {
			action = "Update Provider Commission"
	    } else if strings.Contains(path, "/notes") {
			action = "Add Booking Note"
		}else if method == "POST" && strings.Contains(path, "/create") {
			action = "Create Provider"
		} else if method == "PUT" && !strings.Contains(path, "/status") && 
			!strings.Contains(path, "/kyc") && !strings.Contains(path, "/verify-document") &&
			!strings.Contains(path, "/account-action") && !strings.Contains(path, "/commission") {
			action = "Update Provider"
		} else if method == "DELETE" {
			action = "Delete Provider"
		}
		for i, part := range parts {
			if (part == "status" || part == "kyc" || part == "verify-document" ||
				part == "account-action" || part == "commission") && i+1 < len(parts) {
				entityID = parts[i+1]
				break
			}
		}

	case strings.Contains(path, "/bookings"):
		entityType = "booking"
		if strings.Contains(path, "/cancel") {
			action = "Cancel Booking"
		} else if strings.Contains(path, "/complete") {
			action = "Complete Booking"
		} else if strings.Contains(path, "/notes") {
			action = "Add Booking Note"
		} else if method == "POST" && strings.Contains(path, "/create") {
			action = "Create Booking"
		} else if method == "PUT" && !strings.Contains(path, "/cancel") && 
			!strings.Contains(path, "/complete") && !strings.Contains(path, "/notes") {
			action = "Update Booking"
		} else if method == "DELETE" {
			action = "Delete Booking"
		}
		for i, part := range parts {
			if part == "bookings" && i+1 < len(parts) && parts[i+1] != "stats" && parts[i+1] != "get-invoice" {
				entityID = parts[i+1]
				break
			}
		}

	case strings.Contains(path, "/complaints"):
		entityType = "complaint"
		if strings.Contains(path, "/assessment") {
			action = "Post Complaint Assessment"
		} else if strings.Contains(path, "/notes") {
			action = "Add Complaint Note"
		} else if strings.Contains(path, "/status") {
			action = "Update  Complaint Status"
		} else if method == "POST" && strings.Contains(path, "/create") {
			action = "Create Complaint"
		} else if method == "PUT" && !strings.Contains(path, "/assessment") && 
			!strings.Contains(path, "/notes") && !strings.Contains(path, "/status") {
			action = "Update Complaint"
		} else if method == "DELETE" {
			action = "Delete Complaint"
		}
		for i, part := range parts {
			if part == "complaints" && i+1 < len(parts) && parts[i+1] != "stats" {
				entityID = parts[i+1]
				break
			}
		}

	case strings.Contains(path, "/provider-payout"):
		entityType = "payout"
		if strings.Contains(path, "/6hour") {
			action = "create_6hour_payout"
		} else if method == "POST" {
			action = "Create Payout"
		} else if method == "PUT" {
			action = "Update Payout"
		} else if method == "DELETE" {
			action = "Delete Payout"
		}

	case strings.Contains(path, "/provider-settlement"):
		entityType = "settlement"
		if strings.Contains(path, "/create") {
			action = "Create Settlement"
		} else if method == "POST" && !strings.Contains(path, "/create") {
			action = "Create Settlement"
		} else if method == "PUT" {
			action = "Update Settlement"
		} else if method == "DELETE" {
			action = "Delete Settlement"
		}

	case strings.Contains(path, "/services"):
		entityType = "service"
		if method == "POST" && strings.Contains(path, "/create") {
			action = "create_service"
		} else if method == "PUT" {
			action = "update_service"
		} else if method == "DELETE" {
			action = "delete_service"
		} else if strings.Contains(path, "/status") {
			action = "update_service_status"
		}
		for i, part := range parts {
			if part == "services" && i+1 < len(parts) && parts[i+1] != "create" && parts[i+1] != "stats" {
				entityID = parts[i+1]
				break
			}
		}

	case strings.Contains(path, "/admin/panel"):
		entityType = "admin"
		if strings.Contains(path, "/login") {
			action = "Admin Login"
		} else if strings.Contains(path, "/logout-all") {
			action = "Admin Logout all"
		} else if strings.Contains(path, "/logout") {
			action = "Admin Logout"
		} else if strings.Contains(path, "/change-password") {
			action = "Change Password"
		} else if strings.Contains(path, "/create") {
			action = "Create Admin"
		} else if strings.Contains(path, "/reset-password") {
			action = "Reset Admin Password"
		} else if strings.Contains(path, "/toggle-status") {
			action = "Toggle Admin Status"
		} else if method == "PUT" && !strings.Contains(path, "/change-password") && 
			!strings.Contains(path, "/reset-password") && !strings.Contains(path, "/toggle-status") {
			action = "Update Admin"
		} else if method == "DELETE" {
			action = "Delete Admin"
		}
		for i, part := range parts {
			if part == "panel" && i+1 < len(parts) &&
				parts[i+1] != "create" && parts[i+1] != "all" &&
				parts[i+1] != "stats" && parts[i+1] != "login" &&
				parts[i+1] != "logout" && parts[i+1] != "logout-all" &&
				parts[i+1] != "profile" && parts[i+1] != "change-password" {
				entityID = parts[i+1]
				break
			}
		}
	}

	if action == "" {
		action = strings.ToLower(method) + "_" + entityType
	}

	if entityID == "" && len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		keywords := []string{"create", "status", "kyc", "verify-document", "account-action", 
			"commission", "cancel", "complete", "notes", "assessment", "6hour"}
		isKeyword := false
		for _, keyword := range keywords {
			if lastPart == keyword {
				isKeyword = true
				break
			}
		}
		if !isKeyword && lastPart != "" {
			entityID = lastPart
		}
	}

	return action, entityType, entityID
}

func extractDetails(bodyBytes []byte, path, method string) string {
	if len(bodyBytes) == 0 || method == "DELETE" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return ""
	}

	sensitiveFields := []string{"password", "token", "secret", "api_key", "otp", "pin"}
	for _, field := range sensitiveFields {
		delete(data, field)
	}

	detailsBytes, _ := json.Marshal(data)
	return string(detailsBytes)
}