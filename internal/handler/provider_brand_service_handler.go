package handler

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"provider_management/internal/service"
)

type ProviderBrandServiceHandler struct {
	svc *service.ProviderBrandService
}

func NewProviderBrandServiceHandler(svc *service.ProviderBrandService) *ProviderBrandServiceHandler {
	return &ProviderBrandServiceHandler{
		svc: svc,
	}
}

func (h *ProviderBrandServiceHandler) GetProviderBrands(c *gin.Context) {
	vehicle := strings.ToLower(c.Query("vehicle"))
	if vehicle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle required"})
		return
	}

	data, err := h.svc.GetBrands(c.Request.Context(), vehicle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"brands": data})
}

func (h *ProviderBrandServiceHandler) GetProviderServices(c *gin.Context) {
	vehicle := c.Query("vehicle")
	if vehicle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle required"})
		return
	}

	data, err := h.svc.GetServices(c.Request.Context(), vehicle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": data})
}

// 	if val, ok := m[key]; ok {
// 		if str, ok := val.(string); ok {
// 			return str
// 		}
// 	}
// 	return ""
// }