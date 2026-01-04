package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// Middleware JWT認証ミドルウェア
func Middleware(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("[Auth] Missing authorization header")
				http.Error(w, `{"error":"unauthorized","message":"認証が必要です"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Println("[Auth] Invalid authorization header format")
				http.Error(w, `{"error":"unauthorized","message":"不正な認証形式です"}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]
			claims, err := ValidateToken(token, config)
			if err != nil {
				log.Printf("[Auth] Token validation failed: %v\n", err)
				http.Error(w, `{"error":"unauthorized","message":"認証トークンが無効です"}`, http.StatusUnauthorized)
				return
			}

			// ユーザー情報をコンテキストに保存
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			ctx = context.WithValue(ctx, "claims", claims)

			log.Printf("[Auth] User authenticated: ID=%d, Username=%s, Role=%s", claims.UserID, claims.Username, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware 管理者権限チェックミドルウェア
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("role").(string)
		if !ok || role != "admin" {
			log.Printf("[Auth] Access denied: not an admin (role=%v)", role)
			http.Error(w, `{"error":"forbidden","message":"管理者権限が必要です"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalMiddleware トークンがあれば検証、なくても許可
func OptionalMiddleware(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// トークンなし → そのまま通過
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// 不正な形式 → そのまま通過（エラーにしない）
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]
			claims, err := ValidateToken(token, config)
			if err != nil {
				// トークンが無効 → そのまま通過（エラーにしない）
				log.Printf("[Auth] Optional: token validation failed: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			// トークンが有効 → コンテキストに保存
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			ctx = context.WithValue(ctx, "claims", claims)

			log.Printf("[Auth] Optional: User identified: ID=%d, Username=%s", claims.UserID, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
