package handler

import (
	"context"
	// "net/http"

	"provider_management/internal/domain"
	"provider_management/internal/middleware"
	"provider_management/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleHandler struct {
	repo *repository.RoleRepository
}

func NewRoleHandler(repo *repository.RoleRepository) *RoleHandler {
	return &RoleHandler{repo: repo}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req domain.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid payload"})
		return
	}
	admin := middleware.GetAdminFromContext(c)
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(400, gin.H{"message": "Role name required"})
		return
	}

	// 🔒 Prevent duplicate roles
	if _, err := h.roleRepo.FindByName(c, req.Name); err == nil {
		c.JSON(409, gin.H{"message": "Role already exists"})
		return
	}
	req.CreatedBy = admin.ID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if req.Status == "" {
		req.Status = domain.RoleInactive
	}
	if req.ZoneScope == "" {
		req.ZoneScope = "all"
	}

	if err := h.roleRepo.Create(c, &req); err != nil {
		c.JSON(500, gin.H{"message": "Failed to create role"})
		return
	}

	c.JSON(201, gin.H{
		"message": "Role created successfully",
		"data": gin.H{
			"id": req.ID,
			"name": req.Name,
			"status": req.Status,
		},
	})
}


func (h *RoleHandler) ListRoles(c *gin.Context) {
	status := c.Query("status")

	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	roles, _ := h.repo.List(c, filter)
	c.JSON(200, roles)
}

func (h *RoleHandler) ToggleRoleStatus(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var body struct {
		Status domain.RoleStatus `json:"status"`
	}
	c.BindJSON(&body)

	h.repo.UpdateStatus(c, id, body.Status)
	c.JSON(200, gin.H{"message": "Role status updated"})
}
