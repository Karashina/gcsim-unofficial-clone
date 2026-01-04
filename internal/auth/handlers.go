package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/internal/db"
	"gorm.io/gorm"
)

// RegisterRequest ユーザー登録リクエスト
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
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

		// メールアドレスの重複チェック
		if _, err := userRepo.FindByEmail(req.Email); err == nil {
			log.Printf("[Register] Email already exists: %s", req.Email)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "email_exists",
				"message": "このメールアドレスは既に使用されています",
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
			Email:        req.Email,
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

	if req.Email == "" {
		return errors.New("メールアドレスは必須です")
	}
	if !isValidEmail(req.Email) {
		return errors.New("有効なメールアドレスを入力してください")
	}

	if req.Password == "" {
		return errors.New("パスワードは必須です")
	}
	if len(req.Password) < 8 {
		return errors.New("パスワードは8文字以上で入力してください")
	}

	return nil
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

// isValidEmail メールアドレスの簡易チェック
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
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
