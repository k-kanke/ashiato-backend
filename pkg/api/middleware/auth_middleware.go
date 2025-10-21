// /pkg/api/middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/shared"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		secret := os.Getenv("JWT_SECRET")

		tokenString, err := shared.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token format"})
			c.Abort()
			return
		}

		userID, err := shared.ParseToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
