package handler

import (
	"log"
	"net/http"
	"provider_management/internal/domain"
	"provider_management/internal/middleware"
	"provider_management/internal/service"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(service *service.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type CreateAdminRequestBody struct {
    Name          string   `form:"name" binding:"required"`
    Email         string   `form:"email" binding:"required,email"`
    Phone         string   `form:"phone" binding:"required"`
    Password      string   `form:"password" binding:"required,min=8"`
    Role          string   `form:"role" binding:"required"`
    RoleName      string   `form:"roleName" binding:"required"`
    ServiceZones  []string `form:"serviceZones"`
    AccessModules []string `form:"accessModules"`
}
type UpdateAdminRequestBody struct {
    Name          string   `form:"name"`
    Email         string   `form:"email"`
    Phone         string   `form:"phone"`
    Password      string   `form:"password"`
    Role          string   `form:"role"`
    ServiceZones  []string `form:"serviceZones"`
    AccessModules []string `form:"accessModules"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, admin, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "Account is deactivated" {
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"admin": gin.H{
			"_id":           admin.ID,
			"name":          admin.Name,
			"email":         admin.Email,
			"phone":         admin.Phone,
			"role":          admin.Role,
			"powerLevel":    admin.PowerLevel,
			"profile":       admin.ProfileURL,
			"serviceZones":  admin.ServiceZones,
			"accessModules": admin.AccessModules,
		},
	})
}

func (h *AdminHandler) Logout(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	token := middleware.GetTokenFromContext(c)

	if err := h.service.Logout(c.Request.Context(), admin, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *AdminHandler) LogoutAll(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)

	if err := h.service.LogoutAll(c.Request.Context(), admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices"})
}

func isValidPhoneNumber(phone string) bool {
	re := regexp.MustCompile(`^[6-9]\d{9}$`)
	return re.MatchString(phone)
}


func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequestBody
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if req.Role != "" && !domain.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid role. Allowed roles: subAdmin, admin, superAdmin",
		})
		return
	}

	if !isValidPhoneNumber(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid phone number. Use format (e.g. +919876543210)",
		})
		return
	}

	var profileURL string
	if url, exists := middleware.GetUploadedURL(c, "profileUrl"); exists {
		profileURL = url
	}

	var accessModules []primitive.ObjectID
	for _, id := range req.AccessModules {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			accessModules = append(accessModules, objID)
		}
	}

	serviceReq := service.CreateAdminRequest{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Password:      req.Password,
		Role:          req.Role,
		RoleName:      req.RoleName,
		ServiceZones:  req.ServiceZones,
		AccessModules: accessModules,
		ProfileURL:    profileURL,
	}

	admin := middleware.GetAdminFromContext(c)
	newAdmin, err := h.service.CreateAdmin(c.Request.Context(), serviceReq, admin.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin created successfully",
		"admin":   newAdmin,
	})
}



func (h *AdminHandler) GetAllAdmins(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")

	admins, total, err := h.service.GetAllAdmins(c.Request.Context(), limit, offset, search, status, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"admins": admins,
	})
}

func (h *AdminHandler) GetAdminByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid admin ID"})
		return
	}

	admin, err := h.service.GetAdminByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": admin})
}

func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	log.Println("UpdateAdmin handler called")
	
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid admin ID"})
		return
	}

	var req UpdateAdminRequestBody
	if err := c.ShouldBind(&req); err != nil {
		log.Println("Binding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Printf("Request data: %+v\n", req)

	if req.Role != "" && !domain.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid role. Allowed roles: subAdmin, admin, superAdmin",
		})
		return
	}

	if req.Phone != "" && !isValidPhoneNumber(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid phone number. Use format (e.g. +919876543210)",
		})
		return
	}

	var profileURL string
	if url, exists := middleware.GetUploadedURL(c, "profileUrl"); exists {
		profileURL = url
		log.Println("Profile URL from S3:", profileURL)
	} else {
		log.Println("No profile image uploaded")
	}

	var accessModules []primitive.ObjectID
	for _, idStr := range req.AccessModules {
		if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
			accessModules = append(accessModules, objID)
		} else {
			log.Printf("Invalid access module ID: %s, error: %v\n", idStr, err)
		}
	}

	serviceReq := service.UpdateAdminRequest{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Role:          req.Role,
		ServiceZones:  req.ServiceZones,
		AccessModules: accessModules,
		ProfileURL:    profileURL,
	}

	log.Printf("Service request: %+v\n", serviceReq)

	admin, err := h.service.UpdateAdmin(c.Request.Context(), id, serviceReq)
	if err != nil {
		log.Println("Service error:", err)
		if err.Error() == "Admin not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Println("Admin updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin updated successfully",
		"admin":   admin,
	})
}

func (h *AdminHandler) ToggleAdminStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid admin ID"})
		return
	}

	admin, err := h.service.ToggleAdminStatus(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "Admin not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin status updated successfully", "admin": admin})
}

func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid admin ID"})
		return
	}

	if err := h.service.DeleteAdmin(c.Request.Context(), id); err != nil {
		if err.Error() == "Admin not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin deleted successfully"})
}

func (h *AdminHandler) ResetPasswordBySuperAdmin(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	if admin.Role != domain.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "Only Super Admin can reset passwords"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid admin ID"})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.service.ResetPasswordBySuperAdmin(c.Request.Context(), id, req.NewPassword); err != nil {
		if err.Error() == "Admin not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AdminHandler) ChangeOwnPassword(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	token := middleware.GetTokenFromContext(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.service.ChangeOwnPassword(c.Request.Context(), admin, req.CurrentPassword, req.NewPassword, token); err != nil {
		if err.Error() == "Current password is incorrect" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)

	profile, err := h.service.GetProfile(c.Request.Context(), admin.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": profile})
}