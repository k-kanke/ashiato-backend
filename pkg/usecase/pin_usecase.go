// /pkg/usecase/pin_usecase.go (PinUsecase実装の一部)

package usecase

import (
	"errors"
	"fmt"
	"math"
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
	if err := u.validatePinLocation(userID, lat, lng); err != nil {
		return nil, err
	}

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

func (u *pinUsecase) validatePinLocation(userID string, lat, lng float64) error {
	if math.IsNaN(lat) || math.IsNaN(lng) {
		return errors.New("invalid coordinates: NaN detected")
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return errors.New("invalid coordinates: out of range")
	}

	latestPin, err := u.pinRepo.GetMostRecentPin(userID)
	if err != nil {
		return fmt.Errorf("location validation failed: %w", err)
	}

	// 初回投稿はそのまま許可
	if latestPin == nil {
		return nil
	}

	const permissibleDriftMeters = 5000.0 // 5km 以上離れていたら再認証を求める
	distance := haversineMeters(latestPin.Latitude, latestPin.Longitude, lat, lng)
	if distance > permissibleDriftMeters {
		return fmt.Errorf("location deviation (%.0fm) exceeds the permitted range. Please re-verify your location", distance)
	}

	return nil
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	toRadians := func(deg float64) float64 {
		return deg * math.Pi / 180.0
	}

	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)

	latRad1 := toRadians(lat1)
	latRad2 := toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
