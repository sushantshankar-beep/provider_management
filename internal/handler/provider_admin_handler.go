package handler

import (
	"github.com/gin-gonic/gin"
    "fmt"
	"net/http"
	"provider_management/internal/middleware"
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

	zoneFilter := middleware.GetZoneFilter(c)

	res, err := h.svc.GetAllProviders(
		c.Request.Context(),
		page, limit, sort, search, status, name, mobile,
		providerID, kycStatus, accountStatus, vehicleType, zone, startDate, filter,zoneFilter,
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

func (h *ProviderAdminHandler) DownloadDocument(c *gin.Context) {
	id := c.Param("id")
	documentType := c.Query("documentType") 
	documentID := c.Query("documentId")  

	if documentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "documentType query parameter is required (identity, address, cancel_cheque)",
		})
		return
	}

	result, err := h.svc.GetDocumentURL(c.Request.Context(), id, documentType, documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	resp, err := http.Get(result.File)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to download file from storage",
		})
		return
	}
	defer resp.Body.Close()

	filename := fmt.Sprintf("%s_%s.pdf", documentType, documentID)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")

	c.DataFromReader(http.StatusOK, resp.ContentLength, "application/pdf", resp.Body, nil)
}

func (h *ProviderAdminHandler) AddNote(c *gin.Context) {
	providerID := c.Param("id")

	var req service.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Note content is required",
		})
		return
	}

	if req.AddedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "AddedBy field is required",
		})
		return
	}

	if err := h.svc.AddNote(c.Request.Context(), providerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add note: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Note added successfully",
	})
}