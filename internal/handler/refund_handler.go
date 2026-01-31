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

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filter := domain.RefundFilter{
		Page:   page,
		Limit:  limit,
		Status: c.Query("status"),
		UserID: c.Query("user_id"),
		Mode:   c.Query("mode"),
		Reason: c.Query("reason"),
		Search: c.Query("search"),
	}

	result, err := h.refundService.GetAllRefunds(c.Request.Context(), filter)
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
	id := strings.TrimPrefix(c.Param("id"), "REF")

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

func (h *RefundHandler) InitiateRefund(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "REF")

	log.Printf("Initiating refund with ID: %s", id)

	if err := h.refundService.InitiateRefund(c.Request.Context(), id); err != nil {
		log.Printf("InitiateRefund error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Refund initiated successfully",
	})
}


func (h *RefundHandler) CheckRefundStatus(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "REF")

	log.Printf("Checking refund status for ID: %s", id)

	if err := h.refundService.CheckRefundStatus(c.Request.Context(), id); err != nil {
		log.Printf("CheckRefundStatus error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to check refund status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Refund status checked successfully",
	})
}