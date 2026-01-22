package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"provider_management/internal/domain"
	"provider_management/internal/service"
	"provider_management/internal/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProviderAdminHandler struct {
	svc *service.ProviderAdminService
}

func NewProviderAdminHandler(svc *service.ProviderAdminService) *ProviderAdminHandler {
	return &ProviderAdminHandler{svc: svc}
}

func (h *ProviderAdminHandler) GetAll(c *gin.Context) {

	filters := dto.ProviderFilters{
		Search:        c.Query("search"),
		Status:        c.Query("status"),
		Name:          c.Query("name"),
		Mobile:        c.Query("mobile"),
		ProviderID:    c.Query("providerId"),
		KYCStatus:     c.Query("kycStatus"),
		AccountStatus: c.Query("accountStatus"),
		VehicleType:   c.Query("vehicleType"),
		Zone:          c.Query("zone"),
		StartDate:     c.Query("startDate"),
		Filter:        c.Query("filter"),
	}

	zoneFilter := middleware.GetZoneFilter(c)

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	pagination := dto.ProviderPagination{
		Page:  page,
		Limit: limit,
		Sort:  c.DefaultQuery("sort", "-createdAt"),
	}
    
	res, err := h.svc.GetAllProviders(c.Request.Context(),filters, pagination, zoneFilter)

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

	res, err := h.svc.VerifyDocument( c.Request.Context(), id, body.DocumentType, body.DocumentID, body.Action)

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

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch file from storage",
		})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	extension := ".pdf"
	switch contentType {
	case "image/jpeg", "image/jpg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	case "application/pdf":
		extension = ".pdf"
	}

	filename := fmt.Sprintf("%s_%s%s", documentType, documentID, extension)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache")

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}

	written, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		fmt.Printf("Error copying file: %v", err)
		return
	}

	fmt.Printf("Successfully streamed %d bytes of %s", written, contentType)
}
func (h *ProviderAdminHandler) AddNote(c *gin.Context) {
	providerID := c.Param("id")

	var req dto.AddNoteRequest
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

func (h *ProviderAdminHandler) CreateProvider(c *gin.Context) {
	var req dto.CreateProviderRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Unauthorized",
		})
		return
	}

	if urls, exists := middleware.GetUploadedURLs(c, "profileUrl"); exists && len(urls) > 0 {
		req.ProfileURL = urls[0]
	}

	if identityProofURLs, exists := middleware.GetUploadedURLs(c, "identityProof"); exists {
		req.IdentityProofs = make([]domain.Proof, len(identityProofURLs))
		for i, url := range identityProofURLs {
			req.IdentityProofs[i] = domain.Proof{
				ID:       primitive.NewObjectID(),
				File:     url,
				Type:     "",
				Verified: domain.VerificationPending,
			}
		}
	}

	if addressProofURLs, exists := middleware.GetUploadedURLs(c, "addressProof"); exists {
		req.AddressProofs = make([]domain.Proof, len(addressProofURLs))
		for i, url := range addressProofURLs {
			req.AddressProofs[i] = domain.Proof{
				ID:       primitive.NewObjectID(),
				File:     url,
				Type:     "",
				Verified: domain.VerificationPending,
			}
		}
	}

	if urls, exists := middleware.GetUploadedURLs(c, "cancelCheque"); exists && len(urls) > 0 {
		req.CancelCheque = &domain.CancelCheque{
			File:     urls[0],
			Verified: domain.VerificationPending,
		}
	}

	if _, err := h.svc.CreateProvider( c.Request.Context(), req, admin.ID, admin.Role ); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to create provider: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Provider created successfully",
	})
}

func (h *ProviderAdminHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateProviderRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": true,
		})
		return
	}

	if urls, exists := middleware.GetUploadedURLs(c, "profileUrl"); exists && len(urls) > 0 {
		req.ProfileURL = urls[0]
	}

	identityTypes := c.PostFormArray("identityProofTypes[]")
	addressTypes := c.PostFormArray("addressProofTypes[]")

	if urls, exists := middleware.GetUploadedURLs(c, "identityProof"); exists && len(urls) > 0 {
		req.IdentityProofs = make([]domain.Proof, len(urls))
		for i, url := range urls {
			pt := ""
			if i < len(identityTypes) {
				pt = identityTypes[i]
			}
			req.IdentityProofs[i] = domain.Proof{
				ID:       primitive.NewObjectID(),
				File:     url,
				Type:     pt,
				Verified: domain.VerificationPending,
			}
		}
	}

	if urls, exists := middleware.GetUploadedURLs(c, "addressProof"); exists && len(urls) > 0 {
		req.AddressProofs = make([]domain.Proof, len(urls))
		for i, url := range urls {
			pt := ""
			if i < len(addressTypes) {
				pt = addressTypes[i]
			}
			req.AddressProofs[i] = domain.Proof{
				ID:       primitive.NewObjectID(),
				File:     url,
				Type:     pt,
				Verified: domain.VerificationPending,
			}
		}
	}

	if urls, exists := middleware.GetUploadedURLs(c, "cancelCheque"); exists && len(urls) > 0 {
		req.CancelCheque = &domain.CancelCheque{
			File:     urls[0],
			Verified: domain.VerificationPending,
		}
	}

	if _, err := h.svc.UpdateProvider( c.Request.Context(), id, req, admin.ID ); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Provider updated successfully",
	})
}

func (h *ProviderAdminHandler) GetZoneStats(c *gin.Context) {
	adminZones := middleware.GetAdminZones(c)

	admin := middleware.GetAdminFromContext(c)
	if admin == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	res, err := h.svc.GetZoneStats(c.Request.Context(), adminZones, admin.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch zone statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Zone statistics fetched successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) GetZoneActivationTeam(c *gin.Context) {
	zoneName := c.Param("zone")
	adminZones := middleware.GetAdminZones(c)

	res, err := h.svc.GetZoneActivationTeam(c.Request.Context(), zoneName, adminZones)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch activation team",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activation team fetched successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) GetActivationPersonProviders(c *gin.Context) {
	personID := c.Param("person")
	zoneName := c.Param("zone")

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	filters := dto.ProviderFilters{
		Search:        c.Query("search"),
	}

	pagination := dto.ProviderPagination{
		Page:  page,
		Limit: limit,
		Sort:  c.DefaultQuery("sort", "-createdAt"),
	}

	adminZones := middleware.GetAdminZones(c)

	res, err := h.svc.GetActivationPersonProviders( c.Request.Context(), personID, zoneName, filters, pagination, adminZones)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Providers fetched successfully",
		"data":    res,
	})
}

func (h *ProviderAdminHandler) GetProviderEarnings(c *gin.Context) {
	providerID := c.Query("provider_id")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "provider_id is required",
		})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	sortField := c.Query("sort-field")
	if sortField == "" {
		sortField = "createdAt"
	}

	sortOrder := c.Query("sort-order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	search := c.Query("search")

	earnings, summary, total, err := h.svc.GetProviderEarnings(c.Request.Context(), providerID, page, limit, sortField, sortOrder, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Provider earnings fetched successfully",
		"summary": summary,
		"data":    earnings,
		"pagination": gin.H{
			"page":         page,
			"limit":        limit,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     int64(page) < totalPages,
			"has_previous": page > 1,
		},
	})
}