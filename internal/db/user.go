package db

import (
	"time"

	"gorm.io/gorm"
)

// UserStatus ユーザーのステータス
type UserStatus string

const (
	// UserStatusPending 承認待ち
	UserStatusPending UserStatus = "pending"
	// UserStatusApproved 承認済み
	UserStatusApproved UserStatus = "approved"
	// UserStatusRejected 却下
	UserStatusRejected UserStatus = "rejected"
	// UserStatusSuspended 停止中
	UserStatusSuspended UserStatus = "suspended"
)

// UserRole ユーザーのロール
type UserRole string

const (
	// UserRoleAdmin 管理者
	UserRoleAdmin UserRole = "admin"
	// UserRoleUser 一般ユーザー
	UserRoleUser UserRole = "user"
)

// User ユーザーモデル
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;not null;size:50" json:"username"`
	PasswordHash string     `gorm:"not null" json:"-"` // JSON出力時は除外
	Role         UserRole   `gorm:"not null;default:'user'" json:"role"`
	Status       UserStatus `gorm:"not null;default:'pending'" json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedBy   *uint      `json:"approved_by,omitempty"` // 承認者のユーザーID
}

// TableName テーブル名を明示的に指定
func (User) TableName() string {
	return "users"
}

// IsApproved ユーザーが承認済みかどうか
func (u *User) IsApproved() bool {
	return u.Status == UserStatusApproved
}

// IsAdmin ユーザーが管理者かどうか
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// CanLogin ログイン可能かどうか
func (u *User) CanLogin() bool {
	return u.Status == UserStatusApproved
}

// SafeUser パスワードハッシュを除外した安全なユーザー情報
type SafeUser struct {
	ID         uint       `json:"id"`
	Username   string     `json:"username"`
	Role       UserRole   `json:"role"`
	Status     UserStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
}

// ToSafeUser パスワードハッシュを除外した安全な形式に変換
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:         u.ID,
		Username:   u.Username,
		Role:       u.Role,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
		ApprovedAt: u.ApprovedAt,
	}
}

// UserRepository ユーザーリポジトリインターフェース
type UserRepository interface {
	// Create ユーザーを作成
	Create(user *User) error
	// FindByID IDでユーザーを検索
	FindByID(id uint) (*User, error)
	// FindByUsername ユーザー名でユーザーを検索
	FindByUsername(username string) (*User, error)
	// List ユーザー一覧を取得（ページネーション対応）
	List(offset, limit int) ([]User, int64, error)
	// ListPending 承認待ちユーザー一覧を取得
	ListPending() ([]User, error)
	// Update ユーザー情報を更新
	Update(user *User) error
	// UpdatePassword パスワードを更新
	UpdatePassword(id uint, passwordHash string) error
	// Delete ユーザーを削除
	Delete(id uint) error
}

// userRepository ユーザーリポジトリの実装
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository ユーザーリポジトリのインスタンスを作成
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create ユーザーを作成
func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByID IDでユーザーを検索
func (r *userRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername ユーザー名でユーザーを検索
func (r *userRepository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List ユーザー一覧を取得（ページネーション対応）
func (r *userRepository) List(offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64

	// 総数をカウント
	if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ページネーション付きで取得
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListPending 承認待ちユーザー一覧を取得
func (r *userRepository) ListPending() ([]User, error) {
	var users []User
	err := r.db.Where("status = ?", UserStatusPending).Order("created_at ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Update ユーザー情報を更新
func (r *userRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// UpdatePassword パスワードを更新
func (r *userRepository) UpdatePassword(id uint, passwordHash string) error {
	// Select を使用して password_hash カラムのみを更新対象に指定
	result := r.db.Model(&User{}).
		Where("id = ?", id).
		Select("password_hash").
		Updates(User{PasswordHash: passwordHash})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete ユーザーを削除
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}
