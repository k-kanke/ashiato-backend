package domain

import "time"

type User struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	ProfileImageURL string    `json:"profile_image_url"`
	Bio             string    `json:"bio"`
	IsBanned        bool      `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UserSettings struct {
	UserID                string `json:"user_id"`
	CommentOnMyPin        bool   `json:"comment_on_my_pin"`
	FriendNewPin          bool   `json:"friend_new_pin"`
	FriendRequestReceived bool   `json:"friend_request_received"`
	FriendRequestAccepted bool   `json:"friend_request_accepted"`
}
