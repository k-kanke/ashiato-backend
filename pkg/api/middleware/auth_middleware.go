// /pkg/api/middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/shared"
)

// AuthMiddleware はJWT認証を検証するミドルウェア
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		secret := os.Getenv("JWT_SECRET")

		// 1. ヘッダーからトークンを抽出
		tokenString, err := shared.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token format"})
			c.Abort()
			return
		}

		// 2. トークンを検証し、UserIDを取得
		userID, err := shared.ParseToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 3. 検証成功: UserIDをコンテキストに格納し、次のハンドラーへ
		c.Set("user_id", userID)
		c.Next()
	}
}

// GetUserIDFromContext はコンテキストからUserIDを取得するヘルパー関数
func GetUserIDFromContext(c *gin.Context) string {
	// AuthMiddlewareで設定されたuser_idを安全に取得
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
