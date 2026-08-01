package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/family"
)

// FamilyHandler traduce HTTP <-> dominio para Family.
type FamilyHandler struct {
	svc *family.Service
}

// NewFamilyHandler construye el handler inyectando el service.
func NewFamilyHandler(svc *family.Service) *FamilyHandler {
	return &FamilyHandler{svc: svc}
}

type createFamilyRequest struct {
	Name string `json:"name" binding:"required"`
}

// Create maneja POST /families
func (h *FamilyHandler) Create(c *gin.Context) {
	var req createFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	f, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		if errors.Is(err, family.ErrInvalidName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, f)
}

// GetByID maneja GET /families/:id
func (h *FamilyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	f, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, family.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, f)
}
