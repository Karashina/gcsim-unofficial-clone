# WebUIにパスワード管理機能を統合

**日時**: 2026-01-04T18:56:00+09:00  
**実施者**: GitHub Copilot (Claude Sonnet 4.5)  
**ステータス**: 確認待ち

## 要約

- WebUIにユーザー管理機能を統合しました
- 管理者はユーザー一覧から各ユーザーのパスワードをTestPass1234!にリセット可能
- 一般ユーザーは自分のパスワードを変更可能
- ナビゲーションバーに条件付きボタンを追加（管理者→「管理者」ボタン、一般ユーザー→「PW変更」ボタン）

## 変更ファイル

### 1. webui/index.html
- **変更内容**: ナビゲーションバーに「PW変更」ボタンを追加、ユーザー管理コンテナを追加
- **変更箇所**:
  - Line 1617: `<button class="navbar-tab" id="change-password-tab" style="display:none;">PW変更</button>` 追加
  - Line 1820: `<div id="user-management-container"></div>` 追加

### 2. webui-src/src/app.js
- **変更内容**: initializeAuth関数にユーザー管理機能の初期化処理を追加
- **変更箇所**:
  - Line 2386: `changePasswordTab`の取得
  - Line 2409-2418: 管理者/一般ユーザーのボタン表示制御
  - Line 2461-2465: PW変更ボタンのクリックイベントハンドラ

### 3. webui-src/src/user-management.js（前回作成）
- **機能**: ユーザー管理画面のロジック
  - showUserManagementScreen(): モーダル表示
  - resetUserPassword(): 管理者によるPWリセット
  - showChangePasswordDialog(): PW変更ダイアログ
  - handleChangePassword(): PW変更処理

### 4. webui-src/src/styles/user-management.css（前回作成）
- **機能**: ユーザー管理UIのスタイル定義

### 5. scripts/deploy_webui.ps1
- **変更内容**: 構文エラーの修正（重複コードブロックの削除）
- **変更箇所**: Line 368-369の不要なコードブロックを削除

## 実行コマンド

```powershell
# ビルド
cd c:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src
npm run build

# デプロイ
cd c:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone
.\scripts\deploy_webui.ps1
```

## 検証結果

- ビルド: ✅ PASS
  - CSS built: webui\results.css
  - JavaScript built: webui\app.js

- デプロイ: ✅ PASS
  - remote: deploy complete
  - Cloudflare cache cleared successfully

- 機能テスト: ⏳ 確認待ち
  - 管理者ログイン時に「管理者」ボタンが表示されることを確認
  - 一般ユーザーログイン時に「PW変更」ボタンが表示されることを確認
  - 管理者による他ユーザーのPWリセット機能の動作確認
  - 一般ユーザーによる自身のPW変更機能の動作確認

## UI仕様

### ナビゲーションバー（ログイン後）
- **共通表示**: ユーザー名、ログアウトボタン、テーマ切替ボタン
- **管理者（isAdmin() === true）**:
  - 「管理者」ボタン表示（data-screen="admin"）
  - 「PW変更」ボタン非表示
- **一般ユーザー**:
  - 「PW変更」ボタン表示（クリックでダイアログ表示）
  - 「管理者」ボタン非表示

### パスワードリセット機能（管理者のみ）
- アクセス: 「管理者」タブ → ユーザー一覧 → 「PWリセット」ボタン
- 動作: 選択したユーザーのパスワードを TestPass1234! に強制変更
- API: POST /api/admin/users/{id}/password
- 確認メッセージ: 「{ユーザー名} のパスワードをリセットしますか？」

### パスワード変更機能（全ユーザー）
- アクセス: ナビゲーションバーの「PW変更」ボタン
- 入力フィールド:
  - 現在のパスワード（必須）
  - 新しいパスワード（必須、8文字以上）
  - パスワード確認（必須、新しいパスワードと一致）
- バリデーション:
  - 8文字以上
  - 大文字、小文字、数字、特殊文字を含む
- API: POST /api/change-password
- エラーハンドリング:
  - 現在のパスワードが間違っている場合
  - バリデーション失敗
  - ネットワークエラー

## 技術詳細

### モジュール構成
```
webui-src/src/
├── app.js              # メインアプリケーション（ユーザー管理初期化を追加）
├── auth.js             # 認証機能（既存）
├── user-management.js  # ユーザー管理機能（新規）
└── styles/
    └── user-management.css  # ユーザー管理UI（新規）
```

### ビルドプロセス
```
webui-src/src/ → esbuild → webui/
- JavaScript: 全モジュールをapp.jsにバンドル
- CSS: 全スタイルをresults.cssに結合（user-management.css追加）
```

### デプロイフロー
1. ローカルでnpm run build実行
2. webui/ディレクトリをリモートサーバーにアップロード
3. server.envをアップロード（JWT秘密鍵等の環境変数）
4. gcsim-webuiバイナリをアップロード
5. リモートスクリプトで/var/www/htmlにデプロイ
6. nginxキャッシュクリア＆リロード
7. Cloudflareキャッシュクリア

## 課題 / 残件

- [x] HTMLにPW変更ボタン追加
- [x] app.jsに初期化処理追加
- [x] ビルド実行
- [x] デプロイスクリプトの構文エラー修正
- [x] デプロイ実行
- [ ] ブラウザでUI確認（管理者ログイン）
- [ ] ブラウザでUI確認（一般ユーザーログイン）
- [ ] PWリセット機能の動作確認
- [ ] PW変更機能の動作確認
- [ ] パスワードバリデーションの確認
- [ ] エラーハンドリングの確認

## 注意事項

### 既知の警告
デプロイ時に以下の警告が出ますが、デプロイ自体は成功しています：
```
sudo: /etc/sudoers.d/gcsim-deploy is owned by uid 1000, should be 0
```
この警告は、sudoersファイルの所有者権限に関するもので、将来的に修正が必要ですが、現時点では動作に影響していません。

### セキュリティ注意事項
- パスワードリセット機能は管理者のみ使用可能（バックエンドでロール検証）
- リセット後のパスワード（TestPass1234!）はユーザーに通知する必要あり
- 初回ログイン後にパスワード変更を促すUI実装を検討

## 備考

- ユーザー管理機能は独立したモジュールとして実装され、既存のシミュレーション機能に影響を与えません
- モーダルベースのUIで既存画面とのレイアウト競合を回避
- ダークテーマ対応済み
- レスポンシブデザイン対応

## 次のアクション

1. ブラウザで https://gcsim-uoc.linole.net にアクセス
2. karashinaアカウントでログイン
3. 「管理者」ボタンが表示されることを確認
4. ユーザー一覧でPWリセット機能をテスト
5. ログアウト→一般ユーザーでログイン
6. 「PW変更」ボタンが表示されることを確認
7. パスワード変更機能をテスト

---
**確認をお願いします。**
