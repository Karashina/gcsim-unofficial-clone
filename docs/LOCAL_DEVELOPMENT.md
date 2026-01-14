# ローカル開発環境セットアップガイド

gcsimのローカル開発環境の構築と実行方法を説明します。

## 前提条件

- **Go 1.22.0以上**: メインアプリケーション用
- **Node.js & npm**: フロントエンド（webui）のビルド用
- **PowerShell 5.1以上**: Windowsスクリプト実行用

## ディレクトリ構成

```
gcsim-unofficial-clone/
├── webui/               # 静的フロントエンドファイル（出力先）
│   └── server/          # サーバー関連ファイル
│       ├── cmd/
│       │   └── devserver/       # 開発用サーバー
│       │   └── gcsim-webui/     # 本番用サーバー
│       └── scripts/
│           └── run-local-devserver.ps1  # 開発サーバー起動スクリプト
└── webui-src/           # フロントエンドソース
    ├── src/             # JavaScript/CSSソース
    ├── build.js         # ビルドスクリプト
    └── package.json
```

## クイックスタート

### 1. フロントエンドのビルド

初回のみ、またはフロントエンドを変更した場合：

```powershell
cd webui-src
npm install      # 初回のみ
npm run build    # webui/にビルド出力
```

これにより以下のファイルが生成されます：
- `webui/app.js` - バンドルされたJavaScript
- `webui/results.css` - CSSスタイル

### 2. 開発サーバーの起動

#### 方法A: スクリプトを使用（推奨）

```powershell
# リポジトリルートから実行
.\webui\server\scripts\run-local-devserver.ps1
```

このコマンドは：
- バックグラウンドでdevserverを起動
- ブラウザを自動的に開く（http://localhost:8381/ui/）

#### 方法B: フロントエンドも一緒にビルド

```powershell
.\webui\server\scripts\run-local-devserver.ps1 -BuildFrontend
```

#### 方法C: フォアグラウンドでログを表示

```powershell
.\webui\server\scripts\run-local-devserver.ps1 -Foreground
```

Ctrl+Cで停止できます。

#### 方法D: ポート変更

```powershell
.\webui\server\scripts\run-local-devserver.ps1 -Port 9000
```

### 3. サーバーの停止

#### 方法A: 停止スクリプトを使用（推奨）

```powershell
.\webui\server\scripts\stop-devserver.ps1
```

確認プロンプトが表示されます。強制停止する場合：

```powershell
.\webui\server\scripts\stop-devserver.ps1 -Force
```

#### 方法B: PIDを指定して停止

スクリプト実行時に表示されるPIDを使用：

```powershell
Stop-Process -Id <PID>
```

#### 方法C: プロセスを検索して停止

```powershell
Get-Process | Where-Object {$_.ProcessName -like '*devserver*'} | Stop-Process
```

### 4. ポート競合の自動検出

既にサーバーが起動している場合、スクリプトが自動的に検出して停止を提案します：

```
========================================
警告: Port 8381 is already in use!
Process: devserver (PID: 12345)
========================================

Kill existing process and continue? (y/N): 
```

`y`を入力すると既存プロセスを停止して新しいサーバーを起動します。

## スクリプト一覧

### run-local-devserver.ps1

開発サーバーを起動します。

| パラメータ | 説明 | デフォルト |
|-----------|------|----------|
| `-BuildFrontend` | 起動前にwebui-srcをビルド | `false` |
| `-Foreground` | フォアグラウンドで実行（ログ表示） | `false` |
| `-OpenBrowser` | ブラウザを自動で開く | `true` |
| `-Port <番号>` | サーバーポート | `8381` |
| `-Help` | ヘルプを表示 | - |

### stop-devserver.ps1

実行中の開発サーバーを停止します。

| パラメータ | 説明 | デフォルト |
|-----------|------|----------|
| `-Port <番号>` | 停止するサーバーのポート | `8381` |
| `-Force` | 確認なしで強制停止 | `false` |

### 使用例

```powershell
# 基本的な起動
.\webui\server\scripts\run-local-devserver.ps1

# サーバーを停止
.\webui\server\scripts\stop-devserver.ps1

# 停止（確認なし）
.\webui\server\scripts\stop-devserver.ps1 -Force

# ヘルプを表示
.\webui\server\scripts\run-local-devserver.ps1 -Help

# フロントエンドをビルドしてフォアグラウンドで実行
.\webui\server\scripts\run-local-devserver.ps1 -BuildFrontend -Foreground

# ポート9000でブラウザを開かずに起動
.\webui\server\scripts\run-local-devserver.ps1 -Port 9000 -OpenBrowser:$false
```

## 開発サーバー（devserver）について

`webui/server/cmd/devserver/main.go`は開発専用の軽量HTTPサーバーです。

### 提供機能

#### 1. 静的ファイル配信
- **パス**: `/ui/`
- **ディレクトリ**: `webui/`（`webui/dist/`が存在する場合はそちら）
- **内容**: フロントエンドHTML/JS/CSS

#### 2. モックAPIエンドポイント

##### `/api/simulate` (POST)
シミュレーションリクエストを受け付けます。

**リクエスト例**:
```json
{
  "config": "xiangling attack; xiangling skill;"
}
```

または生テキストでも可：
```
xiangling attack;
xiangling skill;
```

**レスポンス（非同期）**:
```json
{
  "job_id": "job-123456"
}
```

**レスポンス（同期: `?sync=true`）**:
```json
{
  "summary": { "dps": 12345.6, "duration": 60 },
  "characters": [...],
  "dps_samples": [...],
  "timeline": [...]
}
```

##### `/api/result` (GET)
ジョブの結果を取得します。

**リクエスト例**:
```
GET /api/result?id=job-123456
```

**レスポンス**:
```json
{
  "id": "job-123456",
  "status": "done",
  "result": { ... }
}
```

### CORS対応

開発環境用にCORSヘッダーが自動設定されます：
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

## トラブルシューティング

### ポートが既に使用されている

```
Port 8381 is already in use by process: go (PID: 12345)
Kill existing process and restart? (y/N)
```

`y`を入力して既存プロセスを停止するか、別のポートを指定してください。

### ビルドエラー

#### Go関連
```powershell
go version  # Go 1.22.0以上か確認
go mod tidy # 依存関係を更新
```

#### フロントエンド関連
```powershell
cd webui-src
rm -Recurse -Force node_modules  # node_modulesを削除
npm install                       # 再インストール
npm run build                     # 再ビルド
```

### ブラウザが開かない

手動でアクセス：
```
http://localhost:8381/ui/
```

### ログが見えない

フォアグラウンドモードで実行：
```powershell
.\scripts\run-local-devserver.ps1 -Foreground
```

## 開発ワークフロー

### フロントエンド開発

1. **ソース編集**: `webui-src/src/`内のファイルを編集
2. **ビルド**: `cd webui-src; npm run build`
3. **確認**: ブラウザをリロード（http://localhost:8381/ui/）

### バックエンド開発

1. **コード編集**: `webui/server/cmd/devserver/main.go`を編集
2. **再起動**:
   ```powershell
   # 既存のサーバーを停止
   Stop-Process -Name go
   
   # 再起動
   .\webui\server\scripts\run-local-devserver.ps1 -Foreground
   ```

### 統合開発

フロントエンドとバックエンドを同時に開発する場合：

```powershell
# ターミナル1: devserverをフォアグラウンドで実行
.\webui\server\scripts\run-local-devserver.ps1 -Foreground

# ターミナル2: フロントエンドの変更を監視（watch機能は未実装、手動ビルド）
cd webui-src
npm run build  # 変更のたびに実行
```

## 本番環境との違い

開発サーバーは以下の点で本番環境と異なります：

| 項目 | 開発サーバー | 本番環境 |
|-----|------------|---------|
| シミュレーション | モック（固定値） | 実際の計算エンジン |
| データベース | なし | Cloudflare Workers KV |
| 認証 | なし | 必要に応じて実装 |
| CORS | 全許可 | 制限あり |
| パフォーマンス | 最適化なし | 最適化済み |

## 次のステップ

- [ARCHITECTURE.md](./ARCHITECTURE.md) - プロジェクト全体のアーキテクチャ
- [WEBUI.md](./WEBUI.md) - WebUIの詳細仕様
- [IMPLEMENTATION_REFERENCE.md](./IMPLEMENTATION_REFERENCE.md) - キャラクター/武器/聖遺物実装ガイド

---

*Last updated: 2025-12-06*
