package repository

import (
	"time"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
)

type CommentRepository interface {
	// ピンのコメントを取得する
	ListCommentsByPin(pinID string, limit int, after *time.Time) ([]domain.Comment, error)

	// ピンにコメントをする
	CreateComment(comment *domain.Comment) error
}
