package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"provider_management/internal/dto"
	"provider_management/internal/service"
)

type AgreementHandler struct {
	svc *service.AgreementService
}

func NewAgreementHandler(svc *service.AgreementService) *AgreementHandler {
	return &AgreementHandler{
		svc: svc,
	}
}

func (h *AgreementHandler) GetAgreement(c *gin.Context) {
	id := c.Param("id")

	safeHTML := c.Query("safe") == "true"

	if safeHTML {
		agreement, err := h.svc.GetAgreementSafeHTML(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, agreement)
		return
	}

	agreement, err := h.svc.GetAgreement(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, agreement)
}

func (h *AgreementHandler) CreateAgreement(c *gin.Context) {
	var req dto.CreateAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agreement, err := h.svc.CreateAgreement(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agreement"})
		return
	}
	c.JSON(http.StatusCreated, agreement)
}

func (h *AgreementHandler) UpdateAgreement(c *gin.Context) {
	
	id := c.Param("id")
	var req dto.UpdateAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agreement, err := h.svc.UpdateAgreement(c, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update agreement"})
		return
	}
	c.JSON(http.StatusOK, agreement)
}
