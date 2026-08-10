package middleware

import (
	"context"
	"net/http"
	"strings"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
)

// TokenVerifier es la porción del cliente de Firebase Auth que necesita el
// middleware. *fbauth.Client la satisface; en tests se inyecta un doble.
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*fbauth.Token, error)
}

// Auth verifica el ID token de Firebase que llega en
// `Authorization: Bearer <ID_TOKEN>` y deja el UID autenticado en el context
// del request, de donde lo leen los handlers con auth.UID().
//
// Un token ausente, malformado, expirado o inválido corta la cadena con 401.
// El middleware no conoce el dominio: sólo autentica y publica el UID.
func Auth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		idToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortUnauthorized(c, "missing or malformed Authorization header")
			return
		}

		token, err := verifier.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			// El motivo real no se filtra al cliente para no dar pistas sobre
			// la validación; queda en el log de Gin si hace falta depurar.
			abortUnauthorized(c, "invalid ID token")
			return
		}

		if token.UID == "" {
			// Firebase siempre emite un `sub`, pero un UID vacío no identifica
			// a nadie. Se corta acá para que ninguna ruta protegida llegue al
			// handler sin usuario.
			abortUnauthorized(c, "invalid ID token")
			return
		}

		// El UID viaja en el context del request, no en el de Gin: es el mismo
		// que los handlers ya pasan a los services, así que el dominio lo lee
		// sin conocer Gin ni el SDK de Firebase.
		c.Request = c.Request.WithContext(auth.WithUID(c.Request.Context(), token.UID))
		c.Next()
	}
}

// bearerToken extrae el token del header Authorization. El esquema se compara
// sin distinguir mayúsculas porque RFC 7235 lo define case-insensitive.
func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": message})
}
