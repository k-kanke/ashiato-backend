package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type postgresFriendRepository struct {
	client *DBClient
}

func NewFriendRepository(client *DBClient) repository.FriendRepository {
	return &postgresFriendRepository{client: client}
}

func (r *postgresFriendRepository) FindFriendshipStatus(userA, userB string) (*domain.Friendship, error) {
	// ユーザーIDを正規化（userAID > userBID）
	id1, id2 := userA, userB
	if userA > userB {
		id1, id2 = userB, userA
	}

	var friendship domain.Friendship
	query := `SELECT user_a_id, user_b_id, status, action_user_id, created_at, updated_at 
            FROM friends WHERE user_a_id = $1 AND user_b_id = $2`

	err := r.client.DB.QueryRow(query, id1, id2).Scan(
		&friendship.UserAID,
		&friendship.UserBID,
		&friendship.Status,
		&friendship.ActionUserID,
		&friendship.CreatedAt,
		&friendship.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // 関係が存在しない場合はエラーではなく nil を返す（ユースケースで処理するため）
	}
	return &friendship, err
}

func (r *postgresFriendRepository) CreateFriendship(userAID, userBID, actionUserID string) error {
	now := time.Now()
	query := `INSERT INTO friends 
            (user_a_id, user_b_id, status, action_user_id, created_at, updated_at) 
            VALUES ($1, $2, 'pending', $3, $4, $5)`

	_, err := r.client.DB.Exec(query, userAID, userBID, actionUserID, now, now)
	return err
}

func (r *postgresFriendRepository) UpdateFriendshipStatus(userA, userB, newStatus, actionUserID string) error {
	// 1. ユーザーIDを正規化 (userAID < userBID)
	id1, id2 := userA, userB
	if userA > userB {
		id1, id2 = userB, userA
	}

	now := time.Now()

	query := `
        UPDATE friends
        SET 
            status = $3,
            action_user_id = $4,
            updated_at = $5
        WHERE user_a_id = $1 AND user_b_id = $2
    `

	result, err := r.client.DB.Exec(query, id1, id2, newStatus, actionUserID, now)
	if err != nil {
		return fmt.Errorf("failed to update friendship status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("friendship record not found or already updated")
	}

	return nil
}

func (r *postgresFriendRepository) GetFriendsList(userID string) ([]string, error) {
	query := `
        SELECT 
            CASE
                WHEN user_a_id = $1 THEN user_b_id
                ELSE user_a_id
            END AS friend_id
        FROM friends
        WHERE 
            -- user_a_id または user_b_id が自身のIDであり、
            (user_a_id = $1 OR user_b_id = $1)
            -- ステータスが 'accepted' であること
            AND status = 'accepted'
    `

	rows, err := r.client.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query friends list: %w", err)
	}
	defer rows.Close()

	var friendIDs []string
	for rows.Next() {
		var friendID string
		if err := rows.Scan(&friendID); err != nil {
			return nil, fmt.Errorf("failed to scan friend ID: %w", err)
		}
		friendIDs = append(friendIDs, friendID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return friendIDs, nil
}
