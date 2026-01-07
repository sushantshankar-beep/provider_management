package handler

import (
	"net/http"
	"provider_management/internal/middleware"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
)

type ZoneMapHandler struct {
	svc *service.ZoneMapService
}

func NewZoneMapHandler(svc *service.ZoneMapService) *ZoneMapHandler {
	return &ZoneMapHandler{svc: svc}
}

func (h *ZoneMapHandler) GetZoneStats(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
		return
	}

	stats, err := h.svc.GetZoneStats(c.Request.Context(), admin.ID.Hex())
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Zone statistics fetched successfully",
		"data":    stats,
	})
}

func (h *ZoneMapHandler) GetActivationTeam(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
		return
	}

	zoneName := c.Param("zone")

	team, err := h.svc.GetActivationTeam(c.Request.Context(), admin.ID.Hex(), zoneName)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activation team fetched successfully",
		"data":    team,
	})
}

func (h *ZoneMapHandler) GetProvidersByActivator(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
		return
	}

	zoneName := c.Param("zone")
	activatorName := c.Param("activator")

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "-createdAt")

	providers, err := h.svc.GetProvidersByActivator(
		c.Request.Context(),
		admin.ID.Hex(),
		zoneName,
		activatorName,
		page,
		limit,
		sort,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Providers fetched successfully",
		"data":    providers,
	})
}

func (h *ZoneMapHandler) GetMyProviders(c *gin.Context) {
	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
		return
	}

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "-createdAt")

	providers, err := h.svc.GetMyProviders(
		c.Request.Context(),
		admin.ID.Hex(),
		page,
		limit,
		sort,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "My providers fetched successfully",
		"data":    providers,
	})
}