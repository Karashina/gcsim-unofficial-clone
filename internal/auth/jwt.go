package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken JWT検証失敗
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrMissingClaims トークンにクレームが含まれていない
	ErrMissingClaims = errors.New("missing required claims")
)

// Claims JWT用のカスタムクレーム
type Claims struct {
	UserID   uint   `json:"user_id"`  // 数値IDに変更
	Username string `json:"username"` // ユーザー名を追加
	Role     string `json:"role"`     // ロール（admin/user）を追加
	jwt.RegisteredClaims
}

// Config 認証設定
type Config struct {
	// JWTシークレットキー（環境変数 GCSIM_JWT_SECRET から読み込み）
	JWTSecret []byte
	// トークンの有効期間（デフォルト: 24時間）
	TokenDuration time.Duration
	// 管理者パスワード（環境変数 GCSIM_ADMIN_PASSWORD から読み込み）
	AdminPassword string
}

// GenerateToken JWT トークンを生成
func GenerateToken(userID uint, username, role string, config Config) (string, error) {
	if len(config.JWTSecret) == 0 {
		return "", errors.New("JWT secret is not configured")
	}

	duration := config.TokenDuration
	if duration == 0 {
		duration = 24 * time.Hour // デフォルト24時間
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gcsim-webui",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

// ValidateToken JWT トークンを検証してクレームを返す
func ValidateToken(tokenString string, config Config) (*Claims, error) {
	if len(config.JWTSecret) == 0 {
		return nil, errors.New("JWT secret is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名メソッドの検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return config.JWTSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GetUserIDFromContext リクエストコンテキストからユーザーIDを取得
// ミドルウェアで設定されたユーザー情報を取得
func GetUserIDFromContext(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value("user_id").(uint)
	return userID, ok
}

// GetUsernameFromContext リクエストコンテキストからユーザー名を取得
func GetUsernameFromContext(r *http.Request) (string, bool) {
	username, ok := r.Context().Value("username").(string)
	return username, ok
}

// GetRoleFromContext リクエストコンテキストからロールを取得
func GetRoleFromContext(r *http.Request) (string, bool) {
	role, ok := r.Context().Value("role").(string)
	return role, ok
}

// IsAdminFromContext コンテキストから管理者かどうかを確認
func IsAdminFromContext(r *http.Request) bool {
	role, ok := GetRoleFromContext(r)
	return ok && role == "admin"
}

// ValidatePassword 管理者パスワードを検証
func ValidatePassword(password string, config Config) bool {
	if config.AdminPassword == "" {
		return false
	}
	return password == config.AdminPassword
}
