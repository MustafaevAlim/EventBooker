package handlers

import (
	"net/http"
	"strings"

	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/service"
)

func AdminMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			NewErrorResponse(c, http.StatusForbidden, "admin success required")
			return
		}
		c.Next()
	}
}

func AuthMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			NewErrorResponse(c, http.StatusUnauthorized, "authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			NewErrorResponse(c, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			NewErrorResponse(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()

	}
}

func CORSMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
