package repository

import (
	"time"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
)

type NotificationRepository interface {
	// 通知を作成する
	Create(notification *domain.Notification) error
	ListByRecipient(userID string, limit int, before *time.Time) ([]domain.Notification, error)
	CountUnread(userID string) (int, error)
	MarkAsRead(notificationID string) error
	FindByID(notificationID string) (*domain.Notification, error)
}
