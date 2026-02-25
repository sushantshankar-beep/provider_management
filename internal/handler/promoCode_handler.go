package handler

import (
	"net/http"
	"math"
	"provider_management/internal/dto"
	"provider_management/internal/service"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type PromoCodeHandler struct {
	svc *service.PromoCodeService
}

func NewPromoCodeHandler(svc *service.PromoCodeService) *PromoCodeHandler {
	return &PromoCodeHandler{svc: svc}
}

func (h *PromoCodeHandler) CreatePromoCode(c *gin.Context) {
	var req dto.CreatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	promo, err := h.svc.CreatePromoCode(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Promo code created successfully",
		"data":    promo,
	})
}

func (h *PromoCodeHandler) GetPromoCodes(c *gin.Context) {
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

	promos, total, totalPages, err := h.svc.GetPromoCodes(
		c.Request.Context(),
		page,
		limit,
		c.Query("search"),
		c.Query("status"),
		c.Query("service_type"),
		c.Query("sort_by"),
		c.DefaultQuery("sort_order", "desc"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch promo codes: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo codes fetched successfully",
		"data":    promos,
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

func (h *PromoCodeHandler) GetPromoCodeByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Promo code ID is required",
		})
		return
	}

	promo, err := h.svc.GetPromoCodeByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Promo code not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch promo code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo code fetched successfully",
		"data":    promo,
	})
}

func (h *PromoCodeHandler) UpdatePromoCode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Promo code ID is required",
		})
		return
	}

	var req dto.UpdatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	updatedBy, _ := c.Get("admin_id")
	updatedByStr, _ := updatedBy.(string)

	promo, err := h.svc.UpdatePromoCode(c.Request.Context(), id, req, updatedByStr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Promo code not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update promo code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo code updated successfully",
		"data":    promo,
	})
}

func (h *PromoCodeHandler) UpdatePromoCodeStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Promo code ID is required",
		})
		return
	}

	var req dto.UpdatePromoCodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	updatedBy, _ := c.Get("admin_id")
	updatedByStr, _ := updatedBy.(string)

	if err := h.svc.UpdatePromoCodeStatus(c.Request.Context(), id, req, updatedByStr); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Promo code not found",
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
			"message": "Failed to update promo code status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo code status updated successfully",
	})
}

func (h *PromoCodeHandler) DeletePromoCode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Promo code ID is required",
		})
		return
	}

	if err := h.svc.DeletePromoCode(c.Request.Context(), id); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Promo code not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete promo code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo code deleted successfully",
	})
}

func (h *PromoCodeHandler) GetPromoCodeStats(c *gin.Context) {
	stats, err := h.svc.GetPromoCodeStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch promo code stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Promo code stats fetched successfully",
		"data":    stats,
	})
}

func (h *PromoCodeHandler) ListPromoCodeUsage(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	data, total, err := h.svc.ListPromoCodeUsage(
		c.Request.Context(),
		page,
		limit,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(200, gin.H{
		"data":        data,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}