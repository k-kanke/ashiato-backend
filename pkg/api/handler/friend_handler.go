package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k-kanke/ashiato-backend/pkg/api/middleware"
	"github.com/k-kanke/ashiato-backend/pkg/usecase"
)

type FriendHandler struct {
	FriendUsecase usecase.FriendUsecase
}

func NewFriendHandler(uc usecase.FriendUsecase) *FriendHandler {
	return &FriendHandler{FriendUsecase: uc}
}

func (h *FriendHandler) RequestFriendship(c *gin.Context) {
	requesterID := middleware.GetUserIDFromContext(c)
	targetID := c.Param("user_id")

	if err := h.FriendUsecase.RequestFriendship(requesterID, targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent successfully"})
}

func (h *FriendHandler) AcceptFriendship(c *gin.Context) {
	accepterID := middleware.GetUserIDFromContext(c)
	targetID := c.Param("user_id")

	if err := h.FriendUsecase.AcceptFriendship(accepterID, targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted successfully"})
}

func (h *FriendHandler) GetFriendsList(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	friendIDs, err := h.FriendUsecase.GetFriendsList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"friends": friendIDs})
}
