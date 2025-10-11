package usecase

import (
	"errors"
	"fmt"

	"github.com/k-kanke/ashiato-backend/pkg/repository"
)

type FriendUsecase interface {
	// フレンド申請を送信する
	RequestFriendship(requesterID, targetID string) error

	// フレンド申請を承認する
	AcceptFriendship(accepterID, targetID string) error
}

type friendUsecase struct {
	friendRepo repository.FriendRepository
}

func NewFriendUsecase(fr repository.FriendRepository) FriendUsecase {
	return &friendUsecase{friendRepo: fr}
}

// RequestFriendship はフレンド申請ロジックを実行する
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
	// ... DBエラー処理

	// 3. リポジトリで新規申請を作成 (status='pending')
	// userAID < userBID の順序をGoのロジックで保証する必要がある
	userA, userB := requesterID, targetID
	if requesterID > targetID {
		userA, userB = targetID, requesterID
	}

	if err := uc.friendRepo.CreateFriendship(userA, userB, requesterID); err != nil {
		return fmt.Errorf("failed to create friendship request: %w", err)
	}

	// 4. 通知ロジック（後で実装）: TargetID に通知を生成
	// ...

	return nil
}

func (uc *friendUsecase) AcceptFriendship(accepterID, targetID string) error {
	return nil
}
