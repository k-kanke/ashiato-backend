package api

import (
	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/handler"
)

func SetupRouter(userHandler *handler.UserHandler) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		// 認証エンドポイント
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
		}
	}

	return router
}
