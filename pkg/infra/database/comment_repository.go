package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type postgresCommentRepository struct {
	client *DBClient
}

func NewCommentRepository(client *DBClient) repository.CommentRepository {
	return &postgresCommentRepository{client: client}
}

func (r *postgresCommentRepository) ListCommentsByPin(pinID string, limit int, after *time.Time) ([]domain.Comment, error) {
	baseQuery := `
		SELECT
			comment_id,
			pin_id,
			user_id,
			content_text,
			created_at
		FROM comments
		WHERE pin_id = $1
	`

	var (
		rows *sql.Rows
		err  error
	)

	if after != nil {
		query := baseQuery + `
			AND created_at > $2
			ORDER BY created_at ASC
			LIMIT $3
		`
		rows, err = r.client.DB.Query(query, pinID, after.UTC(), limit)
	} else {
		query := baseQuery + `
			ORDER BY created_at ASC
			LIMIT $2
		`
		rows, err = r.client.DB.Query(query, pinID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var comment domain.Comment
		if err := rows.Scan(
			&comment.CommentID,
			&comment.PinID,
			&comment.UserID,
			&comment.ContentText,
			&comment.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate comment rows: %w", err)
	}

	return comments, nil
}

func (r *postgresCommentRepository) CreateComment(comment *domain.Comment) error {
	const query = `
		INSERT INTO comments (
			comment_id,
			pin_id,
			user_id,
			content_text,
			created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	if _, err := r.client.DB.Exec(
		query,
		comment.CommentID,
		comment.PinID,
		comment.UserID,
		comment.ContentText,
		comment.CreatedAt.UTC(),
	); err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	return nil
}
