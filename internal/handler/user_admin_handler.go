package handler

import (
	"log"
	"net/http"
	"strconv"

	"provider_management/internal/service"

	"github.com/gin-gonic/gin"
)

type UserAdminHandler struct {
	svc *service.UserAdminService
}

func NewUserAdminHandler(svc *service.UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{svc: svc}
}

func (h *UserAdminHandler) GetAll(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	res, err := h.svc.GetAllUsers(
		c,
		c.Query("search"),
		c.Query("status"),
		c.Query("zone"),
		page,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *UserAdminHandler) GetByID(c *gin.Context) {
	res, err := h.svc.GetUserByID(c, c.Param("id"))
	log.Println("error",err)
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
