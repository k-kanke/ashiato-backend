package domain

import "time"

type Friendship struct {
	UserAID      string    `json:"user_a_id"`
	UserBID      string    `json:"user_b_id"`
	Status       string    `json:"status"`
	ActionUserID string    `json:"action_user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
