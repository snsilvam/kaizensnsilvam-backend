package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/user"
)

// UserHandler traduce HTTP <-> dominio para User.
type UserHandler struct {
	svc *user.Service
}

// NewUserHandler construye el handler inyectando el service.
func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

type createUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// Create maneja POST /users
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Create(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrInvalidName) || errors.Is(err, user.ErrInvalidEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, u)
}

// GetByID maneja GET /users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}
