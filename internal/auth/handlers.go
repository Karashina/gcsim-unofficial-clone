package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/internal/db"
	"gorm.io/gorm"
)

// RegisterRequest ユーザー登録リクエスト
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest ログインリクエスト
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse ログインレスポンス
type LoginResponse struct {
	Success bool         `json:"success"`
	Token   string       `json:"token,omitempty"`
	User    *db.SafeUser `json:"user,omitempty"`
	Message string       `json:"message"`
}

// RegisterHandler ユーザー登録エンドポイント
func RegisterHandler(userRepo db.UserRepository, config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[Register] Invalid request body: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_request",
				"message": "リクエストが不正です",
			})
			return
		}

		// バリデーション
		if err := validateRegisterRequest(&req); err != nil {
			log.Printf("[Register] Validation error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "validation_error",
				"message": err.Error(),
			})
			return
		}

		// ユーザー名の重複チェック
		if _, err := userRepo.FindByUsername(req.Username); err == nil {
			log.Printf("[Register] Username already exists: %s", req.Username)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "username_exists",
				"message": "このユーザー名は既に使用されています",
			})
			return
		}

		// パスワードハッシュ化
		hash, err := db.HashPassword(req.Password)
		if err != nil {
			log.Printf("[Register] Password hashing failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "internal_error",
				"message": "登録処理に失敗しました",
			})
			return
		}

		// ユーザー作成
		user := &db.User{
			Username:     req.Username,
			PasswordHash: hash,
			Role:         db.UserRoleUser,
			Status:       db.UserStatusPending, // 承認待ち
		}

		if err := userRepo.Create(user); err != nil {
			log.Printf("[Register] User creation failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "internal_error",
				"message": "登録処理に失敗しました",
			})
			return
		}

		log.Printf("[Register] User registered: %s (ID: %d)", user.Username, user.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "登録が完了しました。管理者の承認をお待ちください。",
			"user":    user.ToSafeUser(),
		})
	}
}

// LoginHandlerWithDB ユーザー名+パスワードでログイン
func LoginHandlerWithDB(userRepo db.UserRepository, config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[Login] Invalid request body: %v", err)
			respondJSON(w, http.StatusBadRequest, LoginResponse{
				Success: false,
				Message: "リクエストが不正です",
			})
			return
		}

		// ユーザー検索
		user, err := userRepo.FindByUsername(req.Username)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[Login] User not found: %s", req.Username)
				respondJSON(w, http.StatusUnauthorized, LoginResponse{
					Success: false,
					Message: "ユーザー名またはパスワードが正しくありません",
				})
				return
			}
			log.Printf("[Login] Database error: %v", err)
			respondJSON(w, http.StatusInternalServerError, LoginResponse{
				Success: false,
				Message: "ログイン処理に失敗しました",
			})
			return
		}

		// パスワード検証
		if !db.VerifyPassword(req.Password, user.PasswordHash) {
			log.Printf("[Login] Invalid password for user: %s", req.Username)
			respondJSON(w, http.StatusUnauthorized, LoginResponse{
				Success: false,
				Message: "ユーザー名またはパスワードが正しくありません",
			})
			return
		}

		// ステータスチェック
		if !user.CanLogin() {
			log.Printf("[Login] User not approved: %s (status: %s)", req.Username, user.Status)
			var message string
			switch user.Status {
			case db.UserStatusPending:
				message = "アカウントは承認待ちです。管理者の承認をお待ちください。"
			case db.UserStatusRejected:
				message = "アカウントは拒否されました。"
			case db.UserStatusSuspended:
				message = "アカウントは停止されています。"
			default:
				message = "ログインできません。"
			}
			respondJSON(w, http.StatusForbidden, LoginResponse{
				Success: false,
				Message: message,
			})
			return
		}

		// JWTトークン生成
		token, err := GenerateToken(user.ID, user.Username, string(user.Role), config)
		if err != nil {
			log.Printf("[Login] Token generation failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, LoginResponse{
				Success: false,
				Message: "トークン生成に失敗しました",
			})
			return
		}

		log.Printf("[Login] Successful login: %s (ID: %d)", user.Username, user.ID)
		safeUser := user.ToSafeUser()
		respondJSON(w, http.StatusOK, LoginResponse{
			Success: true,
			Token:   token,
			User:    &safeUser,
			Message: "ログインに成功しました",
		})
	}
}

// validateRegisterRequest 登録リクエストのバリデーション
func validateRegisterRequest(req *RegisterRequest) error {
	if req.Username == "" {
		return errors.New("ユーザー名は必須です")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("ユーザー名は3文字以上50文字以内で入力してください")
	}
	if !isValidUsername(req.Username) {
		return errors.New("ユーザー名に使用できるのは英数字とアンダースコア、ハイフンのみです")
	}

	if req.Password == "" {
		return errors.New("パスワードは必須です")
	}
	if len(req.Password) < 8 {
		return errors.New("パスワードは8文字以上で入力してください")
	}
	if !isValidPassword(req.Password) {
		return errors.New("パスワードは大文字・小文字・数字・記号のうち2種類以上を含む必要があります")
	}

	return nil
}

// isValidPassword パスワードの複雑さをチェック
// 大文字・小文字・数字・記号のうち2種類以上を含む必要がある
func isValidPassword(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:',.<>?/~`", c):
			hasSpecial = true
		}
	}

	// 4種類のうち何種類含まれているかカウント
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	// 2種類以上必要
	return count >= 2
}

// isValidUsername ユーザー名の形式チェック
func isValidUsername(username string) bool {
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// isValidEmail メールアドレスの形式チェック（RFC 5322準拠）
func isValidEmail(email string) bool {
	// RFC 5322準拠の正規表現パターン（簡略版）
	// 完全なRFC 5322準拠は非常に複雑なため、実用的なサブセットを使用
	const emailPattern = `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	// 基本的な長さチェック
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	// @の数をチェック（正確に1つ）
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	// ローカル部とドメイン部に分割
	parts := strings.Split(email, "@")
	localPart := parts[0]
	domainPart := parts[1]

	// ローカル部の長さチェック（最大64文字）
	if len(localPart) == 0 || len(localPart) > 64 {
		return false
	}

	// ドメイン部の長さチェック（最大255文字）
	if len(domainPart) == 0 || len(domainPart) > 255 {
		return false
	}

	// ドメイン部に少なくとも1つのドットが必要
	if !strings.Contains(domainPart, ".") {
		return false
	}

	// 正規表現でパターンマッチ
	re := regexp.MustCompile(emailPattern)
	return re.MatchString(email)
}

// respondJSON JSON形式でレスポンスを返す
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// ListUsersHandler ユーザー一覧取得（管理者用）
func ListUsersHandler(userRepo db.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		users, _, err := userRepo.List(0, 100)
		if err != nil {
			log.Printf("[Admin] List users failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "ユーザー一覧の取得に失敗しました",
			})
			return
		}

		// SafeUserに変換
		safeUsers := make([]db.SafeUser, len(users))
		for i, user := range users {
			safeUsers[i] = user.ToSafeUser()
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": safeUsers,
			"total": len(safeUsers),
		})
	}
}

// ListPendingUsersHandler 承認待ちユーザー一覧取得（管理者用）
func ListPendingUsersHandler(userRepo db.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		users, err := userRepo.ListPending()
		if err != nil {
			log.Printf("[Admin] List pending users failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "承認待ちユーザー一覧の取得に失敗しました",
			})
			return
		}

		// SafeUserに変換
		safeUsers := make([]db.SafeUser, len(users))
		for i, user := range users {
			safeUsers[i] = user.ToSafeUser()
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": safeUsers,
			"total": len(safeUsers),
		})
	}
}

// ApproveUserRequest ユーザー承認リクエスト
type ApproveUserRequest struct {
	UserID uint   `json:"user_id"`
	Action string `json:"action"` // "approve" or "reject"
}

// ApproveUserHandler ユーザー承認/却下（管理者用）
func ApproveUserHandler(userRepo db.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req ApproveUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[Admin] Invalid request body: %v", err)
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_request",
				"message": "リクエストが不正です",
			})
			return
		}

		// ユーザー取得
		user, err := userRepo.FindByID(req.UserID)
		if err != nil {
			log.Printf("[Admin] User not found: %d", req.UserID)
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error":   "user_not_found",
				"message": "ユーザーが見つかりません",
			})
			return
		}

		// 承認者のID（コンテキストから取得）
		approverID, ok := r.Context().Value("user_id").(uint)
		if !ok {
			approverID = 0
		}

		// ステータス更新
		now := time.Now()
		switch req.Action {
		case "approve":
			user.Status = db.UserStatusApproved
			user.ApprovedAt = &now
			user.ApprovedBy = &approverID
			log.Printf("[Admin] User approved: %s (ID: %d) by admin ID: %d", user.Username, user.ID, approverID)
		case "reject":
			user.Status = db.UserStatusRejected
			log.Printf("[Admin] User rejected: %s (ID: %d) by admin ID: %d", user.Username, user.ID, approverID)
		default:
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_action",
				"message": "アクションは 'approve' または 'reject' を指定してください",
			})
			return
		}

		if err := userRepo.Update(user); err != nil {
			log.Printf("[Admin] User update failed: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "ユーザーの更新に失敗しました",
			})
			return
		}

		safeUser := user.ToSafeUser()
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "ユーザーのステータスを更新しました",
			"user":    safeUser,
		})
	}
}

// ChangePasswordRequest パスワード変更リクエスト
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordHandler パスワード変更（認証済みユーザー用）
func ChangePasswordHandler(userRepo db.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// コンテキストからユーザーIDを取得
		userID, ok := r.Context().Value("user_id").(uint)
		if !ok {
			log.Printf("[Auth] User ID not found in context")
			respondJSON(w, http.StatusUnauthorized, map[string]string{
				"error":   "unauthorized",
				"message": "認証が必要です",
			})
			return
		}

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[Auth] Invalid request body: %v", err)
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_request",
				"message": "リクエストが不正です",
			})
			return
		}

		// 入力検証
		if req.CurrentPassword == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "missing_current_password",
				"message": "現在のパスワードが必要です",
			})
			return
		}

		if req.NewPassword == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "missing_new_password",
				"message": "新しいパスワードが必要です",
			})
			return
		}

		// パスワードポリシーチェック
		if !isValidPassword(req.NewPassword) {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "weak_password",
				"message": "パスワードは8文字以上で、大文字・小文字・数字・記号のうち2種類以上を含む必要があります",
			})
			return
		}

		// ユーザーを取得
		user, err := userRepo.FindByID(userID)
		if err != nil {
			log.Printf("[Auth] User not found: %d", userID)
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error":   "user_not_found",
				"message": "ユーザーが見つかりません",
			})
			return
		}

		// 現在のパスワードを検証
		if !db.VerifyPassword(req.CurrentPassword, user.PasswordHash) {
			log.Printf("[Auth] Invalid current password for user: %s", user.Username)
			respondJSON(w, http.StatusUnauthorized, map[string]string{
				"error":   "invalid_password",
				"message": "現在のパスワードが正しくありません",
			})
			return
		}

		// 新しいパスワードをハッシュ化
		newHash, err := db.HashPassword(req.NewPassword)
		if err != nil {
			log.Printf("[Auth] Failed to hash new password: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "パスワードの変更に失敗しました",
			})
			return
		}

		// パスワードを更新
		if err := userRepo.UpdatePassword(userID, newHash); err != nil {
			log.Printf("[Auth] Failed to update password: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "パスワードの更新に失敗しました",
			})
			return
		}

		log.Printf("[Auth] Password changed successfully for user: %s (ID: %d)", user.Username, userID)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "パスワードを変更しました",
		})
	}
}

// AdminChangePasswordRequest 管理者によるパスワード変更リクエスト
type AdminChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// AdminChangePasswordHandler 管理者による他ユーザーのパスワード変更
func AdminChangePasswordHandler(userRepo db.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// URLからユーザーIDを取得（例: /api/admin/users/123/password）
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		log.Printf("[Admin] Path: %s, PathParts: %v, Length: %d", r.URL.Path, pathParts, len(pathParts))

		if len(pathParts) < 5 || pathParts[len(pathParts)-1] != "password" {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_path",
				"message": fmt.Sprintf("不正なパスです (expected: /api/admin/users/{id}/password, got: %s)", r.URL.Path),
			})
			return
		}

		targetUserID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 32)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_user_id",
				"message": fmt.Sprintf("ユーザーIDが不正です: %v", err),
			})
			return
		}

		var req AdminChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[Admin] Invalid request body: %v", err)
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid_request",
				"message": "リクエストが不正です",
			})
			return
		}

		// 新しいパスワードの検証
		if req.NewPassword == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "missing_password",
				"message": "新しいパスワードが必要です",
			})
			return
		}

		// パスワードポリシーチェック
		if !isValidPassword(req.NewPassword) {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "weak_password",
				"message": "パスワードは8文字以上で、大文字・小文字・数字・記号のうち2種類以上を含む必要があります",
			})
			return
		}

		// 対象ユーザーの存在確認
		targetUser, err := userRepo.FindByID(uint(targetUserID))
		if err != nil {
			log.Printf("[Admin] Target user not found: %d", targetUserID)
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error":   "user_not_found",
				"message": "ユーザーが見つかりません",
			})
			return
		}

		// 新しいパスワードをハッシュ化
		newHash, err := db.HashPassword(req.NewPassword)
		if err != nil {
			log.Printf("[Admin] Failed to hash new password: %v", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": "パスワードの変更に失敗しました",
			})
			return
		}

		// パスワードを更新
		log.Printf("[Admin] Attempting to update password for user ID: %d", targetUserID)
		if err := userRepo.UpdatePassword(uint(targetUserID), newHash); err != nil {
			log.Printf("[Admin] Failed to update password for user ID %d: %v", targetUserID, err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "internal_error",
				"message": fmt.Sprintf("パスワードの更新に失敗しました: %v", err),
			})
			return
		}

		adminID, _ := r.Context().Value("user_id").(uint)
		log.Printf("[Admin] Password changed by admin %d for user: %s (ID: %d)", adminID, targetUser.Username, targetUserID)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "パスワードを変更しました",
			"user":    targetUser.ToSafeUser(),
		})
	}
}
