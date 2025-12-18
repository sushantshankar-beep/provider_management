package handler

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"provider_management/internal/service"
	"strconv"
	"strings"
)

type PayoutHandler struct {
	svc *service.PayoutService
}

func NewPayoutHandler(svc *service.PayoutService) *PayoutHandler {
	return &PayoutHandler{svc: svc}
}

func (h *PayoutHandler) Create6HourPayout(c *gin.Context) {
	if err := h.svc.CreatePayoutLast6Hours(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payouts created successfully",
	})
}

func (h *PayoutHandler) GetPayouts(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	search := c.Query("search")
	providerID := c.Query("provider_id")
	status := c.Query("status")
	periodFrom := c.Query("period_from")
	periodTo := c.Query("period_to")
	sortBy := c.Query("sort_by")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	payouts, total, totalPages, err := h.svc.GetPayouts(
		c.Request.Context(),
		page,
		limit,
		search,
		providerID,
		status,
		periodFrom,
		periodTo,
		sortBy,
		sortOrder,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch payouts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payouts fetched successfully",
		"data":    payouts,
		"pagination": gin.H{
			"page":         page,
			"limit":        limit,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     page < totalPages,
			"has_previous": page > 1,
		},
	})
}

func (h *PayoutHandler) GetPayoutServices(c *gin.Context) {
	payoutID := c.Param("id")
	if payoutID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Payout ID is required",
		})
		return
	}

	services, err := h.svc.GetPayoutServices(c.Request.Context(), payoutID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Payout not found",
			})
			return
		}

		if strings.Contains(err.Error(), "invalid payout ID") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Invalid payout ID format. Expected format: SET1765957752215",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch payout services: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payout services fetched successfully",
		"data":    services,
	})
}

func (h *PayoutHandler) GetProviderPayoutDetails(c *gin.Context) {
	payoutID := c.Param("id")
	if payoutID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Payout ID is required",
		})
		return
	}

	details, err := h.svc.GetProviderPayoutDetails(c.Request.Context(), payoutID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Payout not found",
			})
			return
		}

		if strings.Contains(err.Error(), "invalid payout ID") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Invalid payout ID format",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch payout details: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payout details fetched successfully",
		"data":    details,
	})
}
