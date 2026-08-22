package router

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	openapi "github.com/snsilvam/kaizensnsilvam-backend/docs"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/handlers"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/http/middleware"
)

//go:embed swagger/index.html
var swaggerUI []byte

// New arma el router. auth es el middleware de autenticación que protege las
// rutas de negocio; /health, / y la documentación quedan públicas para que el
// health check de Cloud Run y Swagger sigan funcionando sin token.
func New(familyHandler *handlers.FamilyHandler, userHandler *handlers.UserHandler, incomeHandler *handlers.IncomeHandler, pendingPaymentHandler *handlers.PendingPaymentHandler, habit1Handler *handlers.Habit1Handler, dashboardHandler *handlers.DashboardHandler, auth gin.HandlerFunc, allowedOrigins []string) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(allowedOrigins))

	// Sin esto Gin responde 404 al preflight de rutas con método no registrado
	// y el navegador nunca ve los headers de CORS.
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })

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

	// A partir de acá todo exige `Authorization: Bearer <ID_TOKEN>`.
	// El middleware va después de CORS para que un 401 también lleve los
	// headers de Access-Control y el navegador pueda leer la respuesta.
	families := r.Group("/families", auth)
	{
		families.POST("", familyHandler.Create)
		families.GET("/:id", familyHandler.GetByID)
	}

	users := r.Group("/users", auth)
	{
		users.POST("", userHandler.Create)
		users.GET("/:id", userHandler.GetByID)
	}

	incomes := r.Group("/incomes", auth)
	{
		incomes.GET("", incomeHandler.List)
		incomes.POST("", incomeHandler.Register)
		incomes.DELETE("/:id", incomeHandler.Delete)
	}

	pendingPayments := r.Group("/pending-payments", auth)
	{
		pendingPayments.GET("", pendingPaymentHandler.GetAll)
		pendingPayments.POST("", pendingPaymentHandler.Register)
		pendingPayments.PATCH("/:id/mark-as-paid", pendingPaymentHandler.MarkAsPaid)
		pendingPayments.DELETE("/:id", pendingPaymentHandler.Delete)
	}

	habit1Records := r.Group("/habit-1", auth)
	{
		habit1Records.GET("", habit1Handler.List)
		habit1Records.POST("", habit1Handler.Register)
	}

	r.GET("/dashboard", auth, dashboardHandler.Get)

	return r
}
