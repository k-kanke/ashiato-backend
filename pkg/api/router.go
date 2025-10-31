package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/handler"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
)

func SetupRouter(
	userHandler *handler.UserHandler,
	pinHandler *handler.PinHandler,
	friendHandler *handler.FriendHandler,
	commentHandler *handler.CommentHandler,
) *gin.Engine {
	router := gin.Default()

	router.Static("/uploads", "./uploads")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
		protected.PUT("/me/settings", userHandler.UpdateSettings)

		// ピン
		protected.POST("/pins", pinHandler.CreatePin)
		protected.GET("/pins", pinHandler.GetPins)
		protected.GET("/pins/:pin_id/thread", commentHandler.GetThread)
		protected.POST("/pins/:pin_id/comments", commentHandler.PostComment)

		// フレンド関連
		friend := protected.Group("/friends")
		{
			friend.POST("/:user_id/request", friendHandler.RequestFriendship)
			friend.POST("/:user_id/accept", friendHandler.AcceptFriendship)
			friend.GET("", friendHandler.GetFriendsList)
		}
	}

	return router
}
