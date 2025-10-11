package usecase

type UserUsecase interface {
	// 新規ユーザーを登録し、認証トークンを返す
	RegisterUser(username, email, password string) (token string, err error)

	// ユーザーを認証し、認証トークンを返す
	AuthenticateUser(email, password string) (token string, err error)

	// その他のプロフィール更新、フレンド管理メソッド
}
