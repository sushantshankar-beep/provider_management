package handler

import (
	"net/http"
	"provider_management/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{
		svc: svc,
	}
}

func (h *OrderHandler) GetAll(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	planStatus := c.Query("planStatus")
	paymentStatus := c.Query("paymentStatus")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	sortBy := c.DefaultQuery("sortBy", "updatedAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	res, total, err := h.svc.ListOrders(c, page, limit, search, planStatus, paymentStatus, startDate, endDate, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
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
			"currentPage": pageInt,
			"totalPages":  totalPages,
			"totalOrders": total,
			"limit":       limitInt,
		},
	})
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetOrder(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Order not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

func (h *OrderHandler) ExportToCSV(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	paymentStatus := c.Query("paymentStatus")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	csv, err := h.svc.ExportOrdersToCSV(c, search, status, paymentStatus, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to export orders",
		})
		return
	}

	filename := "orders-" + strconv.FormatInt(c.Request.Context().Value("timestamp").(int64), 10) + ".csv"
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(http.StatusOK, csv)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		PlanStatus    string `json:"planStatus"`
		PaymentStatus string `json:"paymentStatus"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	order, err := h.svc.UpdateOrderStatus(c, id, req.PlanStatus, req.PaymentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}
