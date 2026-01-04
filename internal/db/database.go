package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Database データベース接続を管理
type Database struct {
	DB             *gorm.DB
	UserRepository UserRepository
}

// Config データベース設定
type Config struct {
	// DBPath SQLiteデータベースファイルのパス（デフォルト: ./data/gcsim.db）
	DBPath string
	// Debug デバッグモード（SQLログを出力）
	Debug bool
}

var defaultConfig = Config{
	DBPath: "./data/gcsim.db",
	Debug:  false,
}

// NewDatabase データベース接続を初期化
func NewDatabase(config *Config) (*Database, error) {
	if config == nil {
		config = &defaultConfig
	}

	// データベースファイルのディレクトリを作成
	dir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// ログレベルの設定
	logLevel := logger.Silent
	if config.Debug {
		logLevel = logger.Info
	}

	// SQLite接続（Pure Go driver使用）
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        config.DBPath,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// マイグレーション実行
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("[DB] Database initialized at: %s", config.DBPath)

	// リポジトリを作成
	userRepo := NewUserRepository(db)

	return &Database{
		DB:             db,
		UserRepository: userRepo,
	}, nil
}

// Close データベース接続を閉じる
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// InitializeAdminUser 初期管理者ユーザーを作成
// 環境変数 GCSIM_ADMIN_USERNAME, GCSIM_ADMIN_PASSWORD, GCSIM_ADMIN_EMAIL から読み込む
func (d *Database) InitializeAdminUser(username, password, email string) error {
	// 既に管理者が存在するかチェック
	var count int64
	d.DB.Model(&User{}).Where("role = ?", UserRoleAdmin).Count(&count)
	if count > 0 {
		log.Printf("[DB] Admin user already exists, skipping initialization")
		return nil
	}

	// デフォルト値の設定
	if username == "" {
		username = "admin"
	}
	if email == "" {
		email = "admin@gcsim.local"
	}
	if password == "" {
		return fmt.Errorf("admin password is required")
	}

	// パスワードハッシュ化
	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// 管理者ユーザーを作成
	now := time.Now()
	admin := &User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         UserRoleAdmin,
		Status:       UserStatusApproved, // 管理者は自動承認
		ApprovedAt:   &now,
	}

	if err := d.UserRepository.Create(admin); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("[DB] Admin user created: %s", username)
	return nil
}

// GetStats データベース統計情報を取得
func (d *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 総ユーザー数
	var totalUsers int64
	d.DB.Model(&User{}).Count(&totalUsers)
	stats["total_users"] = totalUsers

	// 承認待ちユーザー数
	var pendingUsers int64
	d.DB.Model(&User{}).Where("status = ?", UserStatusPending).Count(&pendingUsers)
	stats["pending_users"] = pendingUsers

	// 承認済みユーザー数
	var approvedUsers int64
	d.DB.Model(&User{}).Where("status = ?", UserStatusApproved).Count(&approvedUsers)
	stats["approved_users"] = approvedUsers

	// 管理者数
	var adminUsers int64
	d.DB.Model(&User{}).Where("role = ?", UserRoleAdmin).Count(&adminUsers)
	stats["admin_users"] = adminUsers

	return stats, nil
}
