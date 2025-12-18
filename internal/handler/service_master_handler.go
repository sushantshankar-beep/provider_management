package handler

import (
	"net/http"
	"provider_management/internal/domain"
	"provider_management/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServiceHandler struct {
	svc *service.ServiceService
}

func NewServiceHandler(svc *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req domain.ServiceMaster
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.svc.CreateService(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Service created successfully",
	})
}

func (h *ServiceHandler) GetServices(c *gin.Context) {
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
	category := c.Query("category")
	vehicleType := c.Query("vehicle_type")
	location := c.Query("location")
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	services, total, totalPages, err := h.svc.GetServices(
		c.Request.Context(),
		page,
		limit,
		search,
		category,
		vehicleType,
		location,
		status,
		sortBy,
		sortOrder,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch services: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Services fetched successfully",
		"data":    services,
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

func (h *ServiceHandler) GetServiceByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Service ID is required",
		})
		return
	}

	service, err := h.svc.GetServiceByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Service not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Service fetched successfully",
		"data":    service,
	})
}

func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Service ID is required",
		})
		return
	}

	var updateData map[string]any
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.svc.UpdateService(c.Request.Context(), id, updateData); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Service not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Service updated successfully",
	})
}

func (h *ServiceHandler) UpdateServiceStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Service ID is required",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.svc.UpdateServiceStatus(c.Request.Context(), id, req.Status); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Service not found",
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
			"message": "Failed to update service status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Service status updated successfully",
	})
}

func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Service ID is required",
		})
		return
	}

	if err := h.svc.DeleteService(c.Request.Context(), id); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Service not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Service deleted successfully",
	})
}

func (h *ServiceHandler) GetServiceStats(c *gin.Context) {
	stats, err := h.svc.GetServiceStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch service stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Service stats fetched successfully",
		"data":    stats,
	})
}