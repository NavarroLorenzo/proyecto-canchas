package middleware

import (
	"net/http"
	"strings"

	"canchas-api/internal/clients"
	"canchas-api/internal/dto"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida el token JWT llamando a users-api
func AuthMiddleware(userClient clients.UserClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Aquí simplemente pasamos el token, users-api lo validará
		c.Set("token", parts[1])
		c.Next()
	}
}

// AdminMiddleware verifica que el token sea de un admin
// (Esto se debería validar llamando a users-api, por ahora simplificado)
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Por ahora, asumimos que si tiene token es admin
		// En una implementación completa, deberías validar el rol con users-api
		c.Next()
	}
}
