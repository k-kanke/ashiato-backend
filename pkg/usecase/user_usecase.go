package usecase

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/k-kanke/ashiato-backend/pkg/domain"
	"github.com/k-kanke/ashiato-backend/pkg/repository"
	"github.com/k-kanke/ashiato-backend/pkg/shared"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	// 新規ユーザーを登録し、認証トークンを返す
	RegisterUser(username, email, password string) (token string, err error)

	// ユーザーを認証し、認証トークンを返す
	AuthenticateUser(email, password string) (token string, err error)

	// ユーザーのプロフィールを取得
	GetUserProfile(userID string) (*ProfileResponse, error)

	// その他のプロフィール更新、フレンド管理メソッド
}

type userUsecase struct {
	userRepo repository.UserRepository
}

type ProfileResponse struct {
	UserID                string `json:"user_id"`
	Username              string `json:"username"`
	Email                 string `json:"email"`
	CommentOnMyPin        bool   `json:"comment_on_my_pin"`
	FriendNewPin          bool   `json:"friend_new_pin"`
	FriendRequestReceived bool   `json:"friend_request_received"`
	FriendRequestAccepted bool   `json:"friend_request_accepted"`
	CreatedAt             string `json:"created_at"`
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) RegisterUser(username, email, password string) (token string, err error) {
	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// ユーザーIDとデフォルト設定の準備
	userID := uuid.New().String()
	now := time.Now()

	newUser := &domain.User{
		UserID:       userID,
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
		// ... その他のデフォルト値
	}

	defaultSettings := &domain.UserSettings{
		UserID:         userID,
		CommentOnMyPin: true,
		// ... その他のデフォルト設定
	}

	// リポジトリ経由でDBに保存
	if err := u.userRepo.CreateUser(newUser, defaultSettings); err != nil {
		// メール重複エラーの処理など
		return "", fmt.Errorf("registration failed: %w", err)
	}

	// 認証トークンの生成（一旦仮トークン）
	token = "generate-jwt-token"

	return token, nil
}

func (u *userUsecase) AuthenticateUser(email, password string) (token string, err error) {
	user, err := u.userRepo.FindUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", fmt.Errorf("authentication error: %w", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("missing JWT secret")
	}

	expiryStr := os.Getenv("TOKEN_EXPIRY_HOURS")
	expiryHours := 24 // default 1 day
	if expiryStr != "" {
		if parsed, parseErr := strconv.Atoi(expiryStr); parseErr == nil {
			expiryHours = parsed
		} else {
			return "", fmt.Errorf("invalid TOKEN_EXPIRY_HOURS: %w", parseErr)
		}
	}

	token, err = shared.GenerateToken(user.UserID, secret, expiryHours)
	if err != nil {
		return "", fmt.Errorf("token generation failed: %w", err)
	}

	return token, nil

}

// GetUserProfile はユーザー情報と設定をまとめて返す
func (u *userUsecase) GetUserProfile(userID string) (*ProfileResponse, error) {
	user, settings, err := u.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user profile: %w", err)
	}

	resp := &ProfileResponse{
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if settings != nil {
		resp.CommentOnMyPin = settings.CommentOnMyPin
		resp.FriendNewPin = settings.FriendNewPin
		resp.FriendRequestReceived = settings.FriendRequestReceived
		resp.FriendRequestAccepted = settings.FriendRequestAccepted
	}

	return resp, nil
}
