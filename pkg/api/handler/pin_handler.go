package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	contentType := c.GetHeader("Content-Type")

	var latitude float64
	var longitude float64
	var contentText string
	var privacySetting string
	var mediaURL string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		latStr := strings.TrimSpace(c.PostForm("latitude"))
		lngStr := strings.TrimSpace(c.PostForm("longitude"))
		if latStr == "" || lngStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
			return
		}

		var err error
		latitude, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
			return
		}

		longitude, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
			return
		}

		contentText = strings.TrimSpace(c.PostForm("content_text"))
		if contentText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content_text is required"})
			return
		}

		privacySetting = strings.TrimSpace(c.PostForm("privacy_setting"))
		if privacySetting == "" {
			privacySetting = "public"
		}

		if privacySetting != "public" && privacySetting != "friends" {
			privacySetting = "public"
		}

		if fileHeader, err := c.FormFile("image"); err == nil {
			savedURL, saveErr := saveUploadedImage(c, fileHeader, "pins")
			if saveErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": saveErr.Error()})
				return
			}
			mediaURL = savedURL
		}
	} else {
		var req CreatePinRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}
		latitude = req.Latitude
		longitude = req.Longitude
		contentText = strings.TrimSpace(req.ContentText)
		if contentText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content_text is required"})
			return
		}
		privacySetting = req.PrivacySetting
		mediaURL = strings.TrimSpace(req.MediaURL)
		if privacySetting != "public" && privacySetting != "friends" {
			privacySetting = "public"
		}
	}

	pin, err := h.PinUsecase.PostNewPin(
		userID,
		latitude,
		longitude,
		contentText,
		mediaURL,
		privacySetting,
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
