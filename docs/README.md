# Docs フォルダ構成

このフォルダには、gcsimの実装、仕様、認証システム、開発ガイドに関するドキュメントが格納されています。

---
**開発者向け技術情報・運用手順はすべてこのdocs/以下に集約されています。**
プロジェクトルートやサブディレクトリのREADME.mdは簡易案内のみです。詳細は必ずこのフォルダの各種ドキュメントをご参照ください。

## ドキュメント一覧

### 開発・実装ガイド (Development & Implementation)

- **[AI_CHARACTER_IMPLEMENTATION_GUIDE.md](AI_CHARACTER_IMPLEMENTATION_GUIDE.md)**
    - AIアシスタント向けのキャラクター実装ガイド。
    - 必要なファイル構成、システム登録手順、テスト方法。
    - **Ascension Data Generation** (asc.md内容) を含みます。

- **[ARCHITECTURE.md](ARCHITECTURE.md)**
    - gcsimの全体アーキテクチャ、フレームワーク、主要コンポーネントの解説。

- **[IMPLEMENTATION_REFERENCE.md](IMPLEMENTATION_REFERENCE.md)**
    - アクション、ステータス、属性、クールダウンなどの詳細な実装リファレンス。

- **[GCSL_REFERENCE.md](GCSL_REFERENCE.md)**
    - gcsimスクリプト言語 (GCSL) の構文、関数、仕様リファレンス。

- **[WEAPON_PROTO_DATA_GUIDE.md](WEAPON_PROTO_DATA_GUIDE.md)**
    - 武器データの実装、プロトタイプデータの扱い方に関するガイド。

- **[CORE_MECHANICS_LUNA_VI.md](CORE_MECHANICS_LUNA_VI.md)**
    - **(New/Luna VI)** Lunar VI 実装に伴うコアメカニクス変更点（Attack Modの仕様変更など）の解説。

### WebUI & 認証 (WebUI & Authentication)

- **[WEBUI.md](WEBUI.md)**
    - WebUIの機能、セットアップ、開発に関するドキュメント。

- **[AUTHENTICATION_GUIDE.md](AUTHENTICATION_GUIDE.md)**
    - **(統合版)** 認証システムの概要、セットアップ、デプロイ、パスワード変更、トラブルシューティング。
    - 元の `AUTHENTICATION.md`, `DEPLOYMENT.md`, `PASSWORD_CHANGE_GUIDE.md` 等を統合。

### 環境構築・履歴 (Setup & History)

- **[LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md)**
    - ローカル開発環境の構築手順。

- **[PROJECT_HISTORY.md](PROJECT_HISTORY.md)**
    - プロジェクトの変更履歴、作業ログのまとめ (Luna VI 実装など)。
