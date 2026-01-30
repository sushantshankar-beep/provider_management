package handler

import (
	"time"
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"provider_management/internal/domain"
	"provider_management/internal/service"
	"provider_management/internal/repository"
)

type ComplaintHandler struct {
	complaintService *service.ComplaintService
	acceptedServiceRepo *repository.AcceptedServiceRepo
}

func NewComplaintHandler(complaintService *service.ComplaintService,acceptedServiceRepo *repository.AcceptedServiceRepo) *ComplaintHandler {
	return &ComplaintHandler{
		complaintService: complaintService,
		acceptedServiceRepo: acceptedServiceRepo,
	}
}

func (h *ComplaintHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filters := dto.ComplaintFilters{}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if raisedBy := c.Query("raised_by"); raisedBy != "" {
		filters.RaisedBy = &raisedBy
	}
	if category := c.Query("category"); category != "" {
		filters.Category = &category
	}
	if search := c.Query("search"); search != "" {
		filters.SearchQuery = &search
	}
	if userId := c.Query("userId"); userId != "" {
		filters.UserID = &userId
	}
	if providerId := c.Query("providerId"); providerId != "" {
		filters.ProviderID = &providerId
	}
	if createdAt := c.Query("createdAt"); createdAt != "" {
		if parsedDate, err := time.Parse("2006-01-02", createdAt); err == nil {
			start := parsedDate
			end := parsedDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.CreatedAtFrom = &start
			filters.CreatedAtTo = &end
		}
	}

	pagination := dto.ComplaintPagination{
		Page:  page,
		Limit: limit,
	}

	res, err := h.complaintService.GetAllComplaints(c.Request.Context(), filters, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch complaints: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "Complaints fetched successfully",
		"data":       res.Data,
		"stats":      res.Stats,
		"pagination": res.Pagination,
	})
}

func (h *ComplaintHandler) GetByID(c *gin.Context) {

	id := c.Param("id")

    res, err := h.complaintService.GetComplaintWithDetails(c.Request.Context(), id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Complaint fetched successfully",
		"data":    res,
	})
}

func (h *ComplaintHandler) PostAssessment(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssessComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.FaultParty == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "fault_party is required",
		})
		return
	}

	if req.RefundToUser == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "refund_to_user is required",
		})
		return
	}

	if req.PayoutToProvider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "payout_to_provider is required",
		})
		return
	}

	complaint, err := h.complaintService.GetComplaintByNumber(c.Request.Context(), id)
	if err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Complaint not found",
		})
		return
	}

	acceptedService, err := h.acceptedServiceRepo.FindByID(c.Request.Context(), complaint.AcceptedService)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Accepted service not found",
		})
		return
	}

	if acceptedService.PaymentStatus != "paid" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Payment not completed for this booking",
		})
		return
	}

	if err := h.complaintService.AssessComplaint(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Complaint assessed successfully",
	})
}

func (h *ComplaintHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	id = strings.TrimPrefix(id, "CMP")

    var req dto.UpdateComplaintStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	adminID, exists := c.Get("admin_id")
	adminIDStr := ""
	if exists {
		adminIDStr = adminID.(string)
	}

	if err := h.complaintService.UpdateComplaintStatus(c.Request.Context(), id, req.Status, adminIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Complaint status updated successfully",
	})
}

func (h *ComplaintHandler) GetStats(c *gin.Context) {
	stats, err := h.complaintService.GetComplaintStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Stats fetched successfully",
		"data":    stats,
	})
}

func (h *ComplaintHandler) StartAssessment(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.complaintService.StartAssessment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Assessment saved successfully and complaint reviewed",
		"status":  domain.ComplaintStatusInReview,
	})
}

func (h *ComplaintHandler) AddNote(c *gin.Context) {

	complaintID := c.Param("id")

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

	if err := h.complaintService.AddNote(c.Request.Context(), complaintID, req); err != nil {
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
