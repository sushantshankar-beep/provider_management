package handler

import (
	"net/http"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	days := c.DefaultQuery("days","7days")

	stats, err := h.svc.GetDashboardStats(c.Request.Context(), period, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}