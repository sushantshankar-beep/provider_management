package handler

import (
	"net/http"
	"provider_management/internal/logger"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	svc *service.PermissionService
	log *logger.Logger
}

func NewPermissionHandler(svc *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		svc: svc,
	}
}

func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.svc.ListPermissions(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": permissions})
}

func (h *PermissionHandler) GetPermissionByID(c *gin.Context) {
	id := c.Param("id")
	permission, err := h.svc.GetPermission(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, permission)
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := h.svc.CreatePermission(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusCreated, permission)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := h.svc.UpdatePermission(c, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, permission)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeletePermission(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

