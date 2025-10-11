package api

import (
	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/handler"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
)

func SetupRouter(userHandler *handler.UserHandler, pinHandler *handler.PinHandler) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		// 認証エンドポイント
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}
	}

	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// プロフィール情報取得
		protected.GET("/me", userHandler.GetProfile)

		// ピンの作成
		protected.POST("/pins", pinHandler.CreatePin)
	}

	return router
}
