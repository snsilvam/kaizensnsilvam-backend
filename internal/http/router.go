package router

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	openapi "github.com/snsilvam/kaizensnsilvam-backend/docs"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/handlers"
)

//go:embed swagger/index.html
var swaggerUI []byte

func New(familyHandler *handlers.FamilyHandler, userHandler *handlers.UserHandler, incomeHandler *handlers.IncomeHandler) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handlers.Health)
	swagger := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUI)
	}
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger")
	})
	r.GET("/swagger", swagger)
	r.GET("/swagger/", swagger)
	r.GET("/docs/swagger.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.SwaggerYAML)
	})

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

	incomes := r.Group("/incomes")
	{
		incomes.POST("", incomeHandler.Register)
	}

	return r
}
