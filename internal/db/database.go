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

	// コネクションプールの設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SQLiteは同時書き込みが1つに制限されるため、適切な値を設定
	sqlDB.SetMaxOpenConns(10)                  // 最大オープン接続数
	sqlDB.SetMaxIdleConns(5)                   // 最大アイドル接続数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 接続の最大生存時間
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // アイドル接続の最大生存時間

	// テーブルが存在しない場合は手動で作成
	if !db.Migrator().HasTable(&User{}) {
		log.Printf("[DB] Creating users table manually")
		createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME,
			updated_at DATETIME,
			approved_at DATETIME,
			approved_by INTEGER
		)`

		if err := db.Exec(createTableSQL).Error; err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}

		// Usernameに UNIQUE INDEX を作成
		if err := db.Exec("CREATE UNIQUE INDEX idx_users_username ON users(username)").Error; err != nil {
			return nil, fmt.Errorf("failed to create username index: %w", err)
		}

		log.Printf("[DB] Users table created successfully")
	} else {
		// 既存テーブルのマイグレーション（emailカラム削除）
		if err := migrateUserTable(db); err != nil {
			log.Printf("[DB] Warning: User table migration failed: %v", err)
		}

		// マイグレーション実行
		if err := db.AutoMigrate(&User{}); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
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

// migrateUserTable usersテーブルのスキーマ更新（email削除対応）
func migrateUserTable(db *gorm.DB) error {
	// テーブルが存在しない場合は何もしない
	if !db.Migrator().HasTable(&User{}) {
		return nil
	}

	// emailカラムが存在する場合は削除
	if db.Migrator().HasColumn(&User{}, "email") {
		log.Printf("[DB] Dropping email column from users table")
		if err := db.Migrator().DropColumn(&User{}, "email"); err != nil {
			log.Printf("[DB] Warning: Failed to drop email column: %v", err)
		} else {
			log.Printf("[DB] Email column dropped successfully")
		}
	}

	// email関連のインデックスを削除
	db.Exec("DROP INDEX IF EXISTS idx_users_email")
	db.Exec("DROP INDEX IF EXISTS uidx_users_email")

	return nil
}

// InitializeAdminUser 初期管理者ユーザーを作成
func (d *Database) InitializeAdminUser(username, password string) error {
	// 既に管理者が存在するかチェック
	var count int64
	d.DB.Model(&User{}).Where("role = ?", UserRoleAdmin).Count(&count)
	if count > 0 {
		log.Printf("[DB] Admin user already exists, skipping initialization")
		return nil
	}

	// パラメータ検証
	if username == "" || password == "" {
		return fmt.Errorf("admin username and password are required")
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
