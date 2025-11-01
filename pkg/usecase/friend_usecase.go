package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type FriendUsecase interface {
	// フレンド申請を送信する
	RequestFriendship(requesterID, targetID string) error

	// フレンド申請を承認する
	AcceptFriendship(accepterID, targetID string) error

	// フレンド一覧を取得する
	GetFriendsList(userID string) ([]FriendSummary, error)
}

type friendUsecase struct {
	friendRepo          repository.FriendRepository
	notificationUsecase NotificationUsecase
}

func NewFriendUsecase(fr repository.FriendRepository, notificationUC NotificationUsecase) FriendUsecase {
	return &friendUsecase{
		friendRepo:          fr,
		notificationUsecase: notificationUC,
	}
}

func (uc *friendUsecase) RequestFriendship(requesterID, targetID string) error {
	// 1. 自己申請のチェック
	if requesterID == targetID {
		return errors.New("cannot request friendship to self")
	}

	// 2. 既存の関係をチェック
	friendship, err := uc.friendRepo.FindFriendshipStatus(requesterID, targetID)
	if err != nil {
		return fmt.Errorf("failed to check existing friendship: %w", err)
	}
	if friendship != nil {
		if friendship.Status == "accepted" {
			return errors.New("already friends")
		}
		if friendship.Status == "pending" {
			return errors.New("request already pending")
		}
	}

	// 3. リポジトリで新規申請を作成 (status='pending')
	// userAID < userBID
	userA, userB := requesterID, targetID
	if requesterID > targetID {
		userA, userB = targetID, requesterID
	}

	if err := uc.friendRepo.CreateFriendship(userA, userB, requesterID); err != nil {
		return fmt.Errorf("failed to create friendship request: %w", err)
	}

	// 4. 通知ロジック: TargetID に通知を生成
	if uc.notificationUsecase != nil {
		if _, err := uc.notificationUsecase.CreateFriendRequestNotification(targetID, requesterID); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	return nil
}

func (uc *friendUsecase) AcceptFriendship(accepterID, targetID string) error {
	friendship, err := uc.friendRepo.FindFriendshipStatus(accepterID, targetID)
	if err != nil {
		return fmt.Errorf("failed to check friendship status: %w", err)
	}
	if friendship == nil {
		return errors.New("friendship request not found")
	}

	if friendship.Status != "pending" {
		return errors.New("no pending request exists")
	}

	// 承認するのは、申請の対象者（つまり、action_user_id ではない方）でなければならないというチェックも必要。

	// 3. リポジトリでステータスを 'accepted' に更新
	if err := uc.friendRepo.UpdateFriendshipStatus(
		friendship.UserAID,
		friendship.UserBID,
		"accepted",
		accepterID,
	); err != nil {
		return fmt.Errorf("failed to accept friendship: %w", err)
	}

	// 4. 通知ロジック: 申請者に承認通知を生成
	if uc.notificationUsecase != nil {
		recipientID := determineFriendRequestActor(friendship, accepterID)
		if recipientID != "" {
			if _, err := uc.notificationUsecase.CreateFriendAcceptedNotification(recipientID, accepterID); err != nil {
				return fmt.Errorf("failed to create acceptance notification: %w", err)
			}
		}
	}

	return nil
}

type FriendSummary struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
}

func (uc *friendUsecase) GetFriendsList(userID string) ([]FriendSummary, error) {
	friends, err := uc.friendRepo.GetFriendsList(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend list: %w", err)
	}

	summaries := make([]FriendSummary, 0, len(friends))
	for _, friend := range friends {
		summary := FriendSummary{
			UserID:   friend.UserID,
			Username: friend.Username,
		}
		if strings.TrimSpace(friend.ProfileImageURL) != "" {
			summary.ProfileImageURL = strings.TrimSpace(friend.ProfileImageURL)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func determineFriendRequestActor(friendship *domain.Friendship, accepterID string) string {
	if friendship == nil {
		return ""
	}

	if strings.TrimSpace(friendship.ActionUserID) != "" {
		return friendship.ActionUserID
	}

	// fallback: notify the other participant
	if accepterID == friendship.UserAID {
		return friendship.UserBID
	}
	return friendship.UserAID
}
