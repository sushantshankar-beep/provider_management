package handler

import (
	"net/http"
	"provider_management/internal/domain"
	"provider_management/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AMCPlanHandler struct {
	svc *service.AMCPlanService
}

func NewAMCPlanHandler(svc *service.AMCPlanService) *AMCPlanHandler {
	return &AMCPlanHandler{svc: svc}
}

func (h *AMCPlanHandler) CreateAMC(c *gin.Context) {
	var req domain.AMCPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	adminVal, exists := c.Get("admin")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin ID not found in context",
		})
		return
	}

	admin := adminVal.(*domain.Admin)

	if err := h.svc.CreateAMC(c.Request.Context(), &req, admin.ID.Hex()); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "plan slug already exists" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "AMC plan created successfully",
	})
}

func (h *AMCPlanHandler) GetAllAMC(c *gin.Context) {
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

	isActive := c.Query("is_active")
	planStatus := c.Query("plan_status")
	planVehicleType := c.Query("plan_vehicle_type")
	planCategory := c.Query("plan_category")
	search := c.Query("search")

	plans, counts, pagination, err := h.svc.GetAllAMC(
		c.Request.Context(),
		page,
		limit,
		isActive,
		planStatus,
		planVehicleType,
		planCategory,
		search,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch AMC plans: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "AMC plans fetched successfully",
		"data":       plans,
		"counts":     counts,
		"pagination": pagination,
	})
}

func (h *AMCPlanHandler) GetAMCByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AMC plan ID is required",
		})
		return
	}

	plan, err := h.svc.GetAMCByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "AMC plan not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch AMC plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "AMC plan fetched successfully",
		"data":    plan,
	})
}

func (h *AMCPlanHandler) UpdateAMC(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AMC plan ID is required",
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

	adminVal, exists := c.Get("admin")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin ID not found in context",
		})
		return
	}

	admin := adminVal.(*domain.Admin)

	if err := h.svc.UpdateAMC(c.Request.Context(), id, updateData, admin.ID.Hex()); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "AMC plan not found",
			})
			return
		}

		statusCode := http.StatusInternalServerError
		if err.Error() == "plan slug already exists" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "AMC plan updated successfully",
	})
}

func (h *AMCPlanHandler) DeleteAMC(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AMC plan ID is required",
		})
		return
	}

	if err := h.svc.DeleteAMC(c.Request.Context(), id); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "AMC plan not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete AMC plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "AMC plan deleted successfully",
	})
}

func (h *AMCPlanHandler) ToggleAMCStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AMC plan ID is required",
		})
		return
	}

	adminVal, exists := c.Get("admin")
	
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin ID not found in context",
		})
		return
	}

	admin := adminVal.(*domain.Admin)

	newStatus, err := h.svc.ToggleAMCStatus(c.Request.Context(), id, admin.ID.Hex())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "AMC plan not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to toggle AMC plan status: " + err.Error(),
		})
		return
	}

	message := "AMC plan deactivated successfully"
	if newStatus {
		message = "AMC plan activated successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": message,
		"data": gin.H{
			"is_active": newStatus,
		},
	})
}

func (h *AMCPlanHandler) GetAMCPlansByCategory(c *gin.Context) {
	vehicleType := c.Param("vehicleType")
	category := c.Param("category")
	cityName := c.Query("city_name")

	if vehicleType == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Vehicle type and category are required",
		})
		return
	}

	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "City name is required",
		})
		return
	}

	plans, err := h.svc.GetAMCPlansByCategory(
		c.Request.Context(),
		vehicleType,
		category,
		cityName,
	)

	if err != nil {
		statusCode := http.StatusNotFound
		if err.Error() == "city name is required" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "AMC plans fetched successfully",
		"data":    plans,
	})
}
