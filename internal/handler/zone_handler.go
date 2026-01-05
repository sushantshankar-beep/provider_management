package handler

import (
	"net/http"
	"provider_management/internal/logger"
	"provider_management/internal/service"
	"strconv"
	"github.com/gin-gonic/gin"
)

type ZoneHandler struct {
	svc *service.ZoneService
	log *logger.Logger
}

func NewZoneHandler(svc *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{
		svc: svc,
	}
}

func (h *ZoneHandler) Create(c *gin.Context) {
	var req struct {
		ZoneName  string `json:"zoneName" binding:"required"`
		StateName string `json:"stateName" binding:"required"`
		IsActive  *bool  `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	zone, err := h.svc.CreateZone(c, req.ZoneName, req.StateName, req.IsActive)
	if err != nil {
		if err.Error() == "zone already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Zone with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Zone created successfully", "data": zone})
}

func (h *ZoneHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	isActive := c.Query("isActive")
	state := c.Query("state")          
	createdAt := c.Query("createdAt")
	updatedAt := c.Query("updatedAt") 

	zones, stats, total, err := h.svc.ListZones(c, page, limit, search, isActive, state, createdAt, updatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	pageInt, _ := strconv.ParseInt(page, 10, 64)
	limitInt, _ := strconv.ParseInt(limit, 10, 64)

	if pageInt < 1 {
		pageInt = 1
	}
	if limitInt < 1 {
		limitInt = 10
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + limitInt - 1) / limitInt
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
		"data":    zones,
		"pagination": gin.H{
			"total": total,
			"page":  pageInt,
			"limit": limitInt,
			"pages": totalPages,
		},
	})
}

func (h *ZoneHandler) GetActive(c *gin.Context) {
	zones, err := h.svc.GetActiveZones(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": zones})
}

func (h *ZoneHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		ZoneName  *string `json:"zoneName"`
		StateName *string `json:"stateName"`
		IsActive  *bool   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	zone, err := h.svc.UpdateZone(c, id, req.ZoneName, req.StateName, req.IsActive)
	if err != nil {
		if err.Error() == "zone not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Zone updated successfully", "data": zone})
}

func (h *ZoneHandler) ToggleStatus(c *gin.Context) {
	id := c.Param("id")

	zone, err := h.svc.ToggleZoneStatus(c, id)
	if err != nil {
		if err.Error() == "zone not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	status := "deactivated"
	if zone.IsActive {
		status = "activated"
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Zone " + status + " successfully", "data": zone})
}

func (h *ZoneHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.svc.DeleteZone(c, id)
	if err != nil {
		if err.Error() == "zone not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Zone deleted successfully"})
}

func (h *ZoneHandler) GetActiveStates(c *gin.Context) {
	zones, err := h.svc.GetActiveStates(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": zones})
}
