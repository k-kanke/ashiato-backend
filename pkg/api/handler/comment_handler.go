package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

type CommentHandler struct {
	CommentUsecase usecase.CommentUsecase
}

func NewCommentHandler(uc usecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{CommentUsecase: uc}
}

type getThreadQuery struct {
	Limit int    `form:"limit,default=50"`
	After string `form:"after"`
}

type postCommentRequest struct {
	ContentText string `json:"content_text" binding:"required"`
}

func (h *CommentHandler) GetThread(c *gin.Context) {
	pinID := c.Param("pin_id")
	var query getThreadQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	var after *time.Time
	if query.After != "" {
		parsed, err := time.Parse(time.RFC3339, query.After)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor format"})
			return
		}
		after = &parsed
	}

	comments, err := h.CommentUsecase.GetThread(pinID, query.Limit, after)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"comments": comments})
	case errors.Is(err, usecase.ErrInvalidPinID),
		errors.Is(err, usecase.ErrInvalidCommentLimit):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch thread"})
	}
}

func (h *CommentHandler) PostComment(c *gin.Context) {
	pinID := c.Param("pin_id")
	userID := middleware.GetUserIDFromContext(c)

	var req postCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	comment, err := h.CommentUsecase.AddComment(pinID, userID, req.ContentText)
	switch {
	case err == nil:
		c.JSON(http.StatusCreated, gin.H{"comment": comment})
	case errors.Is(err, usecase.ErrInvalidPinID),
		errors.Is(err, usecase.ErrInvalidUserID),
		errors.Is(err, usecase.ErrEmptyCommentContent),
		errors.Is(err, usecase.ErrCommentTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
	}
}
