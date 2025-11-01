package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type postgresNotificationRepository struct {
	client *DBClient
}

func NewNotificationRepository(client *DBClient) repository.NotificationRepository {
	return &postgresNotificationRepository{client: client}
}

func (r *postgresNotificationRepository) Create(notification *domain.Notification) error {
	query := `
        INSERT INTO notifications (
            notification_id,
            recipient_user_id,
            actor_user_id,
            type,
            related_entity_id,
            is_read,
            created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	actorUserID := nullableString(notification.ActorUserID)
	relatedID := nullableString(notification.RelatedEntityID)

	_, err := r.client.DB.Exec(
		query,
		notification.NotificationID,
		notification.RecipientUserID,
		actorUserID,
		notification.Type,
		relatedID,
		notification.IsRead,
		notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}

	return nil
}

func (r *postgresNotificationRepository) ListByRecipient(userID string, limit int, before *time.Time) ([]domain.Notification, error) {
	query := `
        SELECT
            n.notification_id,
            n.recipient_user_id,
            n.actor_user_id,
            u.username,
            u.profile_image_url,
            n.type,
            n.related_entity_id,
            n.is_read,
            n.created_at
        FROM notifications n
        LEFT JOIN users u ON u.user_id = n.actor_user_id
        WHERE n.recipient_user_id = $1
    `

	args := []interface{}{userID}

	if before != nil {
		query += " AND n.created_at < $2"
		args = append(args, *before)
	}

	query += " ORDER BY n.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.client.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	results := make([]domain.Notification, 0)
	for rows.Next() {
		var (
			n             domain.Notification
			actorID       sql.NullString
			actorUsername sql.NullString
			actorProfile  sql.NullString
			relatedID     sql.NullString
		)
		if err := rows.Scan(
			&n.NotificationID,
			&n.RecipientUserID,
			&actorID,
			&actorUsername,
			&actorProfile,
			&n.Type,
			&relatedID,
			&n.IsRead,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if actorID.Valid {
			n.ActorUserID = actorID.String
		}
		if actorUsername.Valid {
			n.ActorUsername = actorUsername.String
		}
		if actorProfile.Valid {
			n.ActorProfileImageURL = actorProfile.String
		}
		if relatedID.Valid {
			n.RelatedEntityID = relatedID.String
		}

		results = append(results, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("notification rows iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresNotificationRepository) CountUnread(userID string) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM notifications
        WHERE recipient_user_id = $1 AND is_read = FALSE
    `

	var count int
	if err := r.client.DB.QueryRow(query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

func (r *postgresNotificationRepository) MarkAsRead(notificationID string) error {
	query := `
        UPDATE notifications
        SET is_read = TRUE
        WHERE notification_id = $1
    `

	result, err := r.client.DB.Exec(query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected when marking notification as read: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *postgresNotificationRepository) FindByID(notificationID string) (*domain.Notification, error) {
	query := `
        SELECT
            n.notification_id,
            n.recipient_user_id,
            n.actor_user_id,
            u.username,
            u.profile_image_url,
            n.type,
            n.related_entity_id,
            n.is_read,
            n.created_at
        FROM notifications n
        LEFT JOIN users u ON u.user_id = n.actor_user_id
        WHERE n.notification_id = $1
    `

	var (
		n             domain.Notification
		actorID       sql.NullString
		actorUsername sql.NullString
		actorProfile  sql.NullString
		relatedID     sql.NullString
	)

	err := r.client.DB.QueryRow(query, notificationID).Scan(
		&n.NotificationID,
		&n.RecipientUserID,
		&actorID,
		&actorUsername,
		&actorProfile,
		&n.Type,
		&relatedID,
		&n.IsRead,
		&n.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	if actorID.Valid {
		n.ActorUserID = actorID.String
	}
	if actorUsername.Valid {
		n.ActorUsername = actorUsername.String
	}
	if actorProfile.Valid {
		n.ActorProfileImageURL = actorProfile.String
	}
	if relatedID.Valid {
		n.RelatedEntityID = relatedID.String
	}

	return &n, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
