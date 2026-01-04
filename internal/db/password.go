package db

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptのコスト（10が推奨デフォルト、高いほど安全だが遅い）
	bcryptCost = 10
)

// HashPassword パスワードをbcryptでハッシュ化
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword パスワードとハッシュを照合
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
