package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS responde los headers de Access-Control para que el frontend pueda
// consumir la API desde otro origen (p.ej. http://localhost:5173 en dev).
//
// allowedOrigins es la lista blanca de orígenes. Si contiene "*" se refleja
// cualquier origen. Se refleja el Origin recibido en vez de devolver "*"
// porque con credentials el navegador rechaza el comodín.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	wildcard := false
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			wildcard = true
			continue
		}
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" && (wildcard || allowed[origin]) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			h.Set("Access-Control-Max-Age", "3600")
			// El cache intermedio debe variar por origen, si no puede servir
			// la respuesta de un origen a otro.
			h.Add("Vary", "Origin")
		}

		// El preflight no llega a ningún handler: se corta acá con 204.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
