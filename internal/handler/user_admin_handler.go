package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"provider_management/internal/dto"
	"provider_management/internal/service"
	"strconv"
)

type UserAdminHandler struct {
	svc *service.UserAdminService
}

func NewUserAdminHandler(svc *service.UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{svc: svc}
}

func (h *UserAdminHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)


	filters := dto.UserFilters{
		Search:        c.Query("search"),
		Status:        c.Query("status"),
		AMCStatus:     c.Query("amcStatus"),
		PlatformUsed:  c.Query("platformUsed"),
		VehicleType:   c.Query("vehicleType"),
		Zone:          c.Query("zone"),
		StartDate:     c.Query("startDate"),
	}

	pagination := dto.UserPagination{
		Page:  page,
		Limit: limit,
	}

	res, err := h.svc.GetAllUsers( c, filters , pagination)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserAdminHandler) GetByID(c *gin.Context) {
	res, err := h.svc.GetUserByID(c, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserAdminHandler) UpdateStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	res, err := h.svc.UpdateUserStatus(c, c.Param("id"), body.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserAdminHandler) AddNote(c *gin.Context) {
	userID := c.Param("id")

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

	if err := h.svc.AddNote(c.Request.Context(), userID, req); err != nil {
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