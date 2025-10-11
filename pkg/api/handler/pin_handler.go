// /pkg/api/handler/pin_handler.go

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

type PinHandler struct {
	PinUsecase usecase.PinUsecase
}

func NewPinHandler(uc usecase.PinUsecase) *PinHandler {
	return &PinHandler{PinUsecase: uc}
}

// CreatePinRequest は Pin作成リクエストのJSON構造体
type CreatePinRequest struct {
	Latitude       float64 `json:"latitude" binding:"required"`
	Longitude      float64 `json:"longitude" binding:"required"`
	ContentText    string  `json:"content_text" binding:"required"`
	MediaURL       string  `json:"media_url"`
	PrivacySetting string  `json:"privacy_setting" binding:"required,oneof=public friends"`
}

// CreatePin は POST /v1/pins のハンドラー
func (h *PinHandler) CreatePin(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	var req CreatePinRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	pin, err := h.PinUsecase.PostNewPin(
		userID,
		req.Latitude,
		req.Longitude,
		req.ContentText,
		req.MediaURL,
		req.PrivacySetting,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pin"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pin created successfully", "pin": pin})
}
