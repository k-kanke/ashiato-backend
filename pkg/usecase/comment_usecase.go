package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

const (
	defaultThreadLimit = 50
	maxThreadLimit     = 200
	maxCommentLength   = 500
)

var (
	ErrInvalidPinID        = errors.New("pin id must be provided")
	ErrInvalidUserID       = errors.New("user id must be provided")
	ErrInvalidCommentLimit = errors.New("comment limit must be positive")
	ErrEmptyCommentContent = errors.New("comment content cannot be empty")
	ErrCommentTooLong      = fmt.Errorf("comment content exceeds %d characters", maxCommentLength)
)

type CommentUsecase interface {
	GetThread(pinID string, limit int, after *time.Time) ([]domain.Comment, error)
	AddComment(pinID string, userID string, content string, mediaURL *string) (*domain.Comment, error)
}

type commentUsecase struct {
	commentRepo repository.CommentRepository
}

func NewCommentUsecase(commentRepo repository.CommentRepository) CommentUsecase {
	return &commentUsecase{commentRepo: commentRepo}
}

func (u *commentUsecase) GetThread(pinID string, limit int, after *time.Time) ([]domain.Comment, error) {
	if strings.TrimSpace(pinID) == "" {
		return nil, ErrInvalidPinID
	}

	resolvedLimit := limit
	switch {
	case resolvedLimit <= 0:
		if limit == 0 {
			resolvedLimit = defaultThreadLimit
		} else {
			return nil, ErrInvalidCommentLimit
		}
	case resolvedLimit > maxThreadLimit:
		resolvedLimit = maxThreadLimit
	}

	comments, err := u.commentRepo.ListCommentsByPin(pinID, resolvedLimit, after)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comments: %w", err)
	}

	return comments, nil
}

func (u *commentUsecase) AddComment(pinID string, userID string, content string, mediaURL *string) (*domain.Comment, error) {
	if strings.TrimSpace(pinID) == "" {
		return nil, ErrInvalidPinID
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, ErrEmptyCommentContent
	}
	if utf8.RuneCountInString(trimmed) > maxCommentLength {
		return nil, ErrCommentTooLong
	}

	comment := &domain.Comment{
		CommentID:   uuid.New().String(),
		PinID:       pinID,
		UserID:      userID,
		ContentText: trimmed,
		CreatedAt:   time.Now().UTC(),
	}

	if mediaURL != nil {
		comment.MediaURL = *mediaURL
	}

	if err := u.commentRepo.CreateComment(comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}
