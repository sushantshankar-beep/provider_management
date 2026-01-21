package handler

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/service"
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

func (h *PayoutHandler) GetProviderPayouts(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	filters := dto.PayoutFilters{
		Search:     c.Query("search"),
		ProviderID: c.Query("provider_id"),
		Status:     c.Query("status"),
		PeriodFrom: c.Query("period_from"),
		PeriodTo:   c.Query("period_to"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.DefaultQuery("sort_order", "desc"),
	}

	sort := dto.PayoutSort{
		SortBy:    c.Query("sort_by"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	pagination := dto.PaginationParams{
		Page:  page,
		Limit: limit,
	}

	res , err := h.svc.GetProviderPayouts(c.Request.Context(), filters,sort, pagination)

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
		"data":    res.Data,
		"pagination": res.Pagination,
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

