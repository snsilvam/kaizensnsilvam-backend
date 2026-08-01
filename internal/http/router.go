package router

import (
	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/handlers"
)

func New(familyHandler *handlers.FamilyHandler, userHandler *handlers.UserHandler) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handlers.Health)

	families := r.Group("/families")
	{
		families.POST("", familyHandler.Create)
		families.GET("/:id", familyHandler.GetByID)
	}

	users := r.Group("/users")
	{
		users.POST("", userHandler.Create)
		users.GET("/:id", userHandler.GetByID)
	}

	return r
}
