package domain

import "time"

type Comment struct {
	CommentID   string    `json:"comment_id"`
	PinID       string    `json:"pin_id"`
	UserID      string    `json:"user_id"`
	ContentText string    `json:"content_text"`
	CreatedAt   time.Time `json:"created_at"`
}
