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
	// 矩形範囲を定義するPostGIS関数: ST_MakeEnvelope(min_lng, min_lat, max_lng, max_lat, SRID)
	// ST_MakeEnvelope(139.7, 35.6, 139.8, 35.7, 4326) のように使います

	// NOTE: SQLの可読性を重視し、ユーザーIDとフレンドシップをチェックする
	// 複雑なJOINとWHERE句を構築します。
	sql := `
        SELECT 
            p.pin_id, p.user_id, ST_Y(p.location::geometry) AS latitude, ST_X(p.location::geometry) AS longitude,
            p.content_text, p.media_url, p.privacy_setting, p.created_at
        FROM pins p
        -- フレンドシップテーブルをLEFT JOINし、フレンド関係が存在するかチェック
        LEFT JOIN friends f 
            ON f.status = 'accepted' 
            AND (
                (f.user_a_id = p.user_id AND f.user_b_id = $1) OR 
                (f.user_b_id = p.user_id AND f.user_a_id = $1)
            )
        WHERE 
            -- 1. ジオメトリ検索: ピンが指定された矩形内にあること
            ST_Within(p.location::geometry, 
                ST_SetSRID(ST_MakeEnvelope($2, $3, $4, $5), 4326)
            )
            -- 2. 権限チェック:
            AND (
                p.privacy_setting = 'public' 
                OR p.user_id = $1 -- 自分のピンは常に表示 
                OR (p.privacy_setting = 'friends' AND f.status = 'accepted') -- フレンド限定ピンでフレンド関係がacceptedである
            )
        ORDER BY p.created_at DESC
    `

	rows, err := r.client.DB.Query(sql, userID, minLng, minLat, maxLng, maxLat)
	if err != nil {
		return nil, fmt.Errorf("failed to query pins: %w", err)
	}
	defer rows.Close()

	pins := make([]domain.Pin, 0)
	for rows.Next() {
		var pin domain.Pin
		if err := rows.Scan(
			&pin.PinID,
			&pin.UserID,
			&pin.Latitude,
			&pin.Longitude,
			&pin.ContentText,
			&pin.MediaURL,
			&pin.PrivacySetting,
			&pin.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pin: %w", err)
		}
		pins = append(pins, pin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return pins, nil
}

func (r *postgresPinRepository) CreateComment(comment *domain.Comment) error {
	// コメント挿入処理は後で実装予定
	return nil
}
