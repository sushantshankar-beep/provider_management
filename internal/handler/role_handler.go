package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"provider_management/internal/domain"
	"provider_management/internal/middleware"
	"provider_management/internal/repository"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleHandler struct {
	svc  *service.RoleService
	repo *repository.RoleRepository
}

func NewRoleHandler(repo *repository.RoleRepository,svc *service.RoleService) *RoleHandler {
	return &RoleHandler{repo: repo, svc: svc}
}

// --------------------
// CREATE ROLE
// --------------------
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req domain.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload"})
		return
	}

	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Role name required"})
		return
	}

	// 🔒 Prevent duplicate role
	if _, err := h.repo.FindByName(c, req.Name); err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "Role already exists"})
		return
	}

	// Defaults & metadata
	req.CreatedBy = admin.ID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if req.Status == "" {
		req.Status = domain.RoleInactive
	}
	if req.ZoneScope == "" {
		req.ZoneScope = domain.ZoneScopeAll
	}

	if err := h.repo.Create(c, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"data": gin.H{
			"id":     req.ID,
			"name":   req.Name,
			"status": req.Status,
		},
	})
}

// --------------------
// LIST ROLES
// --------------------
func (h *RoleHandler) ListRoles(c *gin.Context) {
	status := c.Query("status")

	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	roles, err := h.repo.ListWithCreator(c, filter)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to fetch roles"})
		return
	}

	c.JSON(200, roles)
}


// --------------------
// ACTIVATE / DEACTIVATE ROLE
// --------------------
func (h *RoleHandler) ToggleRoleStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}

	var body struct {
		Status domain.RoleStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload"})
		return
	}

	if err := h.repo.UpdateStatus(c, id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update role status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role status updated"})
}
//
// CREATE ROLE (already working)
//

// --------------------
// CLONE ROLE
// --------------------
func (h *RoleHandler) CloneRole(c *gin.Context) {
	sourceID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Role name required"})
		return
	}

	admin := middleware.GetAdminFromContext(c)

	sourceRole, err := h.repo.FindByID(c, sourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Source role not found"})
		return
	}

	newName := strings.TrimSpace(body.Name)
	if newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Role name cannot be empty"})
		return
	}
	if _, err := h.repo.FindByName(c, newName); err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "Role name already exists"})
		return
	}
	newRole := domain.Role{
		Name:        newName,
		RoleType:    sourceRole.RoleType,
		ZoneScope:   sourceRole.ZoneScope,
		Description: sourceRole.Description,
		Permissions: clonePermissions(sourceRole.Permissions),
		Status:    domain.RoleInactive,
		CreatedBy: admin.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.repo.Create(c, &newRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to clone role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role cloned successfully",
		"data": gin.H{
			"id":     newRole.ID,
			"name":   newRole.Name,
			"status": newRole.Status,
		},
	})
}


// --------------------
// DELETE ROLE
// --------------------
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}

	role, err := h.repo.FindByID(c, roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Role not found"})
		return
	}
	if role.Name == domain.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "SuperAdmin role cannot be deleted"})
		return
	}
	if role.Status == domain.RoleActive {
		c.JSON(http.StatusForbidden, gin.H{"message": "Deactivate role before deleting"})
		return
	}
	if err := h.repo.DeleteByID(c, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
func clonePermissions(src map[string][]string) map[string][]string {
	if src == nil {
		return nil
	}

	dst := make(map[string][]string, len(src))
	for module, actions := range src {
		actionsCopy := make([]string, len(actions))
		copy(actionsCopy, actions)
		dst[module] = actionsCopy
	}
	return dst
}
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	// --------------------
	// Parse Role ID
	// --------------------
	roleID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}

	// --------------------
	// Bind Request Body
	// --------------------
	var body struct {
		Name        string              `json:"name"`
		Description string              `json:"description"`
		RoleType    string              `json:"roleType"`
		ZoneScope   string              `json:"zoneScope"`
		Permissions map[string][]string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload"})
		return
	}

	// --------------------
	// Fetch Existing Role
	// --------------------
	role, err := h.repo.FindByID(c, roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Role not found"})
		return
	}

	// --------------------
	// Protect System Role
	// --------------------
	if role.Name == domain.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "SuperAdmin role cannot be edited",
		})
		return
	}

	// --------------------
	// Prepare Update Payload
	// --------------------
	update := bson.M{}

	// Name (with duplicate check)
	if body.Name != "" {
		name := strings.TrimSpace(body.Name)

		existing, err := h.repo.FindByName(c, name)
		if err == nil && existing.ID != roleID {
			c.JSON(http.StatusConflict, gin.H{
				"message": "Role name already exists",
			})
			return
		}

		update["name"] = name
	}

	// Description
	if body.Description != "" {
		update["description"] = body.Description
	}

	// Role Type
	if body.RoleType != "" {
		update["roleType"] = body.RoleType
	}

	// Zone Scope
	if body.ZoneScope != "" {
		update["zoneScope"] = body.ZoneScope
	}

	// Permissions
	if body.Permissions != nil {
		update["permissions"] = body.Permissions
	}

	// --------------------
	// Nothing to Update Check
	// --------------------
	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Nothing to update",
		})
		return
	}

	update["updatedAt"] = time.Now()

	// --------------------
	// Persist Update
	// --------------------
	if err := h.repo.UpdateByID(c, roleID, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update role",
		})
		return
	}

	// --------------------
	// Success Response
	// --------------------
	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
	})
}
func (h *RoleHandler) GetRoleByID(c *gin.Context) {
	roleID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role ID"})
		return
	}

	role, err := h.repo.FindByIDWithCreator(c, roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}


func (h *RoleHandler) GetRoleTypes(c *gin.Context) {
	roleTypes, err := h.svc.GetRoleTypes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roleTypes})
}

func (h *RoleHandler) GetRoleNamesByType(c *gin.Context) {
	roleType := c.Query("roleType")
	log.Println("roleType",roleType)
	if roleType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roleType required"})
		return
	}

	roleNames, err := h.svc.GetRoleNamesByType(c, roleType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roleNames})
}