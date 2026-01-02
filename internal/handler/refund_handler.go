package handler

import (
	"log"
	"net/http"
	"provider_management/internal/domain"
	"provider_management/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RefundHandler struct {
	refundService *service.RefundService
}

func NewRefundHandler(refundService *service.RefundService) *RefundHandler {
	return &RefundHandler{
		refundService: refundService,
	}
}

func (h *RefundHandler) GetAllRefunds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	userID := c.Query("user_id")
	mode :=   c.Query("mode")
	reason := c.Query("reason")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.refundService.GetAllRefunds(c.Request.Context(), domain.RefundFilter{
		Page:   page,
		Limit:  limit,
		Status: status,
		UserID: userID,
		Mode: mode,
		Reason: reason,
		Search: search,
	})

	if err != nil {
		log.Printf("GetAllRefunds error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch refunds: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Refunds fetched successfully",
		"data":    result,
	})
}
func (h *RefundHandler) GetRefundByID(c *gin.Context) {
	id := c.Param("id")
	id = strings.TrimPrefix(id, "REF")

	log.Printf("Fetching refund with ID: %s", id)

	refund, err := h.refundService.GetRefundByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("GetRefundByID error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Refund not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Refund fetched successfully",
		"data":    refund,
	})
}