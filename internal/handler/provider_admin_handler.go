package handler

import (
	"github.com/gin-gonic/gin"

	"net/http"
	"provider_management/internal/service"
)

type ProviderAdminHandler struct {
	svc *service.ProviderAdminService
}

func NewProviderAdminHandler(svc *service.ProviderAdminService) *ProviderAdminHandler {
	return &ProviderAdminHandler{svc: svc}
}

func (h *ProviderAdminHandler) GetAll(c *gin.Context) {

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "-createdAt")
	search := c.Query("search")
	status := c.Query("status")
	name := c.Query("name")
	mobile := c.Query("mobile")
	providerID := c.Query("providerId")
	kycStatus := c.Query("kycStatus")
	accountStatus := c.Query("accountStatus")
	vehicleType := c.Query("vehicleType")
	zone := c.Query("zone")
	startDate := c.Query("startDate")
	filter := c.Query("filter")
	res, err := h.svc.GetAllProviders(
		c.Request.Context(),
		page, limit, sort, search, status, name, mobile,
		providerID, kycStatus, accountStatus, vehicleType, zone, startDate, filter,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch providers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Providers fetched successfully",
		"data":    res,
	})
}
func (h *ProviderAdminHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	res, err := h.svc.GetProviderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Provider not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Provider fetched successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	res, err := h.svc.UpdateProviderStatus(c.Request.Context(), id, body.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Provider status updated successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) UpdateKYC(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		KYCStatus string `json:"kycStatus"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	res, err := h.svc.UpdateProviderKYC(c.Request.Context(), id, body.KYCStatus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "KYC Status Updated",
		"data":    gin.H{"kycStatus": res.Status},
	})
}

func (h *ProviderAdminHandler) VerifyDocument(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		DocumentType string `json:"documentType"`
		DocumentID   string `json:"documentId"`
		Action       string `json:"action"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	res, err := h.svc.VerifyDocument(
		c.Request.Context(),
		id, body.DocumentType, body.DocumentID, body.Action,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	actionMessage := "approved"
	if body.Action == "reject" {
		actionMessage = "rejected"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Document " + actionMessage + " successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) UpdateAccountAction(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Action string `json:"action"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	res, err := h.svc.UpdateProviderAccountAction(c.Request.Context(), id, body.Action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	messages := map[string]string{
		"activate":  "Provider activated successfully",
		"suspend":   "Provider suspended successfully",
		"blacklist": "Provider blacklisted successfully",
	}

	message := messages[body.Action]
	if message == "" {
		message = "Action completed successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": message,
		"data":    res,
	})
}
func (h *ProviderAdminHandler) UpdateCommission(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		CommissionPercentage float64 `json:"commissionPercentage"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	res, err := h.svc.UpdateProviderCommission(c.Request.Context(), id, body.CommissionPercentage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Commission updated successfully",
		"data":    gin.H{"commissionPercentage": res.CommissionPercentage},
	})
}
