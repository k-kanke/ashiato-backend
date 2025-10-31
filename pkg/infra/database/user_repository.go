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
	var profileImageURL sql.NullString
	var bio sql.NullString

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
		&profileImageURL,
		&bio,
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

	if profileImageURL.Valid {
		user.ProfileImageURL = profileImageURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}

	return user, nil
}

func (r *postgresUserRepository) FindUserByID(userID string) (*domain.User, *domain.UserSettings, error) {
	user := &domain.User{}
	var profileImageURL sql.NullString
	var bio sql.NullString

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
		&profileImageURL,
		&bio,
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

	if profileImageURL.Valid {
		user.ProfileImageURL = profileImageURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
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

func (r *postgresUserRepository) UpdateUserSettings(settings *domain.UserSettings) error {
	const query = `
		UPDATE user_settings
		SET
			comment_on_my_pin = $2,
			friend_new_pin = $3,
			friend_request_received = $4,
			friend_request_accepted = $5
		WHERE user_id = $1
	`

	result, err := r.client.DB.Exec(
		query,
		settings.UserID,
		settings.CommentOnMyPin,
		settings.FriendNewPin,
		settings.FriendRequestReceived,
		settings.FriendRequestAccepted,
	)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("no user settings updated for user_id %s", settings.UserID)
	}

	return nil
}

func (r *postgresUserRepository) SearchUsers(keyword string, requesterID string, limit int) ([]repository.UserSearchResult, error) {
	const query = `
		SELECT
			u.user_id,
			u.username,
			u.profile_image_url,
			f.status,
			f.action_user_id
		FROM users u
		LEFT JOIN friends f ON
			f.user_a_id = LEAST(u.user_id, $2)
			AND f.user_b_id = GREATEST(u.user_id, $2)
		WHERE
			u.user_id <> $2
			AND u.is_banned = FALSE
			AND (
				u.username ILIKE '%' || $1 || '%'
				OR u.email ILIKE '%' || $1 || '%'
			)
		ORDER BY u.username ASC
		LIMIT $3
	`

	rows, err := r.client.DB.Query(query, keyword, requesterID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	results := make([]repository.UserSearchResult, 0)
	for rows.Next() {
		var (
			result          repository.UserSearchResult
			profileImageURL sql.NullString
			friendStatus    sql.NullString
			actionUserID    sql.NullString
		)

		if err := rows.Scan(
			&result.UserID,
			&result.Username,
			&profileImageURL,
			&friendStatus,
			&actionUserID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user search result: %w", err)
		}

		if profileImageURL.Valid {
			result.ProfileImageURL = profileImageURL.String
		}
		if friendStatus.Valid {
			result.FriendStatus = friendStatus.String
		}
		if actionUserID.Valid {
			result.ActionUserID = actionUserID.String
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user search rows: %w", err)
	}

	return results, nil
}
