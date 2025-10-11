package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

type UserHandler struct {
	UserUsecase usecase.UserUsecase
}

type RegisterRequest struct {
	Username string `josn:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	token, err := h.UserUsecase.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		// 後でエラーの種類に応じてHTTPステータスコードを変えて返却するロジックを追加する
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
	})

}
