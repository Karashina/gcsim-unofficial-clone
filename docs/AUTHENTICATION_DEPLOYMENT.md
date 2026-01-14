# JWT 認証機能付きシステムのデプロイ手順

## 概要

このドキュメントは、JWT 認証機能を含む gcsim-webui を Debian 12 サーバへデプロイする手順を説明します。

## 前提条件

### ローカル環境（Windows）
- OpenSSH クライアント（ssh, scp）がインストールされていること
- Go 1.x がインストールされていること（Linux バイナリのクロスコンパイル用）
- デプロイ用の SSH 鍵（例: `C:\Users\linol\HL-Creds\id_ed25519`）

### サーバー環境（Debian 12）
- SSH アクセスが可能なこと
- nginx がインストールされていること（または自動インストール）
- Cloudflare Tunnel（cloudflared）がセットアップされていること
- sudo 権限があること

## デプロイの全体フロー

```
1. Linux バイナリのビルド（ローカル）
2. server.env の準備（JWT シークレット等の設定）
3. 認証コード、WebUI、環境設定のアップロード（deploy_webui.ps1実行）
4. サービスの自動起動と動作確認
5. 初期管理者パスワードの変更
```

## 1. Linux バイナリのビルド

Windows 環境から Linux 用のバイナリをクロスコンパイルします：

```powershell
# プロジェクトルートで実行
cd webui/server/cmd/gcsim-webui
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ../../gcsim-webui .
cd ../../../..

# ビルド成功を確認
if (Test-Path gcsim-webui) {
    Write-Host "Linux binary built successfully" -ForegroundColor Green
} else {
    Write-Error "Build failed"
}
```

## 2. デプロイスクリプトの実行

`webui/server/scripts/deploy_webui.ps1` を使用してファイルをアップロードします：

```powershell
# 基本的なデプロイ（認証コード込み）
.\webui\server\scripts\deploy_webui.ps1 `
    -Server "192.168.1.233" `
    -User "uocuser" `
    -KeyFile "C:\Users\linol\HL-Creds\id_ed25519" `
    -RemotePath "/var/www/html" `
    -ReloadNginx
```

### デプロイスクリプトが行うこと

- `webui/` ディレクトリを `/var/www/html` へアップロード
- `internal/auth/` と `internal/db/` を `/opt/gcsim/internal/` へアップロード
- `gcsim-webui` バイナリを `/usr/local/bin/` へ配置
- `server.env` を `/etc/gcsim-webui/.env` へ配置（環境変数設定）
- systemd ユニットファイルを `/etc/systemd/system/gcsim-webui.service` に作成
- `/etc/gcsim-webui/` ディレクトリを作成（パーミッション設定済み）
- `/var/www/html/data/` ディレクトリを作成（データベース用）
- サービスを自動的に再起動

**重要**: デプロイスクリプトは完全自動化されており、手動でのサーバー設定は不要です。

## 3. デプロイ後の確認

デプロイスクリプト実行後、サービスが正常に起動したか確認します：

```bash
# SSH でサーバーに接続
ssh -i ~/.ssh/id_ed25519 uocuser@192.168.1.233

# サービスステータス確認
sudo systemctl status gcsim-webui
```

### ログの確認

```bash
# リアルタイムログ表示
sudo journalctl -u gcsim-webui -f

# 起動時のログを確認（JWT設定など）
sudo journalctl -u gcsim-webui -n 50
```

期待されるログ出力：
```
[DB] Database initialized at: /var/www/html/data/gcsim.db
[DB] Admin user created: karashina
[Auth] JWT authentication enabled
[Auth] JWT secret length: 64 characters
[CORS] Allowed origins: https://gcsim-uoc.linole.net
[Proxy] Trusted proxies: [127.0.0.1]
gcsim-webui: listening on :8382
```

**注意**: `server.env`ファイルはプロジェクトルートに配置されており、デプロイスクリプトが自動的にアップロードします。環境変数を変更する必要がある場合は、ローカルの`server.env`を編集してから再デプロイしてください。

## 4. 初回パスワード変更（重要）

デプロイ直後、セキュリティのため初期管理者パスワードを変更してください：

```powershell
# PowerShell で実行
# 1. ログイン
$token = (Invoke-RestMethod -Uri "https://gcsim-uoc.linole.net/api/login" `
    -Method Post `
    -Body (@{ username="karashina"; password="TestPass1234!" } | ConvertTo-Json) `
    -ContentType "application/json").token

# 2. パスワード変更
Invoke-RestMethod -Uri "https://gcsim-uoc.linole.net/api/change-password" `
    -Method Post `
    -Headers @{ Authorization = "Bearer $token" } `
    -Body (@{ 
        current_password="TestPass1234!"
        new_password="YourSecurePassword123!" 
    } | ConvertTo-Json) `
    -ContentType "application/json"

Write-Host "✅ パスワード変更完了"
```

詳細は [パスワード変更ガイド](PASSWORD_CHANGE_GUIDE.md) を参照してください。

## 5. nginx の設定確認

Cloudflare Tunnel 経由でアクセスする場合、nginx の設定を確認します：

```bash
# nginx 設定ファイルを確認
sudo nano /etc/nginx/sites-available/gcsim-deploy-*
```

API エンドポイントへのプロキシ設定が必要な場合：

```nginx
# /api/* のリクエストをバックエンドへプロキシ
location /api/ {
    proxy_pass http://localhost:8382;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

設定変更後：

```bash
# 設定のテスト
sudo nginx -t

# nginx をリロード
sudo systemctl reload nginx
```

## 6. 動作確認

### ヘルスチェック

```bash
# サーバー内部から確認
curl http://localhost:8382/healthz

# 期待される応答
{"status":"ok"}
```

### 初期管理者アカウント

システムの初回起動時、以下の管理者アカウントが自動的に作成されます：

- **ユーザー名**: `karashina`
- **パスワード**: `TestPass1234!`
- **メールアドレス**: `karashina@gcsim.local`
- **ロール**: `admin`

**セキュリティ重要**: このパスワードはコード内にハードコードされているため、初回ログイン後は必ず変更してください。詳細は [パスワード変更ガイド](PASSWORD_CHANGE_GUIDE.md) を参照してください。

### ブラウザからアクセス

1. `https://gcsim-uoc.linole.net/` にアクセス
2. ログインページが表示されることを確認
3. 初期管理者アカウント（karashina / TestPass1234!）でログイン

### API 動作確認

```bash
# ログインテスト（初期管理者アカウント）
curl -X POST https://gcsim-uoc.linole.net/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "karashina",
    "password": "TestPass1234!"
  }'

# 成功時のレスポンス
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "karashina",
    "role": "admin",
    "approved": true
  }
}
```

**セキュリティ重要**: 初回ログイン後は必ずパスワード変更を行ってください：

```bash
# パスワード変更API（認証済みユーザー）
curl -X POST https://gcsim-uoc.linole.net/api/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "current_password": "TestPass1234!",
    "new_password": "YourNewSecurePassword123!"
  }'
```

## トラブルシューティング

### サービスが起動しない

```bash
# エラーログを確認
sudo journalctl -u gcsim-webui -n 100 --no-pager

# よくあるエラー
# 1. "GCSIM_JWT_SECRET environment variable is not set"
#    → .env ファイルが正しい場所にあるか確認

# 2. "failed to open database"
#    → データディレクトリのパーミッション確認

# 3. "bind: address already in use"
#    → ポート 8382 が既に使用されていないか確認
```

### 環境変数が読み込まれない

```bash
# systemd unit ファイルを確認
sudo cat /etc/systemd/system/gcsim-webui.service

# EnvironmentFile 行があることを確認
# EnvironmentFile=-/etc/gcsim-webui/.env

# ファイルが存在し、読み取り可能か確認
sudo -u www-data cat /etc/gcsim-webui/.env
```

### データベース関連のエラー

```bash
# データベースファイルの確認
ls -la /var/www/html/data/

# パーミッションの修正
sudo chown -R www-data:www-data /var/www/html/data
sudo chmod 750 /var/www/html/data
sudo chmod 640 /var/www/html/data/*.db
```

### CORS エラー

ブラウザのコンソールで CORS エラーが出る場合：

1. `.env` の `GCSIM_CORS_ALLOWED_ORIGINS` を確認
2. サービスを再起動: `sudo systemctl restart gcsim-webui`
3. ログで CORS 設定を確認: `sudo journalctl -u gcsim-webui -n 20 | grep CORS`

### プロキシ経由の IP アドレスが正しく取得できない

1. `.env` の `GCSIM_TRUSTED_PROXIES` を確認
2. nginx の設定で `X-Real-IP` または `X-Forwarded-For` ヘッダーが送信されているか確認
3. Cloudflare Tunnel の場合は `127.0.0.1` を trusted proxy に追加

## セキュリティ上の注意事項

### 本番環境での必須設定

1. **JWT シークレット**
   - 最低 32 文字、推奨 48 文字以上
   - `openssl rand -base64 48` で生成
   - 絶対に Git にコミットしない

2. **CORS 設定**
   - `*` は使用しない
   - 実際のドメインのみを許可
   - 複数ドメインの場合はカンマ区切り

3. **ファイルパーミッション**
   - `.env` ファイル: `600` (所有者のみ読み書き可能)
   - データベースファイル: `640` (所有者書き込み可、グループ読み込み可)
   - データディレクトリ: `750`

4. **パスワードポリシー**
   - 最低 8 文字
   - 2 種類以上の文字種（大文字・小文字・数字・記号）

5. **レート制限**
   - デフォルトで有効（1 IP あたり秒間 10 リクエスト）
   - 必要に応じてコードで調整

## バックアップ手順

### データベースのバックアップ

```bash
# データベースをバックアップ
sudo systemctl stop gcsim-webui
sudo cp /var/www/html/data/gcsim.db /var/www/html/data/gcsim.db.backup-$(date +%Y%m%d)
sudo systemctl start gcsim-webui

# 古いバックアップの削除（30日以上前）
find /var/www/html/data/ -name "gcsim.db.backup-*" -mtime +30 -delete
```

### 環境変数のバックアップ

```bash
# .env ファイルをバックアップ
sudo cp /etc/gcsim-webui/.env /etc/gcsim-webui/.env.backup
```

## アップデート手順

新しいバージョンをデプロイする場合：

```bash
# 1. Windows 側で新しいバイナリをビルド
cd cmd/gcsim-webui
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ../../gcsim-webui .
cd ../..

# 2. デプロイスクリプトを実行
.\scripts\deploy_webui.ps1 -ReloadNginx

# 3. サーバー側でサービスを再起動
ssh uocuser@192.168.1.233 "sudo systemctl restart gcsim-webui"

# 4. 動作確認
curl https://gcsim-uoc.linole.net/healthz
```

## 関連ファイル

- [scripts/deploy_webui.ps1](../scripts/deploy_webui.ps1) - デプロイスクリプト
- [scripts/DEPLOY_WEBUI.md](../scripts/DEPLOY_WEBUI.md) - 一般的なデプロイ手順
- [.env.production.example](../.env.production.example) - 本番環境用 .env テンプレート
- [.env.example](../.env.example) - 開発環境用 .env テンプレート
- [docs/AUTHENTICATION.md](./AUTHENTICATION.md) - 認証システムの詳細仕様

## サポート

問題が発生した場合は、以下を確認してください：

1. システムログ: `sudo journalctl -u gcsim-webui -f`
2. nginx ログ: `sudo tail -f /var/log/nginx/error.log`
3. サービスステータス: `sudo systemctl status gcsim-webui`
4. 環境変数: `.env` ファイルの内容と systemd unit の EnvironmentFile 設定
