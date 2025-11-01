package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

const (
	NotificationTypeFriendRequestReceived = "friend_request_received"
	NotificationTypeFriendRequestAccepted = "friend_request_accepted"

	defaultNotificationListLimit = 20
	maxNotificationListLimit     = 100
)

var (
	ErrInvalidRecipientID       = errors.New("recipient id must be provided")
	ErrInvalidActorID           = errors.New("actor id must be provided")
	ErrInvalidNotificationID    = errors.New("notification id must be provided")
	ErrNotificationAccessDenied = errors.New("notification does not belong to user")
	ErrNotificationNotFound     = errors.New("notification not found")
)

type NotificationUsecase interface {
	CreateFriendRequestNotification(recipientID, actorID string) (*domain.Notification, error)
	CreateFriendAcceptedNotification(recipientID, actorID string) (*domain.Notification, error)
	ListNotifications(userID string, limit int, before *time.Time) ([]domain.Notification, error)
	CountUnreadNotifications(userID string) (int, error)
	MarkNotificationAsRead(userID, notificationID string) error
}

type notificationUsecase struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationUsecase(repo repository.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{notificationRepo: repo}
}

func (u *notificationUsecase) CreateFriendRequestNotification(recipientID, actorID string) (*domain.Notification, error) {
	return u.createNotification(recipientID, actorID, NotificationTypeFriendRequestReceived, actorID)
}

func (u *notificationUsecase) CreateFriendAcceptedNotification(recipientID, actorID string) (*domain.Notification, error) {
	return u.createNotification(recipientID, actorID, NotificationTypeFriendRequestAccepted, actorID)
}

func (u *notificationUsecase) ListNotifications(userID string, limit int, before *time.Time) ([]domain.Notification, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidRecipientID
	}

	resolvedLimit := limit
	switch {
	case resolvedLimit <= 0:
		resolvedLimit = defaultNotificationListLimit
	case resolvedLimit > maxNotificationListLimit:
		resolvedLimit = maxNotificationListLimit
	}

	notifications, err := u.notificationRepo.ListByRecipient(userID, resolvedLimit, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	return notifications, nil
}

func (u *notificationUsecase) CountUnreadNotifications(userID string) (int, error) {
	if strings.TrimSpace(userID) == "" {
		return 0, ErrInvalidRecipientID
	}

	count, err := u.notificationRepo.CountUnread(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

func (u *notificationUsecase) MarkNotificationAsRead(userID, notificationID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidRecipientID
	}
	if strings.TrimSpace(notificationID) == "" {
		return ErrInvalidNotificationID
	}

	notification, err := u.notificationRepo.FindByID(notificationID)
	if err != nil {
		return fmt.Errorf("failed to load notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}
	if notification.RecipientUserID != userID {
		return ErrNotificationAccessDenied
	}

	if notification.IsRead {
		return nil
	}

	if err := u.notificationRepo.MarkAsRead(notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (u *notificationUsecase) createNotification(recipientID, actorID, notificationType, relatedEntityID string) (*domain.Notification, error) {
	if strings.TrimSpace(recipientID) == "" {
		return nil, ErrInvalidRecipientID
	}
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrInvalidActorID
	}
	if strings.TrimSpace(notificationType) == "" {
		return nil, errors.New("notification type must be provided")
	}

	now := time.Now().UTC()
	notification := &domain.Notification{
		NotificationID:  uuid.New().String(),
		RecipientUserID: recipientID,
		ActorUserID:     actorID,
		Type:            notificationType,
		RelatedEntityID: relatedEntityID,
		IsRead:          false,
		CreatedAt:       now,
	}

	if err := u.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

type NotificationDTO struct {
	NotificationID       string  `json:"notification_id"`
	Type                 string  `json:"type"`
	ActorUserID          *string `json:"actor_user_id,omitempty"`
	ActorUsername        *string `json:"actor_username,omitempty"`
	ActorProfileImageURL *string `json:"actor_profile_image_url,omitempty"`
	RelatedEntityID      *string `json:"related_entity_id,omitempty"`
	IsRead               bool    `json:"is_read"`
	CreatedAt            string  `json:"created_at"`
}

func MapNotificationsToDTOs(notifications []domain.Notification) []NotificationDTO {
	dtos := make([]NotificationDTO, 0, len(notifications))

	for _, n := range notifications {
		dto := NotificationDTO{
			NotificationID: n.NotificationID,
			Type:           n.Type,
			IsRead:         n.IsRead,
			CreatedAt:      n.CreatedAt.Format(time.RFC3339),
		}

		if strings.TrimSpace(n.ActorUserID) != "" {
			actor := n.ActorUserID
			dto.ActorUserID = &actor
		}
		if strings.TrimSpace(n.ActorUsername) != "" {
			username := n.ActorUsername
			dto.ActorUsername = &username
		}
		if strings.TrimSpace(n.ActorProfileImageURL) != "" {
			profile := n.ActorProfileImageURL
			dto.ActorProfileImageURL = &profile
		}

		if strings.TrimSpace(n.RelatedEntityID) != "" {
			related := n.RelatedEntityID
			dto.RelatedEntityID = &related
		}

		dtos = append(dtos, dto)
	}

	return dtos
}
