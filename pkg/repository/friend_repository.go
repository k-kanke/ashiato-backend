package repository

import "github.com/k-kanke/ashiato-backend/pkg/domain"

type FriendRepository interface {
	// フレンド申請を作成する (status='pending')
	CreateFriendship(userAID, userBID, actionUserID string) error

	// 既存の関係をステータスで検索する
	FindFriendshipStatus(userA, userB string) (*domain.Friendship, error)

	// フレンド申請を承認/拒否/ブロックなどで更新する
	UpdateFriendshipStatus(userA, userB, newStatus, actionUserID string) error

	// ユーザーIDを元にフレンド一覧を取得する
	GetFriendsList(userID string) ([]string, error)
}
