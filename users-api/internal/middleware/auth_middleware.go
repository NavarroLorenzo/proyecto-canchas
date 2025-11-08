package middleware

import (
	"net/http"
	"strings"
	"users-api/internal/dto"
	"users-api/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida el token JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Authorization header required",
			})
			c.Abort()
			return
		}

		// El formato esperado es: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token)

		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Invalid token",
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		// Guardar información del usuario en el contexto
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware valida que el usuario sea administrador
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")

		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
