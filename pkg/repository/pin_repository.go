package repository

import "github.com/k-kanke/ashiato-backend/pkg/domain"

type PinRepository interface {
	// ピンを作成
	CreatePin(pin *domain.Pin) error

	// 特定の矩形範囲内のPin情報を取得する
	GetPinsInArea(
		userID string,
		minLat, maxLat, minLng, maxLng float64,
		privacySetting string,
	) ([]domain.Pin, error)

	// ユーザーの最新のピンを取得する
	GetMostRecentPin(userID string) (*domain.Pin, error)

	// Pinにコメントを追加する
	CreateComment(comment *domain.Comment) error
}
