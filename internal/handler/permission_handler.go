package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"provider_management/internal/service"
)

type PermissionHandler struct {
	svc *service.PermissionService
}

func NewPermissionHandler(svc *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
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
	var req dto.CreatePermissionRequest
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
	
	var req dto.UpdatePermissionRequest
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