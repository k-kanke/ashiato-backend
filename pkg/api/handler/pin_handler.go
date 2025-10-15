// /pkg/api/handler/pin_handler.go

package handler

import (
	"errors"
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

type GetPinsRequest struct {
	NeLat          float64 `form:"ne_lat" binding:"required"` // 北東 緯度
	NeLng          float64 `form:"ne_lng" binding:"required"` // 北東 経度
	SwLat          float64 `form:"sw_lat" binding:"required"` // 南西 緯度
	SwLng          float64 `form:"sw_lng" binding:"required"` // 南西 経度
	PrivacySetting string  `form:"privacy"`                   // 表示する公開設定（デフォルト: public）
}

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
		switch {
		case errors.Is(err, usecase.ErrInvalidPinCoordinates):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrPinLocationDeviation):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pin"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pin created successfully", "pin": pin})
}

// GetPins は GET /v1/pins のハンドラー
func (h *PinHandler) GetPins(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	var req GetPinsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	privacy := req.PrivacySetting
	if privacy == "" {
		privacy = "public"
	}

	pins, err := h.PinUsecase.GetPinsForMap(
		userID,
		req.SwLat, // 最小緯度
		req.NeLat, // 最大緯度
		req.SwLng, // 最小経度
		req.NeLng, // 最大経度
		privacy,
	)

	if err != nil {
		// ... エラー処理
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pins"})
		return
	}

	// 3. 成功レスポンスの返却
	c.JSON(http.StatusOK, gin.H{"pins": pins})
}
