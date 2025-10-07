package domain

import "time"

type Pin struct {
	PinID          string    `json:"pin_id"`
	UserID         string    `json:"user_id"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	ContentText    string    `json:"content_text"`
	MediaURL       string    `json:"media_url"`
	PrivacySetting string    `json:"privacy_setting"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
