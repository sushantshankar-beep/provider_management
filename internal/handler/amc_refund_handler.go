package handler

import (
	"net/http"
	"provider_management/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AMCRefundHandler struct {
	svc *service.AMCRefundService
}

func NewAMCRefundHandler(svc *service.AMCRefundService) *AMCRefundHandler {
	return &AMCRefundHandler{
		svc: svc,
	}
}

func (h *AMCRefundHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.DefaultQuery("search", "")
	status := c.DefaultQuery("status", "")
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	res, total, err := h.svc.ListRefundRequests(c, page, limit, search, status, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
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
		"data":    res,
		"pagination": gin.H{
			"currentPage":  pageInt,
			"totalPages":   totalPages,
			"totalItems":   total,
			"itemsPerPage": limitInt,
		},
	})
}

func (h *AMCRefundHandler) GetDetails(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetRefundDetails(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Refund request not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func (h *AMCRefundHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")

	var req struct {
		AdminNote string `json:"adminNote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	res, err := h.svc.ApproveRefund(c, id, req.AdminNote, adminID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "Refund request not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Refund approved and initiated successfully",
		"data":    res,
	})
}

func (h *AMCRefundHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")

	var req struct {
		AdminNote string `json:"adminNote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	if req.AdminNote == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Rejection reason is required"})
		return
	}

	res, err := h.svc.RejectRefund(c, id, req.AdminNote, adminID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "Refund request not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Refund request rejected successfully",
		"data":    res,
	})
}

func (h *AMCRefundHandler) CheckStatus(c *gin.Context) {
	id := c.Param("id")

	res, err := h.svc.CheckRefundStatus(c, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "Refund request not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Refund status checked",
		"data":    res,
	})
}

func (h *AMCRefundHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetRefundStats(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch refund stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
