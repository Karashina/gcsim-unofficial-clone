# WebUI JWT認証システムの実装

**日時**: 2026-01-04T19:00:00+09:00
**実施者**: GitHub Copilot (Claude Sonnet 4.5)
**ステータス**: 確認待ち

## 要約

gcsim WebUIに対して、JWT（JSON Web Token）ベースの安全な認証システムを実装しました。これにより、シミュレーションAPIへの不正アクセスを防ぎ、レート制限によるDoS攻撃対策も導入されました。

### 実装内容

- **認証方式**: JWT（HS256署名）
- **トークン有効期限**: 24時間
- **レート制限**: 開発環境10req/s、本番環境5req/s
- **保護対象**: `/api/simulate`, `/api/optimize`

---

## 変更ファイル

### バックエンド

#### 新規作成

- `internal/auth/jwt.go` - JWT生成・検証ロジック
- `internal/auth/middleware.go` - 認証ミドルウェア
- `internal/auth/ratelimit.go` - レート制限実装

#### 変更

- `go.mod` - 依存関係追加（golang-jwt/jwt/v5, golang.org/x/time）
- `cmd/devserver/main.go` - 認証とレート制限を統合
  - `/api/login`, `/api/verify` エンドポイント追加
  - 認証ミドルウェア適用
  - レート制限実装
  - CORS設定にAuthorizationヘッダー追加
- `cmd/gcsim-webui/main.go` - 認証とレート制限を統合
  - 同様の認証機能を追加

### フロントエンド

#### 新規作成

- `webui-src/src/auth.js` - 認証管理モジュール
  - トークン管理（LocalStorage）
  - ログイン/ログアウト処理
  - 認証付きfetchラッパー

#### 変更

- `webui/index.html` - ログイン画面UI追加
  - ログイン画面スタイル
  - ログインフォーム
- `webui-src/src/app.js` - 認証ロジック統合
  - auth.jsのインポート
  - fetch呼び出しをauthenticatedFetchに置換
  - ログイン初期化処理追加

### ドキュメント

- `.env.example` - 環境変数設定例
- `docs/AUTHENTICATION.md` - 認証システム詳細ドキュメント
  - セットアップ手順
  - API仕様
  - トラブルシューティング

---

## 実行コマンド

### 依存関係のインストール

```powershell
go mod tidy
```

### フロントエンドのビルド

```powershell
cd webui-src
npm install
npm run build
cd ..
```

### 環境変数の設定

```powershell
# .envファイルをコピー
Copy-Item .env.example .env

# 環境変数を設定（セッション単位）
$env:GCSIM_JWT_SECRET = "your-secret-key-here"
$env:GCSIM_ADMIN_PASSWORD = "your-password"
```

### サーバー起動

```powershell
# 開発サーバー
go run ./cmd/devserver/main.go

# 本番サーバー
go run ./cmd/gcsim-webui/main.go -addr :8382
```

---

## 検証結果

### コンパイルチェック

- [ ] `go mod tidy` - 実行必要（go.modを変更したため）
- [ ] `go build ./cmd/devserver` - 実行必要
- [ ] `go build ./cmd/gcsim-webui` - 実行必要

### 機能テスト（手動確認が必要）

- [ ] ログイン画面の表示確認
- [ ] パスワード認証の動作確認
- [ ] トークン保存・取得の確認
- [ ] 認証付きAPI呼び出しの確認
- [ ] トークン期限切れ時の自動ログアウト確認
- [ ] レート制限の動作確認

---

## 課題 / 残件

### 必須対応

1. **go mod tidyの実行**
   - `golang-jwt/jwt/v5`と`golang.org/x/time`の依存関係を解決
   
2. **フロントエンドのビルド**
   - `npm run build`を実行して`webui/app.js`を生成

3. **環境変数の設定**
   - 本番環境で適切な`GCSIM_JWT_SECRET`と`GCSIM_ADMIN_PASSWORD`を設定

### 今後の改善候補

- [ ] 複数ユーザーのサポート（現在は単一管理者のみ）
- [ ] トークンリフレッシュ機能
- [ ] ログイン履歴の記録
- [ ] OAuth2連携（Google/GitHub）
- [ ] 2要素認証（2FA）
- [ ] パスワードのハッシュ化（現在は平文比較）

---

## 備考

### セキュリティに関する注意

1. **JWTシークレットキー**: 本番環境では必ず64文字以上のランダム文字列を使用すること
2. **パスワード**: 推測されにくい強力なパスワードを設定すること
3. **HTTPS**: 本番環境では必ずHTTPSを使用すること
4. **環境変数**: `.env`ファイルを`.gitignore`に追加し、Gitにコミットしないこと

### 既知の制限事項

- 現在は単一ユーザー（管理者）のみをサポート
- パスワードは平文で比較（将来的にはbcryptなどでハッシュ化を推奨）
- トークンのブラックリスト機能なし（ログアウト時にサーバー側で無効化できない）

### デプロイ時の推奨事項

本番環境でのデプロイ前に以下を確認：

1. 強力なJWTシークレットキーの設定
2. 強力な管理者パスワードの設定
3. HTTPSの有効化
4. CORS設定の厳格化（必要に応じて特定オリジンのみ許可）
5. レート制限の適切な調整
6. 定期的なパスワード変更ポリシーの策定

---

## 次のアクション

1. ✅ 実装完了
2. ⏳ ユーザーによる確認とテスト
3. ⏳ `go mod tidy`の実行
4. ⏳ フロントエンドのビルド
5. ⏳ 動作確認
6. ⏳ セマンティックコミット

---

**確認をお願いします。**

問題がなければ、以下の手順で動作確認を行ってください：

```powershell
# 1. 依存関係の解決
go mod tidy

# 2. フロントエンドのビルド
cd webui-src
npm install
npm run build
cd ..

# 3. 環境変数の設定
$env:GCSIM_JWT_SECRET = "test-secret-key-for-development-only"
$env:GCSIM_ADMIN_PASSWORD = "admin"

# 4. 開発サーバーの起動
go run ./cmd/devserver/main.go

# 5. ブラウザでアクセス
# http://localhost:8381/ui/
```
