package handler

import (
	"net/http"
	"strconv"

	"provider_management/internal/domain"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
)

type ComplaintHandler struct {
	complaintService service.ComplaintService
}

func NewComplaintHandler(complaintService service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{
		complaintService: complaintService,
	}
}

// Create godoc
// @Summary Create a new complaint
// @Description Create a new complaint for a booking
// @Tags complaints
// @Accept json
// @Produce json
// @Param complaint body service.CreateComplaintRequest true "Complaint details"
// @Success 201 {object} domain.Complaint
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints [post]
func (h *ComplaintHandler) Create(c *gin.Context) {
	var req service.CreateComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	complaint, err := h.complaintService.CreateComplaint(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, complaint)
}

// GetByID godoc
// @Summary Get complaint by ID
// @Description Get detailed information about a complaint
// @Tags complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID (ObjectID or internal ID)"
// @Success 200 {object} domain.Complaint
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints/{id} [get]
func (h *ComplaintHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	complaint, err := h.complaintService.GetComplaint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, complaint)
}

// GetAll godoc
// @Summary List complaints
// @Description Get a paginated list of complaints with optional filters
// @Tags complaints
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (pending, in_review, resolved)"
// @Param raised_by query string false "Filter by who raised it (user, provider)"
// @Param category query string false "Filter by category"
// @Param search query string false "Search in problem, booking numbers, names"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} ListComplaintsResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints [get]
func (h *ComplaintHandler) GetAll(c *gin.Context) {
	filter := domain.ComplaintFilter{
		Page:  1,
		Limit: 10,
	}

	// Parse query parameters
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

	complaints, total, err := h.complaintService.ListComplaints(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response := ListComplaintsResponse{
		Complaints: complaints,
		Pagination: PaginationResponse{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      total,
			TotalPages: (total + int64(filter.Limit) - 1) / int64(filter.Limit),
		},
	}

	c.JSON(http.StatusOK, response)
}

// PostAssessment godoc
// @Summary Assess a complaint
// @Description Admin assesses a complaint and determines fault, refunds, and payouts
// @Tags complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID"
// @Param assessment body service.AssessComplaintRequest true "Assessment details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints/{id}/assessment [post]
func (h *ComplaintHandler) PostAssessment(c *gin.Context) {
	complaintID := c.Param("id")

	var req service.AssessComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// Get admin ID from context (assuming middleware sets this)
	// You can adjust this based on your auth middleware
	adminID, exists := c.Get("admin_id")
	if exists {
		req.AssessedBy = adminID.(string)
	}

	if err := h.complaintService.AssessComplaint(c.Request.Context(), complaintID, req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Complaint assessed successfully",
	})
}

// UpdateStatus godoc
// @Summary Update complaint status
// @Description Update the status of a complaint
// @Tags complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID"
// @Param status body UpdateStatusRequest true "New status"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints/{id}/status [patch]
func (h *ComplaintHandler) UpdateStatus(c *gin.Context) {
	complaintID := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	adminIDStr := ""
	if exists {
		adminIDStr = adminID.(string)
	}

	if err := h.complaintService.UpdateComplaintStatus(c.Request.Context(), complaintID, req.Status, adminIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Complaint status updated successfully",
	})
}

// AddNote godoc
// @Summary Add a note to complaint
// @Description Add a note/comment to a complaint
// @Tags complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID"
// @Param note body service.AddNoteRequest true "Note content"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /complaints/{id}/notes [post]
func (h *ComplaintHandler) AddNote(c *gin.Context) {
	complaintID := c.Param("id")

	var req service.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if exists {
		req.AddedBy = userID.(string)
	}

	if err := h.complaintService.AddNote(c.Request.Context(), complaintID, req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Note added successfully",
	})
}

// GetStats godoc
// @Summary Get complaint statistics
// @Description Get overall complaint statistics
// @Tags complaints
// @Accept json
// @Produce json
// @Success 200 {object} domain.ComplaintStats
// @Failure 500 {object} ErrorResponse
// @Router /complaints/stats [get]
func (h *ComplaintHandler) GetStats(c *gin.Context) {
	stats, err := h.complaintService.GetComplaintStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Response types
type ListComplaintsResponse struct {
	Complaints []*domain.Complaint `json:"complaints"`
	Pagination PaginationResponse  `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
