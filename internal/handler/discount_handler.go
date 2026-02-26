package handler

import (
	"net/http"
	"provider_management/internal/dto"
	"provider_management/internal/service"
	"strconv"
	"strings"
    "math"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type DiscountHandler struct {
	svc *service.DiscountService
}

func NewDiscountHandler(svc *service.DiscountService) *DiscountHandler {
	return &DiscountHandler{svc: svc}
}

func (h *DiscountHandler) CreateDiscount(c *gin.Context) {
	var req dto.CreateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	discount, err := h.svc.CreateDiscount(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Discount created successfully",
		"data":    discount,
	})
}

func (h *DiscountHandler) GetDiscounts(c *gin.Context) {
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

	discounts, total, totalPages, err := h.svc.GetDiscounts(
		c.Request.Context(),
		page,
		limit,
		c.Query("search"),
		c.Query("status"),
		c.Query("scope"),
		c.Query("sort_by"),
		c.DefaultQuery("sort_order", "desc"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch discounts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discounts fetched successfully",
		"data":    discounts,
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

func (h *DiscountHandler) GetDiscountByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Discount ID is required",
		})
		return
	}

	discount, err := h.svc.GetDiscountByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Discount not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch discount: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discount fetched successfully",
		"data":    discount,
	})
}

func (h *DiscountHandler) UpdateDiscount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Discount ID is required",
		})
		return
	}

	var req dto.UpdateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	updatedBy, _ := c.Get("admin_id")
	updatedByStr, _ := updatedBy.(string)

	discount, err := h.svc.UpdateDiscount(c.Request.Context(), id, req, updatedByStr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Discount not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update discount: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discount updated successfully",
		"data":    discount,
	})
}

func (h *DiscountHandler) UpdateDiscountStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Discount ID is required",
		})
		return
	}

	var req dto.UpdateDiscountStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	updatedBy, _ := c.Get("admin_id")
	updatedByStr, _ := updatedBy.(string)

	if err := h.svc.UpdateDiscountStatus(c.Request.Context(), id, req, updatedByStr); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Discount not found",
			})
			return
		}
		if strings.Contains(err.Error(), "invalid status") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update discount status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discount status updated successfully",
	})
}

func (h *DiscountHandler) DeleteDiscount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Discount ID is required",
		})
		return
	}

	if err := h.svc.DeleteDiscount(c.Request.Context(), id); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Discount not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete discount: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discount deleted successfully",
	})
}

func (h *DiscountHandler) GetDiscountStats(c *gin.Context) {
	stats, err := h.svc.GetDiscountStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch discount stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Discount stats fetched successfully",
		"data":    stats,
	})
}

func (h *DiscountHandler) ListDiscountUsage(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	search := c.Query("search")
	status := c.Query("status")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	data, total, totalPages, err := h.svc.ListDiscountsWithUsage(
		c.Request.Context(),
		page, limit,
		search, status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

func (h *DiscountHandler) GetDiscountUsageByDiscountID(c *gin.Context) {
	discountID := c.Param("id")
	if discountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount ID is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userID := c.Query("userId")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	data, total, err := h.svc.GetDiscountUserUsage(
		c.Request.Context(),
		discountID, userID,
		page, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}