package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"provider_management/internal/service"
)

type VehicleBrandHandler struct {
	svc *service.VehicleBrandService
}

func NewVehicleBrandHandler(svc *service.VehicleBrandService) *VehicleBrandHandler {
	return &VehicleBrandHandler{
		svc: svc,
	}
}

func (h *VehicleBrandHandler) GetBrands(c *gin.Context) {
	vehicleType := c.Query("vehicleType")
	brandName := c.Query("brandName")

	brands, err := h.svc.GetVehicleBrands(c, vehicleType, brandName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": brands})
}

func (h *VehicleBrandHandler) GetModels(c *gin.Context) {
	vehicleType := c.Query("vehicleType")
	brandName := c.Query("brandName")

	if vehicleType == "" || brandName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "vehicleType and brandName are required"})
		return
	}

	models, err := h.svc.GetVehicleModels(c, vehicleType, brandName)
	if err != nil {
		if err.Error() == "brand not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Brand not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": models})
}