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
		return nil, nil
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

func (r *postgresFriendRepository) GetFriendsList(userID string) ([]repository.FriendSummary, error) {
	query := `
	        SELECT 
	            u.user_id,
	            u.username,
	            u.profile_image_url
	        FROM friends f
	        JOIN users u ON u.user_id = CASE
	            WHEN f.user_a_id = $1 THEN f.user_b_id
	            ELSE f.user_a_id
	        END
	        WHERE 
	            (f.user_a_id = $1 OR f.user_b_id = $1)
	            AND f.status = 'accepted'
	        ORDER BY u.username ASC
	    `

	rows, err := r.client.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query friends list: %w", err)
	}
	defer rows.Close()

	friends := make([]repository.FriendSummary, 0)
	for rows.Next() {
		var (
			friend   repository.FriendSummary
			imageURL sql.NullString
		)
		if err := rows.Scan(&friend.UserID, &friend.Username, &imageURL); err != nil {
			return nil, fmt.Errorf("failed to scan friend summary: %w", err)
		}
		if imageURL.Valid {
			friend.ProfileImageURL = imageURL.String
		}
		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return friends, nil
}
