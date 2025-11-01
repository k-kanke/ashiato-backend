package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

type NotificationHandler struct {
	notificationUsecase usecase.NotificationUsecase
}

func NewNotificationHandler(uc usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: uc}
}

type notificationListResponse struct {
	Notifications []usecase.NotificationDTO `json:"notifications"`
}

type unreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	limitParam := c.DefaultQuery("limit", "")
	limit := 0
	if limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			limit = parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	var before *time.Time
	if beforeParam := c.Query("before"); beforeParam != "" {
		parsed, err := time.Parse(time.RFC3339, beforeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before parameter"})
			return
		}
		before = &parsed
	}

	notifications, err := h.notificationUsecase.ListNotifications(userID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtoList := usecase.MapNotificationsToDTOs(notifications)
	c.JSON(http.StatusOK, notificationListResponse{Notifications: dtoList})
}

func (h *NotificationHandler) CountUnread(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	count, err := h.notificationUsecase.CountUnreadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, unreadCountResponse{UnreadCount: count})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	notificationID := c.Param("notification_id")

	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification_id is required"})
		return
	}

	if err := h.notificationUsecase.MarkNotificationAsRead(userID, notificationID); err != nil {
		switch err {
		case usecase.ErrNotificationAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case usecase.ErrNotificationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}
