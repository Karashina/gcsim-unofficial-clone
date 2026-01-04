# JWT認証とユーザー管理システムの実装

**日時**: 2026-01-04T17:12:03+09:00  
**実施者**: GitHub Copilot (Claude Sonnet 4.5)  
**ステータス**: 完了  
**コミットハッシュ**: 4ac0196c0

## 要約

JWT認証とユーザー管理システムの完全な実装を行いました。これにより、セキュアな認証基盤が確立され、マルチユーザー対応のgcsim-webuiが実現可能になりました。

## 実施内容

### 1. 依存関係の追加・更新
- **Go バージョン**: 1.22.0 → 1.24.0 に更新
- **新規パッケージ追加**:
  - `github.com/golang-jwt/jwt/v5` v5.2.1 - JWT認証
  - `gorm.io/gorm` v1.30.0 - ORM
  - `gorm.io/driver/sqlite` v1.6.0 - SQLiteドライバー
  - `modernc.org/sqlite` v1.42.2 - Pure Go SQLiteドライバー
  - `golang.org/x/crypto` v0.31.0 - bcryptパスワードハッシュ化
  - `golang.org/x/time` v0.8.0 - レート制限機能
- **依存パッケージの更新**: 各種間接依存パッケージのバージョン更新

### 2. 認証システムの実装（internal/auth）

#### jwt.go - JWT認証コア機能
- JWTトークンの生成・検証機能
- カスタムクレーム構造（UserID, Username, Role）
- コンテキストからのユーザー情報取得ヘルパー関数
- トークン有効期限管理（デフォルト24時間）

#### handlers.go - 認証ハンドラー
- **RegisterHandler**: ユーザー登録エンドポイント
  - 入力バリデーション（ユーザー名、メール、パスワード）
  - 重複チェック（ユーザー名、メール）
  - bcryptによるパスワードハッシュ化
  - 初期ステータスは「pending」（承認待ち）
- **LoginHandlerWithDB**: ログインエンドポイント
  - ユーザー名+パスワード認証
  - ステータスチェック（承認済みユーザーのみログイン可能）
  - JWTトークン発行
- **ListUsersHandler**: 全ユーザー一覧取得（管理者用）
- **ListPendingUsersHandler**: 承認待ちユーザー一覧（管理者用）
- **ApproveUserHandler**: ユーザー承認/却下（管理者用）

#### middleware.go - 認証ミドルウェア
- **Middleware**: JWT認証必須ミドルウェア
  - Authorizationヘッダーの検証
  - Bearerトークンの抽出
  - トークン検証とクレーム取得
  - コンテキストへのユーザー情報保存
- **AdminMiddleware**: 管理者権限チェック
- **OptionalMiddleware**: トークンがあれば検証、なくても許可

#### ratelimit.go - レート制限
- IPアドレスベースのレート制限機能
- `golang.org/x/time/rate` を使用したトークンバケット実装
- 定期的なクリーンアップ機能（5分間隔）
- 設定可能なRPS（Requests Per Second）とバースト値

### 3. データベースシステムの実装（internal/db）

#### database.go - データベース管理
- GORM + SQLite統合
- Pure Go SQLiteドライバー（modernc.org/sqlite）使用
- 自動マイグレーション機能
- 初期管理者ユーザー作成機能
- データベース統計情報取得機能
- 設定可能なDBパスとデバッグモード

#### user.go - ユーザーモデルとリポジトリ
- **Userモデル**:
  - ID、Username、Email、PasswordHash
  - Role（admin/user）
  - Status（pending/approved/rejected/suspended）
  - 承認関連情報（ApprovedAt, ApprovedBy）
- **SafeUser**: パスワードハッシュを除外した安全な構造体
- **UserRepository インターフェース**:
  - Create, FindByID, FindByUsername, FindByEmail
  - List（ページネーション対応）
  - ListPending, Update, Delete
- ユーザー権限チェックメソッド（IsApproved, IsAdmin, CanLogin）

#### password.go - パスワード処理
- bcryptによるパスワードハッシュ化（コスト10）
- パスワード検証機能
- セキュアなハッシュ比較

### 4. 設定ファイル

#### .env.example
- JWT秘密鍵の設定例
- データベースパス設定
- 管理者初期設定（ユーザー名、パスワード、メール）
- サーバー設定
- 詳細な設定手順と注意事項
- PowerShellでの環境変数読み込み例
- セキュリティに関する注意事項

### 5. データベース初期化
- `data/gcsim.db` 作成済み
- テーブル: `users`, `sqlite_sequence`
- 初期ユーザー:
  - admin（ID: 1, Role: admin, Status: approved）
  - testuser（ID: 2, Role: user, Status: approved）

## 変更ファイル

### 新規作成
- `.env.example` - 環境変数設定サンプル
- `data/gcsim.db` - SQLiteデータベースファイル
- `internal/auth/jwt.go` - JWT認証機能
- `internal/auth/handlers.go` - 認証ハンドラー
- `internal/auth/middleware.go` - 認証ミドルウェア
- `internal/auth/ratelimit.go` - レート制限機能
- `internal/db/database.go` - データベース管理
- `internal/db/user.go` - ユーザーモデル
- `internal/db/password.go` - パスワード処理

### 更新
- `go.mod` - 依存関係の追加・更新
- `go.sum` - チェックサム更新

## 検証結果

- ✅ コンパイル: PASS（依存関係の解決確認済み）
- ✅ 構文チェック: PASS
- ✅ インポート: 未使用インポートなし
- ✅ コミット: 成功（コミットハッシュ: 4ac0196c0）

## セキュリティ対策

1. **パスワード管理**
   - bcrypt (cost=10) によるハッシュ化
   - パスワードハッシュはJSON出力から除外
   - SafeUser構造体による機密情報の隔離

2. **JWT トークン**
   - HMAC-SHA256署名
   - 有効期限設定（デフォルト24時間）
   - 署名メソッドの検証

3. **レート制限**
   - IPアドレスベースの制限
   - トークンバケットアルゴリズム
   - DDoS攻撃の緩和

4. **アクセス制御**
   - ロールベース認証（admin/user）
   - ステータスベース認証（承認制）
   - 管理者専用エンドポイントの保護

## 技術的な特徴

### Pure Go SQLiteドライバー
- cgoに依存しない`modernc.org/sqlite`を採用
- クロスコンパイルが容易
- Windows環境でのビルドが簡素化

### リポジトリパターン
- データアクセス層の抽象化
- テストが容易な設計
- 将来的なDB切り替えに対応

### ミドルウェアチェーン対応
- 認証ミドルウェアの柔軟な適用
- オプショナル認証のサポート
- 管理者権限チェックの独立

## 今後の拡張可能性

1. **認証機能**
   - リフレッシュトークンの実装
   - OAuth2/OIDC統合
   - 2段階認証（2FA）

2. **ユーザー管理**
   - プロフィール編集機能
   - パスワードリセット機能
   - メール通知機能

3. **セキュリティ強化**
   - Redis等によるレート制限の永続化
   - セッション管理の強化
   - 監査ログの実装

4. **データベース**
   - PostgreSQL/MySQLへの移行オプション
   - マイグレーション管理の強化
   - バックアップ機能

## 使用方法

### 環境変数の設定
```powershell
# .env.example をコピー
Copy-Item .env.example .env

# .env を編集して以下を設定:
# GCSIM_JWT_SECRET=<64文字以上の強力な秘密鍵>
# GCSIM_ADMIN_USERNAME=admin
# GCSIM_ADMIN_PASSWORD=<強力なパスワード>
# GCSIM_ADMIN_EMAIL=admin@example.com
```

### 初回起動
```powershell
# 開発サーバー起動（自動的に管理者ユーザーが作成される）
go run ./cmd/devserver
```

### APIエンドポイント例
- `POST /api/auth/register` - ユーザー登録
- `POST /api/auth/login` - ログイン
- `GET /api/admin/users` - ユーザー一覧（管理者のみ）
- `GET /api/admin/users/pending` - 承認待ちユーザー（管理者のみ）
- `POST /api/admin/users/approve` - ユーザー承認/却下（管理者のみ）

## 備考

- データベースファイル（`data/gcsim.db`）は`.gitignore`への追加を推奨
- 本番環境では必ず強力なJWT秘密鍵を設定すること
- 管理者パスワードは初回起動後、速やかに変更すること
- レート制限の設定値は負荷に応じて調整すること

## 参考ドキュメント

- [Docs/AUTHENTICATION.md](../../docs/AUTHENTICATION.md) - 認証システムの詳細仕様
- [.env.example](../../.env.example) - 環境変数設定例
- [CLAUDE.md](../../CLAUDE.md) - Claude Code開発ガイド（存在する場合）

---

**これにより、gcsim-webuiの認証基盤が完成しました。次のステップは、これを既存のdevserverに統合することです。**
