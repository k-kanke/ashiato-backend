// /pkg/usecase/pin_usecase.go (PinUsecase実装の一部)

package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type PinUsecase interface {
	// 新規ピンを作成
	PostNewPin(
		userID string,
		lat float64,
		lng float64,
		content string,
		mediaURL string,
		privacy string,
	) (*domain.Pin, error)

	// 地図表示用のピンを取得
	GetPinsForMap(
		userID string,
		minLat float64,
		maxLat float64,
		minLng float64,
		maxLng float64,
		privacy string,
	) ([]domain.Pin, error)
}

type pinUsecase struct {
	pinRepo repository.PinRepository
	// ... 他のリポジトリ
}

func NewPinUsecase(pinRepo repository.PinRepository) PinUsecase {
	return &pinUsecase{pinRepo: pinRepo}
}

// PostNewPin は新規Pin投稿の全ロジックを実行する
func (u *pinUsecase) PostNewPin(
	userID string,
	lat, lng float64,
	content string,
	mediaURL string,
	privacy string,
) (*domain.Pin, error) {
	// 1. 位置認証ロジック (例: 過去のアクセス履歴や、現在地の厳密な検証)
	// ここで位置情報の偽装チェックを行う
	// if !u.CheckLocationValidity(userID, lat, lng) { return nil, errors.New("location not verified") }

	newPin := &domain.Pin{
		PinID:          uuid.New().String(),
		UserID:         userID,
		Latitude:       lat,
		Longitude:      lng,
		ContentText:    content,
		MediaURL:       mediaURL,
		PrivacySetting: privacy,
		Status:         "active", // デフォルトはアクティブ
		CreatedAt:      time.Now(),
	}

	if err := u.pinRepo.CreatePin(newPin); err != nil {
		return nil, fmt.Errorf("pin creation failed: %w", err)
	}

	return newPin, nil
}

// GetPinsForMap は地図表示のためのPinを取得する
func (u *pinUsecase) GetPinsForMap(
	userID string,
	minLat, maxLat, minLng, maxLng float64,
	privacy string,
) ([]domain.Pin, error) {
	// 1. バリデーション: 矩形範囲が妥当かチェック
	if minLat >= maxLat || minLng >= maxLng {
		return nil, errors.New("invalid map bounding box coordinates")
	}

	// 2. リポジトリの呼び出し
	pins, err := u.pinRepo.GetPinsInArea(userID, minLat, maxLat, minLng, maxLng, privacy)
	if err != nil {
		return nil, fmt.Errorf("usecase failed to get pins: %w", err)
	}

	return pins, nil
}
