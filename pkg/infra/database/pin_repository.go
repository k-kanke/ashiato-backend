package database

import (
	"fmt"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type postgresPinRepository struct {
	client *DBClient
}

func NewPinRepository(client *DBClient) repository.PinRepository {
	return &postgresPinRepository{client: client}
}

func (r *postgresPinRepository) CreatePin(pin *domain.Pin) error {
	sql := `
        INSERT INTO pins (pin_id, user_id, location, content_text, media_url, privacy_setting, status, created_at) 
        VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, $8, $9)
    `
	// ST_MakePoint(経度, 緯度) で PostGIS の Point 型を作成
	_, err := r.client.DB.Exec(
		sql,
		pin.PinID,
		pin.UserID,
		pin.Longitude,
		pin.Latitude,
		pin.ContentText,
		pin.MediaURL,
		pin.PrivacySetting,
		pin.Status,
		pin.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert pin: %w", err)
	}
	return nil
}

func (r *postgresPinRepository) GetPinsInArea(
	userID string,
	minLat, maxLat, minLng, maxLng float64,
	privacySetting string,
) ([]domain.Pin, error) {
	// 今後の実装に備え、空の結果を返す
	return []domain.Pin{}, nil
}

func (r *postgresPinRepository) CreateComment(comment *domain.Comment) error {
	// コメント挿入処理は後で実装予定
	return nil
}
