package domain

import "time"

type Notification struct {
	NotificationID  string    `json:"notification_id"`
	RecipientUserID string    `json:"recipient_user_id"`
	ActorUserID     string    `json:"actor_user_id"`
	Type            string    `json:"type"`
	RelatedEntityID string    `json:"related_entity_id"`
	IsRead          bool      `json:"is_read"`
	CreatedAt       time.Time `json:"created_at"`
}
