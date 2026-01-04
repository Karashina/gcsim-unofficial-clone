# シミュレーション結果処理の修正（同期レスポンス対応）

**日時**: 2026-01-04T20:45:00+09:00  
**実施者**: GitHub Copilot (Claude Sonnet 4.5)  
**ステータス**: 完了

## 要約

シミュレーション実行時に「No job_id returned from server」エラーが発生していた問題を修正。バックエンドが同期的に完全な結果を返すのに対し、フロントエンドが非同期job_idシステムを期待していた設計不一致を解消。

## 問題の詳細

### 発生していた問題

1. **シミュレーションモード**: HTTP 200成功だが「No job_id returned from server」エラー
2. **Optimizerモード**: 最適化済みコンフィグは表示されるが、シミュレーション結果が空

### 根本原因

- **バックエンド**: `simulator.Run()`で同期的に実行し、完全な結果を直接返却
- **フロントエンド**: `job_id`ベースの非同期ポーリングシステムを期待
- **設計不一致**: 40行以上の不要なポーリングロジックが動作を阻害

## 実施内容

### 1. シミュレーションモードの修正

**ファイル**: `webui-src/src/app.js` (Lines 940-995)

**変更前**:
```javascript
const submitResult = await response.json();
if (!submitResult.job_id) {
    throw new Error('No job_id returned from server');
}

// 60秒間のポーリング処理（40行以上）
let attempts = 0;
const maxAttempts = 60;
while (attempts < maxAttempts) {
    // ... polling logic ...
}

if (!result) {
    throw new Error('Simulation timed out');
}
displayResults(result);
```

**変更後**:
```javascript
// Backend returns the complete simulation result synchronously
const result = await response.json();
debugLog('[WebUI] Simulation result received:', result);

loading.style.display = 'none';
runButton.disabled = false;

debugLog('[WebUI] Displaying results');
displayResults(result);
```

**効果**:
- 不要なポーリングロジック40行を削除
- レスポンスタイム大幅短縮（ポーリングオーバーヘッド排除）
- コード可読性向上

### 2. Optimizerモードの修正

**ファイル**: `webui-src/src/app.js` (Lines 120-180)

**変更前**:
```javascript
const result = await response.json();

// Display optimized config
if (result.optimized_config) {
    cmEditorOptimized.setValue(result.optimized_config);
}

// If results are available, display them
if (result.statistics) {  // ← 存在しないプロパティをチェック
    displayResults(result.statistics);
}
```

**変更後**:
```javascript
const result = await response.json();

// Display optimized config
if (result.optimized_config) {
    cmEditorOptimized.setValue(result.optimized_config);
}

// Display simulation results
// The backend returns the full simulation result with optimized_config added
debugLog('[WebUI] Displaying results...');
displayResults(result);  // ← 完全な結果を直接渡す
```

**効果**:
- 誤った`result.statistics`チェックを削除
- バックエンドが返す完全な結果構造（32フィールド）を正しく処理
- Optimizer結果が正常に表示されるように

### 3. デプロイプロセスの改善

**ファイル**: `scripts/deploy_webui.ps1` (Lines 65-105)

**変更前**:
```powershell
# Start-Processによる複雑な呼び出し
$proc = Start-Process -FilePath powershell -ArgumentList '...' -Wait -PassThru
if ($proc.ExitCode -ne 0) { Write-Warning "..." }
```

**変更後**:
```powershell
# 直接実行に変更
Push-Location $localSrcPath
npm install
if ($LASTEXITCODE -ne 0) { 
    Write-Error "npm install failed"
    exit $LASTEXITCODE
}

npm run build
if ($LASTEXITCODE -ne 0) { 
    Write-Error "npm run build failed"
    exit $LASTEXITCODE
}

# ビルド検証を追加
$builtAppJs = Join-Path $LocalWebuiPath "app.js"
if (-not (Test-Path -LiteralPath $builtAppJs)) {
    Write-Error "Build verification failed: webui/app.js not found"
    exit 6
}
Write-Log "Build verification: webui/app.js exists"
```

**効果**:
- ビルドエラーを確実に検出（エラーコード即座に停止）
- ビルド後の検証ステップ追加
- 不要なフォールバックロジック削除（BuildCmdパラメータ）
- デバッグ出力改善（ファイルサイズ表示）

## 検証結果

### シミュレーションモード
✅ HTTP 200成功
✅ 完全な結果JSONを受信（32フィールド）
✅ 結果画面に正常表示
✅ チャート描画成功

### Optimizerモード
✅ HTTP 200成功
✅ 最適化済みコンフィグが右パネルに表示
✅ シミュレーション結果が正常表示（統計、キャラDPS、ダメージ分布など）
✅ チャート描画成功

### デプロイプロセス
✅ ビルド必須化（エラー時即座に停止）
✅ ビルド検証（webui/app.js存在確認）
✅ ブラウザキャッシュ問題解消（必ず最新ファイルがデプロイされる）

## 変更ファイル

- `webui-src/src/app.js` - シミュレーション/Optimizerレスポンス処理修正
- `scripts/deploy_webui.ps1` - ビルドプロセス改善・検証追加

## 実行コマンド

```powershell
# デプロイ実行
.\scripts\deploy_webui.ps1

# ブラウザでスーパーリロード（キャッシュクリア）
# Ctrl+F5 または Ctrl+Shift+R
```

## 技術的詳細

### バックエンドの実装（参考）

**simulateHandler** (`cmd/gcsim-webui/main.go`):
```go
result, err := simulator.Run(ctx, opts)
data, err := result.MarshalJSON()
w.Write(data)  // 完全な結果を直接返却（job_idなし）
```

**optimizeHandler** (`cmd/gcsim-webui/main.go`):
```go
result, err := simulator.Run(ctx, opts)
resultMap["optimized_config"] = string(optimizedConfig)
finalData, err := json.Marshal(resultMap)
w.Write(finalData)  // 完全な結果 + optimized_configフィールド
```

### レスポンス形式

**シミュレーションモード**:
```json
{
  "schema_version": {...},
  "sim_version": "ca46285d...",
  "character_details": [...],
  "dps": 12345.67,
  "statistics": {...},
  // ... 30以上の他のフィールド
}
```

**Optimizerモード**:
```json
{
  "schema_version": {...},
  "optimized_config": "nefer char lvl=90/90...",
  "character_details": [...],
  "dps": 12345.67,
  // ... 30以上の他のフィールド
}
```

## 課題 / 残件

- なし（すべて動作確認済み）

## 備考

### 設計判断の理由

1. **同期処理を維持**: 
   - gcsimシミュレーターは180秒タイムアウト内で完了
   - 非同期ジョブシステムのオーバーヘッド不要
   - コード複雑性削減

2. **job_idシステム削除**:
   - バックエンドにjob_id管理機構が存在しない
   - `/api/result`エンドポイントも未実装
   - フロントエンドの期待値が実装と不一致

3. **デプロイ検証強化**:
   - ブラウザキャッシュ問題で同じエラーが繰り返された経験
   - ビルドプロセスを必須化し、失敗時即座に停止
   - ファイル存在確認でデプロイ成功を保証

### パフォーマンス改善

- **レスポンスタイム**: 1秒以上のポーリング遅延を排除
- **コードサイズ**: 約50行削減（不要なポーリングロジック）
- **CPU使用率**: ポーリングによる無駄なHTTPリクエスト削除

### 今後の拡張性

現在の同期実装で十分だが、将来的に非同期処理が必要になった場合の対応方針：

1. **Server-Sent Events (SSE)**: リアルタイム進捗通知
2. **WebSocket**: 双方向通信
3. **Job Queue**: Redis + Celeryなどのジョブキューシステム

ただし、現時点では180秒タイムアウトで十分高速なため、実装不要。
