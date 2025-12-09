package handler

import (
	"net/http"

	"provider_management/internal/domain"
	"provider_management/internal/logger"
	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ComplaintHandler struct {
	svc       *service.ComplaintService
	log       *logger.Logger
	validator *validator.Validate
}

func NewComplaintHandler(svc *service.ComplaintService, log *logger.Logger) *ComplaintHandler {
	return &ComplaintHandler{
		svc:       svc,
		log:       log,
		validator: validator.New(),
	}
}

func (h *ComplaintHandler) Router() *gin.Engine {
	r := gin.Default()

	r.GET("/complaints", h.GetAll)
	r.GET("/complaints/:id", h.GetByID)
	r.POST("/complaints/:id/assessment", h.PostAssessment)

	return r
}

func (h *ComplaintHandler) GetAll(c *gin.Context) {
	res, err := h.svc.ListComplaints(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ComplaintHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetComplaint(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ComplaintHandler) PostAssessment(c *gin.Context) {
	id := c.Param("id")

	var req domain.Assessment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.PostAssessment(c, id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "assessment saved"})
}
