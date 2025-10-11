package repository

import "github.com/k-kanke/ashiato-backend/pkg/domain"

type UserRepository interface {
	// ユーザーアカウントを新規作成する
	CreateUser(
		user *domain.User,
		setting *domain.UserSettings,
	) error

	// メールアドレスを基にユーザーを検索する
	FindUserByEmail(email string) (*domain.User, error)

	// UserIDを基にユーザーと設定を検索する
	FindUserByID(userID string) (*domain.User, *domain.UserSettings, error)

	// その他のフレンドや設定更新メソッド
}
