package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type postgresUserRepository struct {
	client *DBClient
}

// NewUserRepository は UserRepository の新しいインスタンスを返す
func NewUserRepository(client *DBClient) repository.UserRepository {
	return &postgresUserRepository{client: client}
}

func (r *postgresUserRepository) CreateUser(user *domain.User, settings *domain.UserSettings) error {
	// データベースへの挿入ロジック（トランザクション処理）

	// Userテーブルへの挿入
	sqlUser := `INSERT INTO users (user_id, username, email, password_hash, profile_image_url, created_at, updated_at) 
                VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.client.DB.Exec(sqlUser, user.UserID, user.Username, user.Email, user.PasswordHash, user.ProfileImageURL, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to  insert user: %w", err)
	}

	// UserSettingsテーブルへの挿入
	sqlSettings := `INSERT INTO user_settings (user_id, comment_on_my_pin, friend_new_pin, friend_request_received, friend_request_accepted) 
                    VALUES ($1, $2, $3, $4, $5)`
	_, err = r.client.DB.Exec(sqlSettings, settings.UserID, settings.CommentOnMyPin, settings.FriendNewPin, settings.FriendRequestReceived, settings.FriendRequestAccepted)
	if err != nil {
		// ユーザー挿入成功後に設定挿入失敗の場合、ロールバック
		return fmt.Errorf("failed to insert user settings: %w", err)
	}

	return nil
}

func (r *postgresUserRepository) FindUserByEmail(email string) (*domain.User, error) {
	user := &domain.User{}

	const query = `
		SELECT
			user_id,
			username,
			email,
			password_hash,
			profile_image_url,
			bio,
			is_banned,
			created_at,
			updated_at
		FROM users
		WHERE email = $1`

	err := r.client.DB.QueryRow(query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.ProfileImageURL,
		&user.Bio,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) FindUserByID(userID string) (*domain.User, *domain.UserSettings, error) {
	user := &domain.User{}

	const userQuery = `
		SELECT
			user_id,
			username,
			email,
			password_hash,
			profile_image_url,
			bio,
			is_banned,
			created_at,
			updated_at
		FROM users
		WHERE user_id = $1`

	err := r.client.DB.QueryRow(userQuery, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.ProfileImageURL,
		&user.Bio,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	settings := &domain.UserSettings{}

	const settingsQuery = `
		SELECT
			user_id,
			comment_on_my_pin,
			friend_new_pin,
			friend_request_received,
			friend_request_accepted
		FROM user_settings
		WHERE user_id = $1`

	err = r.client.DB.QueryRow(settingsQuery, userID).Scan(
		&settings.UserID,
		&settings.CommentOnMyPin,
		&settings.FriendNewPin,
		&settings.FriendRequestReceived,
		&settings.FriendRequestAccepted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to find user settings by id: %w", err)
	}

	return user, settings, nil
}
