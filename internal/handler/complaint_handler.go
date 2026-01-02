package handler

import (
	"log"
	"net/http"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"provider_management/internal/service"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

type ComplaintListItem struct {
	ID                string `json:"id"`
	InternalID        string `json:"complaint_id"`
	AcceptedServiceNo string `json:"booking_no"`
	RaisedBy          string `json:"raised_by"`
	Problem           string `json:"category"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}

func (h *ComplaintHandler) GetAll(c *gin.Context) {
	filter := domain.ComplaintFilter{
		Page:  1,
		Limit: 10,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if raisedBy := c.Query("raised_by"); raisedBy != "" {
		filter.RaisedBy = &raisedBy
	}

	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}

	if search := c.Query("search"); search != "" {
		filter.SearchQuery = &search
	}

	if userId := c.Query("userId"); userId != "" {
		filter.UserID = &userId
	}

	if providerId := c.Query("providerId"); providerId != "" {
		filter.ProviderID = &providerId
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &parsedDate
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			parsedDate = parsedDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.DateTo = &parsedDate
		}
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	complaints, total, stats, err := h.complaintService.ListComplaints(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch complaints: " + err.Error(),
		})
		return
	}

	limitedData := make([]ComplaintListItem, len(complaints))
	for i, complaint := range complaints {
		indianTime := complaint.CreatedAt.Add(5*time.Hour + 30*time.Minute)
		limitedData[i] = ComplaintListItem{
			ID:                complaint.ID,
			InternalID:        "CMP" + strconv.FormatInt(complaint.InternalID, 10),
			AcceptedServiceNo: "BK" + strconv.FormatInt(complaint.AcceptedServiceNo, 10),
			RaisedBy:          complaint.RaisedBy,
			Problem:           complaint.Problem,
			Status:            complaint.Status,
			CreatedAt:         indianTime.Format("2006-01-02 15:04:05"),
		}
	}

	totalPages := (total + int64(filter.Limit) - 1) / int64(filter.Limit)

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Complaints fetched successfully",
		"data":    limitedData,
		"stats":   stats,
		"pagination": gin.H{
			"page":         filter.Page,
			"limit":        filter.Limit,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     int64(filter.Page) < totalPages,
			"has_previous": filter.Page > 1,
		},
	})
}
func (h *ComplaintHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	id = strings.TrimPrefix(id, "CMP")

	log.Println("Fetching complaint with ID:", id)

	complaint, err := h.complaintService.GetComplaintWithDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	response := gin.H{
		"_id":               complaint.ID,
		"complaint_id":      "CMP" + strconv.FormatInt(complaint.InternalID, 10),
		"booking_id":        complaint.AcceptedServiceID,
		"booking_no":        "BK" + strconv.FormatInt(complaint.AcceptedServiceNo, 10),
		"user_id":           complaint.UserID,
		"provider_id":       complaint.ProviderID,
		"raised_by":         complaint.RaisedBy,
		"problem":           complaint.Problem,
		"photos":            complaint.Photos,
		"status":            complaint.Status,
		"timeline":          complaint.Timeline,
		"assessment":        complaint.Assessment,
		"notes":             complaint.Notes,
		"actions_triggered": complaint.ActionsTriggered,
		"created_at":        complaint.CreatedAt,
		"updated_at":        complaint.UpdatedAt,
		"updated_by_admin":  complaint.UpdatedByAdmin,
		"admin_updated_at":  complaint.AdminUpdatedAt,
		"payment_tracking":  complaint.PaymentTracking,
	}

	if complaint.UserDetails != nil {
		response["user_details"] = complaint.UserDetails
	}

	if complaint.ProviderDetails != nil {
		response["provider_details"] = complaint.ProviderDetails
	}

	if complaint.BookingDetails != nil {
		response["booking_details"] = complaint.BookingDetails
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Complaint fetched successfully",
		"data":    response,
	})
}

func (h *ComplaintHandler) PostAssessment(c *gin.Context) {
	id := c.Param("id")
	id = strings.TrimPrefix(id, "CMP")

	log.Println("Assessing complaint with ID:", id)

	var req service.AssessComplaintRequest
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

	complaint, err := h.complaintService.GetComplaint(c.Request.Context(), id)
	if err != nil {
		log.Println("Failed to fetch complaint:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Complaint not found",
		})
		return
	}

	acceptedService, err := h.acceptedServiceRepo.FindByID(c.Request.Context(), complaint.AcceptedServiceID)
	if err != nil {
		log.Println("Failed to fetch accepted service:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Accepted service not found",
		})
		return
	}

	if acceptedService.OrderID == "" {
		log.Println("No OrderID found in accepted service")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "No transaction associated with this booking",
		})
		return
	}

	req.TxnID = acceptedService.OrderID
	log.Printf("Using TxnID from accepted service: %s", req.TxnID)

	if err := h.complaintService.AssessComplaint(c.Request.Context(), id, req); err != nil {
		log.Println("Assessment error:", err)
	
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

	var req struct {
		Status string `json:"status" binding:"required"`
	}
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

func (h *ComplaintHandler) AddNote(c *gin.Context) {
	complaintID := c.Param("id")

	complaintID = strings.TrimPrefix(complaintID, "CMP")
	internalID, err := strconv.ParseInt(complaintID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid complaint id",
		})
		return
	}
	log.Println("complaintID", complaintID)
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

	if err := h.complaintService.AddNote(c.Request.Context(), internalID, req); err != nil {
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
		"status":  "in_review",
	})
}