package handler

import (
	"net/http"
	"strings"

	"provider_management/internal/constants"
	"provider_management/internal/logger"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	log *logger.Logger
}

func NewVehicleHandler(log *logger.Logger) *VehicleHandler {
	return &VehicleHandler{
		log: log,
	}
}

type BrandResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Logo        string `json:"logo"`
	VehicleType string `json:"vehicleType"`
	BrandType   string `json:"brandType"`
}

type ServiceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	VehicleType string `json:"vehicleType"`
}

func (h *VehicleHandler) GetVehicleBrands(c *gin.Context) {
	vehicleType := c.Query("vehicleType")

	filteredBrands := []map[string]interface{}{}

	if vehicleType != "" {
		for _, brand := range constants.BrandData {
			if brandVehicleType, ok := brand["vehicleType"].(string); ok {
				if strings.EqualFold(brandVehicleType, vehicleType) {
					filteredBrands = append(filteredBrands, brand)
				}
			}
		}
	} else {
		filteredBrands = constants.BrandData
	}

	formattedBrands := []BrandResponse{}
	for _, brand := range filteredBrands {
		formattedBrand := BrandResponse{
			ID:          getString(brand, "_id"),
			Name:        getString(brand, "name"),
			Slug:        getString(brand, "slug"),
			Logo:        getString(brand, "logo"),
			VehicleType: getString(brand, "vehicleType"),
			BrandType:   getString(brand, "brandType"),
		}
		formattedBrands = append(formattedBrands, formattedBrand)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Brands fetched successfully",
		"data":    formattedBrands,
	})
}

func (h *VehicleHandler) GetVehicleServices(c *gin.Context) {
	vehicleType := c.Query("vehicleType")

	filteredServices := []map[string]interface{}{}

	if vehicleType != "" {
		for _, service := range constants.ServicesData {
			if serviceVehicleType, ok := service["vehicleType"].(string); ok {
				if strings.EqualFold(serviceVehicleType, vehicleType) {
					filteredServices = append(filteredServices, service)
				}
			}
		}
	} else {
		filteredServices = constants.ServicesData
	}

	formattedServices := []ServiceResponse{}
	for _, service := range filteredServices {
		formattedService := ServiceResponse{
			ID:          getString(service, "_id"),
			Name:        getString(service, "name"),
			Logo:        getString(service, "logo"),
			VehicleType: getString(service, "vehicleType"),
		}
		formattedServices = append(formattedServices, formattedService)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Services fetched successfully",
		"data":    formattedServices,
	})
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}