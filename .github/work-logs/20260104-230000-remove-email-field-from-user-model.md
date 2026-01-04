# emailフィールド完全削除によるユーザー登録問題の解決

**日時**: 2026-01-04T23:00:00+09:00  
**実施者**: GitHub Copilot (Claude Sonnet 4.5)  
**ステータス**: 完了

## 要約

ユーザー登録時にHTTP 500エラーが発生する問題を調査した結果、emailフィールドのUNIQUE制約が原因であることが判明。複数の修正方法を試みたが、GORMのAutoMigrate機能により制約が維持され続ける問題が発生。最終的に、emailフィールドをユーザーモデルから完全に削除することで問題を解決した。

## 問題の経緯

### 初期症状
- ユーザー登録画面で「登録処理に失敗しました」エラー
- 1人目のユーザーは登録可能、2人目以降は登録不可
- サーバーログに`UNIQUE constraint failed: users.email`エラー

### 根本原因
- `users`テーブルの`email`カラムに`UNIQUE`制約が設定されていた
- ユーザー登録時にemailフィールドを送信しないため、NULL/空文字列となる
- SQLiteのUNIQUE制約は複数のNULL値を許可しないため、2人目以降の登録で制約違反が発生

### 試行した解決策（失敗）

1. **Email を *string ポインタ型に変更**
   - 結果: 失敗
   - 理由: GORMがNULL値でもUNIQUE制約を維持

2. **Migration関数でインデックス削除**
   - 結果: 失敗
   - 理由: AutoMigrate()が毎回インデックスを再作成

3. **Manual table creation**
   - 結果: 失敗
   - 理由: AutoMigrate()が既存テーブルを上書き

4. **ensureEmailIndexIsNotUnique() 関数追加**
   - 結果: 失敗
   - 理由: 実行されていなかった、またはAutoMigrate()が後で上書き

### デプロイ時の追加問題
- 初回デプロイ後、サーバーバイナリのタイムスタンプが古いまま（11:09 UTC）
- データベーススキーマにemailカラムが残存
- deploy_remote.shがバイナリを正しく配置していなかった可能性
- 手動でSCP経由でバイナリをアップロードして解決

## 最終解決策：emailフィールドの完全削除

ユーザーの指示により、emailフィールドをすべてのコードから完全に削除することで根本的に解決。

### 変更ファイルと変更内容

#### 1. `internal/db/user.go` (5,718 bytes)
- **User構造体**: `Email *string` フィールドを削除
- **SafeUser構造体**: `Email *string` フィールドを削除
- **ToSafeUser()メソッド**: Emailフィールドのマッピングを削除
- **UserRepositoryインターフェース**: `FindByEmail(email string) (*User, error)` メソッドを削除
- **userRepository実装**: FindByEmail()の実装全体（約17行）を削除

最終的なUserスキーマ:
```go
type User struct {
    ID           uint      `gorm:"primarykey"`
    Username     string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    Role         UserRole  `gorm:"not null;default:'user'"`
    Status       UserStatus `gorm:"not null;default:'pending'"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ApprovedAt   *time.Time
    ApprovedBy   *uint
}
```

#### 2. `internal/auth/handlers.go` (22KB)
- **RegisterRequest構造体**: `Email string` フィールドを削除
- **RegisterHandler**: email重複チェックロジック（約15行）を削除
- **RegisterHandler**: user.Email代入処理を削除
- **validateRegisterRequest**: emailバリデーションロジックを削除

登録に必要な情報:
- username: 3-50文字、英数字・アンダースコア・ハイフン
- password: 8文字以上、大文字・小文字・数字・記号のうち2種類以上

#### 3. `internal/db/database.go` (6,474 bytes)
- **migrateUserTable()**: emailカラムとインデックスを削除するロジックに変更
  ```go
  // Drop email column if it exists
  db.Exec("ALTER TABLE users DROP COLUMN email")
  // Drop email-related indexes
  db.Exec("DROP INDEX IF EXISTS idx_users_email")
  db.Exec("DROP INDEX IF EXISTS idx_users_email_unique")
  ```
- **Manual table creation**: emailカラム定義を削除
- **InitializeAdminUser()**: シグネチャを `(username, password, email)` から `(username, password)` に変更
- **ensureEmailIndexIsNotUnique()**: 関数全体を削除

#### 4. `cmd/gcsim-webui/main.go`
- **InitializeAdminUser呼び出し**: 3パラメータから2パラメータに変更
  ```go
  // 変更前: database.InitializeAdminUser("karashina", "TestPass1234!", "")
  // 変更後: database.InitializeAdminUser("karashina", "TestPass1234!")
  ```

#### 5. `cmd/devserver/main.go`
- **adminEmail変数**: 削除
- **InitializeAdminUser呼び出し**: 3パラメータから2パラメータに変更

## 実行コマンド

### ビルド
```powershell
cd c:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\cmd\gcsim-webui
$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o gcsim-webui
```

### デプロイ
```powershell
# SCPでバイナリをアップロード
scp c:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\cmd\gcsim-webui\gcsim-webui uocuser@192.168.1.233:/tmp/gcsim-webui

# サーバーで実行
ssh -t uocuser@192.168.1.233 'sudo systemctl stop gcsim-webui; sudo rm /var/www/html/data/gcsim.db; sudo mv /tmp/gcsim-webui /usr/local/bin/gcsim-webui; sudo chmod +x /usr/local/bin/gcsim-webui; sudo systemctl start gcsim-webui'
```

## 検証結果

### ✅ コンパイルチェック
```
go build: 成功
バイナリサイズ: 39.3 MB
```

### ✅ データベーススキーマ
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME,
    updated_at DATETIME,
    approved_at DATETIME,
    approved_by INTEGER
)
```

確認コマンド:
```bash
sudo strings /var/www/html/data/gcsim.db | grep -i "email"
# 結果: マッチなし（exit code 1） ✅
```

### ✅ インデックス
```sql
CREATE UNIQUE INDEX idx_users_username ON users(username)
```

`idx_users_email`インデックスは完全に削除されました。

### ✅ サービス起動ログ
```
Jan 04 13:57:47 [DB] Creating users table manually
Jan 04 13:57:47 [DB] Users table created successfully
Jan 04 13:57:47 [DB] Database initialized at: ./data/gcsim.db
Jan 04 13:57:47 [DB] Admin user created: karashina
Jan 04 13:57:47 [Auth] JWT authentication enabled
```

### ✅ ユーザー登録テスト
- WebUI (http://192.168.1.233:8382/ui/index.html) から登録テスト実施
- 複数ユーザーの登録が正常に動作することを確認
- HTTP 500エラーは発生しなくなった

## 影響範囲

### 互換性の破壊
- **データベース**: 既存のemailカラムは削除される（データ保持不要と確認済み）
- **API**: RegisterRequestからemailフィールドが削除
- **既存ユーザー**: 影響なし（emailデータは使用されていなかった）

### 機能への影響
- メールアドレスによるユーザー検索機能は削除
- パスワードリセット機能（email送信）は使用不可
- 現在のシステムではemailフィールドは不要のため、影響なし

## 学んだ教訓

### GORMの挙動
1. **AutoMigrate()の制限**:
   - フィールド削除やインデックス削除は自動で行わない
   - 既存のUNIQUE制約を自動で削除できない
   - マイグレーション関数を実行しても、AutoMigrate()が後で上書きする可能性がある

2. **ポインタ型の挙動**:
   - `*string`型でもUNIQUE制約はNULL値に対して機能する
   - SQLiteはUNIQUE制約で複数のNULLを許可しない（PostgreSQLとは異なる）

### デプロイの教訓
1. **バイナリ更新の確認**:
   - デプロイスクリプトがバイナリを正しく配置しているか確認
   - `stat`コマンドでタイムスタンプを確認
   - デプロイ後は必ずバイナリバージョンを検証

2. **データベース状態の確認**:
   - `strings`コマンドでスキーマを確認
   - マイグレーションが正しく適用されているか検証
   - deploy_remote.shがdataディレクトリを保持するため、スキーマ変更時は手動削除が必要

### 問題解決アプローチ
1. **段階的修正の限界**:
   - 既存の制約を回避する方法は複雑で不安定
   - 根本的な原因（不要なフィールド）を削除する方が確実

2. **最小限の変更**:
   - 使用されていない機能は削除する
   - 「とりあえず残す」は将来の技術的負債となる

## 今後の推奨事項

### システム設計
- 必要ない機能は実装時に追加しない
- 将来の拡張性よりも現在の要件に集中
- YAGNI原則（You Aren't Gonna Need It）の徹底

### デプロイメント
- deploy_remote.shの改善を検討（バイナリ配置の確実性）
- デプロイ後の自動検証スクリプトの追加
- データベースマイグレーションの明示的な管理

### モニタリング
- ユーザー登録エラーのアラート設定
- データベース制約違反の監視
- API エラーレートの追跡

## 備考

### 削除されたコード量
- 合計約100行以上のコード削除
- 5ファイルに変更を適用
- テスト実行なし（テストコードが存在しないため）

### セキュリティへの影響
- メールアドレスによる本人確認機能は削除
- パスワードリセット機能は実装されていない
- 現状のシステムでは管理者承認フローで十分なセキュリティレベル

### パフォーマンスへの影響
- emailカラムとインデックスの削除により、わずかにディスク使用量が削減
- クエリパフォーマンスへの影響はほぼなし
- メモリ使用量もわずかに削減

---

**作業は完了し、複数ユーザーの登録が正常に動作することを確認しました。**
