// /pkg/usecase/pin_usecase.go (PinUsecase実装の一部)

package usecase

import (
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
}

type pinUsecase struct {
	pinRepo repository.PinRepository
	// ... 他のリポジトリ
}

func NewPinUsecase(pr repository.PinRepository) PinUsecase {
	return &pinUsecase{pinRepo: pr}
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

// GetPinsForMap
