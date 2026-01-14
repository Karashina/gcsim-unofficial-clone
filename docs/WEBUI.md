# WebUI 詳細ドキュメント

## 概要

gcsim WebUIは原神シミュレータのWebインターフェースです。ブラウザ上でGCSL設定を記述・実行し、結果をグラフや表で確認できます。

---

## アーキテクチャ

### フロントエンド

**技術スタック:**
- Vanilla JavaScript (ES2017)
- CodeMirror 5 (エディタ)
- Chart.js (グラフ描画)
- esbuild (ビルドツール)

**ファイル構成:**
```
webui/
├── index.html              # メインHTML（1350行）
├── app.js                  # バンドルされたJS
├── results.css             # スタイル
├── jp_mappings.generated.js # 日本語マッピング
└── assets/                 # 画像等

webui-src/
├── build.js               # ビルドスクリプト
├── package.json
└── src/
    ├── app.js             # メインアプリケーション（2134行）
    ├── localization.js    # 多言語化ヘルパー
    ├── screens/           # 画面モジュール
    │   ├── results-main.js
    │   ├── results-characters.js
    │   ├── results-statistics.js
    │   ├── results-targets.js
    │   └── results-charts.js
    └── styles/
        └── results.css
```

### バックエンド

**サーバー実装:**
- `webui/server/cmd/gcsim-webui/main.go` - 本番用サーバー（ポート8382）
- `webui/server/cmd/devserver/main.go` - 開発用サーバー（ポート8381）

**APIエンドポイント:**
- `POST /api/simulate` - シミュレーション実行
- `POST /api/optimize` - サブスタット最適化
- `GET /healthz` - ヘルスチェック
- `GET /ui/*` - 静的ファイル配信

---

## 画面構成

### 1. シミュレート画面 (screen-simulate)

**機能:**
- GCSLエディタ（CodeMirror統合）
- シンタックスハイライト
- Ctrl/Cmd+Enterで実行
- エラー表示エリア
- ローディングインジケーター

**デフォルト設定:**
```gcsl
nefer char lvl=90/90 cons=0 talent=9,9,9;
nefer add weapon="blackmarrowlantern" refine=5 lvl=90/90;
nefer add set="notsu" count=4;
nefer add stats hp=4780 atk=311 em=187 em=187 cd=0.622;
# ... (省略)

options swap_delay=12 iteration=1000;
target lvl=100 resist=0.1 radius=2 pos=2.1,1.5 hp=999999999;
energy every interval=480,720 amount=1;

active nahida;

for let i=0; i<4; i=i+1 {
  nahida skill;
  if .nahida.burst.ready && .nahida.energymax {
    nahida burst;
  }
  # ... (省略)
}
```

### 2. 最適化画面 (screen-optimizer)

**機能:**
- 左パネル: 元の設定入力
- 右パネル: 最適化後の設定表示
- 実行ボタン
- サブスタット配分の最適化

### 3. 結果画面 (screen-results)

**セクション構成:**

#### 3.1 統計サマリー
- **DPS** (Damage Per Second) - 秒間ダメージ
- **EPS** (Energy Per Second) - 秒間エネルギー
- **RPS** (Reactions Per Second) - 秒間元素反応回数
- **HPS** (Healing Per Second) - 秒間回復量
- **SHP** (Shield HP Per Second) - 秒間シールド量
- **Duration** - シミュレーション時間

各統計は平均値±標準偏差で表示。

#### 3.2 ターゲット情報
- ターゲット名
- レベル
- HP
- 元素耐性（元素別）

#### 3.3 キャラクター情報
各キャラクターカードに表示:
- **基本情報**
  - 日本語名 / 英語名
  - レベル / 最大レベル
  - 命ノ星座 (Cx)
  - 天賦レベル (通常/スキル/爆発)
  
- **装備**
  - 武器名（日本語/英語）
  - 武器レベル / 精錬ランク
  - 聖遺物セット名（装備数）

- **ステータス詳細**
  - HP, ATK, DEF
  - 元素熟知 (EM)
  - 元素チャージ効率 (ER)
  - 会心率 (CR) / 会心ダメージ (CD)
  - ダメージバフ (元素別)

#### 3.4 グラフ
- **キャラクターDPS円グラフ** - キャラクター別ダメージ割合
- **ダメージソース横棒グラフ** - スキル/爆発/通常攻撃等の内訳
- **Raw statistics JSON** - デバッグ用生データ（折りたたみ）

---

## フロントエンド詳細

### メインアプリケーション (app.js)

#### 初期化処理 (DOMContentLoaded)

```javascript
document.addEventListener('DOMContentLoaded', function() {
    // 1. 画面ナビゲーション設定
    setupScreenNavigation();
    
    // 2. モード切り替え設定
    setupModeSwitch();
    
    // 3. CodeMirrorエディタ初期化
    cmEditor = CodeMirror.fromTextArea(textarea, {
        mode: 'gcsl',
        lineNumbers: true,
        theme: isDark ? 'material' : 'default',
        tabSize: 2,
        indentWithTabs: false
    });
    
    // 4. キーバインド設定
    cmEditor.addKeyMap({
        'Ctrl-Enter': runSimulation,
        'Cmd-Enter': runSimulation
    });
    
    // 5. テーマ切り替えボタン
    setupThemeToggle();
    
    // 6. デフォルト設定ロード
    loadDefaultConfig();
});
```

#### シミュレーション実行フロー

```javascript
async function runSimulation() {
    // 1. 設定取得
    const config = cmEditor.getValue();
    
    // 2. UI更新（ローディング表示）
    loading.style.display = 'block';
    errorMsg.style.display = 'none';
    resultsContainer.style.display = 'none';
    
    // 3. API呼び出し
    const response = await fetch('/api/simulate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ config })
    });
    
    // 4. エラーハンドリング
    if (!response.ok) {
        const error = await response.json();
        handleError(error);
        return;
    }
    
    // 5. 結果表示
    const result = await response.json();
    displayResults(result);
    
    // 6. 結果画面に切り替え
    document.querySelector('.navbar-tab[data-screen="results"]').click();
}
```

#### エラーハンドリング

```javascript
function handleError(error) {
    let message = error.message || error.error;
    
    // パースエラーの場合
    if (error.parse_errors) {
        message = 'パースエラー:\n';
        error.parse_errors.forEach(pe => {
            message += `行 ${pe.line}: ${pe.message}\n`;
        });
    }
    
    errorMsg.textContent = message;
    errorMsg.style.display = 'block';
}
```

### 結果表示モジュール

#### results-main.js - オーケストレーター

```javascript
export function displayResults(result) {
    // 1. 結果コンテナ表示
    const resultsContainer = document.getElementById('results-container');
    resultsContainer.style.display = 'block';
    
    // 2. 各セクション表示
    displayTargets(result);
    displayCharacters(result);
    displayStatistics(result);
    displayCharts(result, resultsContainer);
    
    // 3. スムーズスクロール
    setTimeout(() => {
        resultsContainer.classList.add('visible');
        resultsContainer.scrollIntoView({ behavior: 'smooth' });
    }, 100);
}
```

#### results-characters.js - キャラクター表示

```javascript
export function displayCharacters(result) {
    const container = document.getElementById('characters-list');
    container.innerHTML = '';
    
    const gridDiv = document.createElement('div');
    gridDiv.className = 'characters-grid';
    
    result.character_details.forEach((char, idx) => {
        const charCard = createCharacterCard(char, idx);
        gridDiv.appendChild(charCard);
    });
    
    container.appendChild(gridDiv);
}

function createCharacterCard(char, idx) {
    const charDiv = document.createElement('div');
    charDiv.className = 'char-card';
    
    // 日本語名取得
    const name = toJPCharacter(char.name);
    const weaponJP = toJPWeapon(char.weapon?.name);
    
    // ステータス取得
    const stats = extractStats(char.snapshot_stats);
    
    // HTMLビルド
    charDiv.innerHTML = `
        <div class="char-header">
            <div class="char-name-line">
                ${name}
                <span class="char-constellation">C${char.cons}</span>
            </div>
        </div>
        <!-- 省略 -->
    `;
    
    return charDiv;
}
```

#### results-statistics.js - 統計表示

```javascript
export function displayStatistics(result) {
    const stats = result.statistics || {};
    
    // 各統計を更新
    updateStat('stat-dps', stats.dps?.mean, stats.dps?.sd);
    updateStat('stat-eps', stats.eps?.mean, stats.eps?.sd);
    updateStat('stat-rps', stats.rps?.mean, stats.rps?.sd);
    // ...
}

function updateStat(elementId, value, stdev) {
    const element = document.getElementById(elementId);
    element.innerHTML = formatStatWithStdev(value, stdev);
}

function formatStatWithStdev(value, stdev) {
    const formatted = value.toLocaleString('ja-JP', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    });
    
    if (stdev > 0) {
        const stdFormatted = stdev.toLocaleString('ja-JP', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        });
        return `${formatted}<br><span class="stat-stdev">±${stdFormatted}</span>`;
    }
    
    return formatted;
}
```

#### results-charts.js - グラフ描画

```javascript
export function displayCharts(result, resultsContainer) {
    // Chart.js存在確認
    if (typeof Chart === 'undefined') {
        console.warn('Chart.js not available');
        showChartPlaceholders();
        return;
    }
    
    // デバッグパネル作成
    createDebugPanel(result, resultsContainer);
    
    // グラフ描画
    renderCharacterDPSChart(result);
    renderSourceDPSChart(result);
}

function createDebugPanel(result, container) {
    const panel = document.createElement('details');
    panel.id = 'raw-stats-panel';
    
    const summary = document.createElement('summary');
    summary.textContent = 'Raw statistics JSON (debug)';
    panel.appendChild(summary);
    
    const pre = document.createElement('pre');
    pre.textContent = JSON.stringify(result.statistics, null, 2);
    panel.appendChild(pre);
    
    container.insertBefore(panel, container.firstChild);
}
```

### ローカライゼーション (localization.js)

```javascript
/**
 * キャラクター名を日本語に変換
 */
export function toJPCharacter(key) {
    if (!key) return key;
    return window.CHAR_TO_JP?.[key] || key;
}

/**
 * 武器名を日本語に変換
 */
export function toJPWeapon(key) {
    if (!key) return key;
    return window.WEAPON_TO_JP?.[key] || key;
}

/**
 * 聖遺物名を日本語に変換
 */
export function toJPArtifact(key) {
    if (!key) return key;
    return window.ARTIFACT_TO_JP?.[key] || key;
}
```

---

## バックエンド詳細

### cmd/gcsim-webui/main.go

#### サーバー起動

```go
func main() {
    var addr string
    flag.StringVar(&addr, "addr", ":8382", "address to listen on")
    flag.Parse()

    mux := http.NewServeMux()
    
    // APIエンドポイント
    mux.HandleFunc("/api/simulate", simulateHandler)
    mux.HandleFunc("/api/optimize", optimizeHandler)
    mux.HandleFunc("/healthz", healthzHandler)
    
    // 静的ファイル配信
    fs := http.FileServer(http.Dir("webui"))
    mux.Handle("/ui/", http.StripPrefix("/ui/", fs))

    srv := &http.Server{Addr: addr, Handler: mux}
    log.Printf("gcsim-webui: listening on %s", addr)
    srv.ListenAndServe()
}
```

#### simulateHandler

```go
func simulateHandler(w http.ResponseWriter, r *http.Request) {
    // 1. CORS対応
    if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.WriteHeader(http.StatusNoContent)
        return
    }
    
    // 2. リクエストボディ読み込み
    var payload struct {
        Config string `json:"config"`
    }
    body, _ := io.ReadAll(r.Body)
    json.Unmarshal(body, &payload)
    
    // 3. 一時ファイル保存
    tmpFile := filepath.Join(os.TempDir(), 
        "gcsim_webui_config_" + time.Now().Format("20060102_150405") + ".txt")
    os.WriteFile(tmpFile, []byte(payload.Config), 0o600)
    defer os.Remove(tmpFile)
    
    // 4. シミュレーション実行（180秒タイムアウト）
    ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
    defer cancel()
    
    opts := simulator.Options{ConfigPath: tmpFile}
    result, err := simulator.Run(ctx, opts)
    
    // 5. エラーハンドリング
    if err != nil {
        handleSimulateError(w, err)
        return
    }
    
    // 6. レスポンス返却
    data, _ := result.MarshalJSON()
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
```

#### エラーハンドリング

```go
func handleSimulateError(w http.ResponseWriter, err error) {
    // タイムアウト
    if errors.Is(err, context.DeadlineExceeded) {
        w.WriteHeader(http.StatusGatewayTimeout)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "timeout",
            "message": "simulation exceeded 180s timeout"
        })
        return
    }
    
    // パースエラー（行番号抽出）
    errStr := err.Error()
    re := regexp.MustCompile(`ln(\d+):\s*(.+)`)
    matches := re.FindAllStringSubmatch(errStr, -1)
    
    if len(matches) > 0 {
        out := map[string]interface{}{
            "error": "parse error",
            "message": errStr,
        }
        pe := make([]map[string]interface{}, 0)
        for _, m := range matches {
            line, _ := strconv.Atoi(m[1])
            pe = append(pe, map[string]interface{}{
                "line": line,
                "message": m[2],
            })
        }
        out["parse_errors"] = pe
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(out)
        return
    }
    
    // その他のエラー
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error":"simulation failed"}`))
}
```

#### optimizeHandler

```go
func optimizeHandler(w http.ResponseWriter, r *http.Request) {
    // simulate同様の処理 + 最適化後の設定を返却
    
    // ... (省略) ...
    
    // 最適化後の設定をファイルから読み込み
    optimizedConfig, _ := os.ReadFile(tmpFile)
    
    // 結果JSONに追加
    var resultMap map[string]interface{}
    json.Unmarshal(data, &resultMap)
    resultMap["optimized_config"] = string(optimizedConfig)
    
    finalData, _ := json.Marshal(resultMap)
    w.Write(finalData)
}
```

---

## 日本語マッピング

### 生成プロセス

```
[CSVファイル]
  ├─ cmd/gcsim/chatracterData/charactertoJP.csv
  ├─ cmd/gcsim/artifactData/artifact.csv
  └─ cmd/gcsim/weaponData/*.csv
      ↓
[cmd/generate-webui-mappings/main.go]
  1. CSVファイル読み込み
  2. map[string]string構築
  3. JavaScriptコード生成
      ↓
[webui/jp_mappings.generated.js]
  window.CHAR_TO_JP = {...}
  window.WEAPON_TO_JP = {...}
  window.ARTIFACT_TO_JP = {...}
```

### CSV形式

**charactertoJP.csv:**
```csv
日本語名,英語キー
香菱,xiangling
甘雨,ganyu
神里綾華,ayaka
```

**artifact.csv:**
```csv
日本語名,英語キー
燃え盛る炎の魔女,crimson
剣闘士のフィナーレ,gladiator
絶縁の旗印,emblem
```

### 生成コード例

```javascript
(function(){
window.CHAR_TO_JP = Object.assign(window.CHAR_TO_JP || {}, {
    "xiangling": "香菱",
    "ganyu": "甘雨",
    "ayaka": "神里綾華",
    // ...
});
window.WEAPON_TO_JP = Object.assign(window.WEAPON_TO_JP || {}, {
    "staffofhoma": "護摩の杖",
    "mistsplitterreforged": "霧切の廻光",
    // ...
});
window.ARTIFACT_TO_JP = Object.assign(window.ARTIFACT_TO_JP || {}, {
    "crimson": "燃え盛る炎の魔女",
    "gladiator": "剣闘士のフィナーレ",
    // ...
});
})();
```

---

## ビルド・デプロイ

### ローカルビルド

```powershell
# 1. マッピング生成
go run cmd/generate-webui-mappings/main.go

# 2. フロントエンドビルド
cd webui-src
npm install
npm run build

# 3. バックエンドビルド（オプション）
go build -o gcsim-webui.exe ./cmd/gcsim-webui
```

### 開発サーバー起動

```powershell
# 方法1: gcsim-webui (実際のシミュレーション実行)
go run cmd/gcsim-webui/main.go
# → http://localhost:8382/ui/

# 方法2: devserver (モックデータ)
go run cmd/devserver/main.go
# → http://localhost:8381/ui/
```

### 本番デプロイ

```powershell
# 完全自動デプロイ
.\scripts\deploy_webui.ps1

# カスタムパラメータ
.\scripts\deploy_webui.ps1 `
  -Server "192.168.1.233" `
  -User "uocuser" `
  -RemotePath "/var/www/html" `
  -KeyFile "C:\path\to\ssh_key" `
  -ReloadNginx `
  -ClearCloudflareCache
```

詳細は `scripts/DEPLOY_WEBUI.md` を参照。

---

## テーマシステム

### CSS変数定義

```css
:root {
    --primary-1: #ffc14d;
    --primary-2: #ffb347;
    --bg: #f5f5f5;
    --card-bg: #ffffff;
    --text: #222;
    --muted: #666;
}

[data-theme="dark"] {
    --primary-1: #ffc14d;
    --primary-2: #ffb347;
    --bg: #1a1a1a;
    --card-bg: #2a2a2a;
    --text: #e0e0e0;
    --muted: #999;
}
```

### テーマ切り替え

```javascript
const themeBtn = document.getElementById('theme-toggle');
themeBtn.addEventListener('click', () => {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? '' : 'dark';
    
    if (next) {
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('gcsim_theme', next);
    } else {
        document.documentElement.removeAttribute('data-theme');
        localStorage.removeItem('gcsim_theme');
    }
    
    // CodeMirrorテーマも変更
    cmEditor.setOption('theme', next ? 'material' : 'default');
});
```

---

## パフォーマンス最適化

### フロントエンド

1. **遅延描画**
   - Chart.jsのグラフは結果画面表示時に初回描画
   - `visibility: hidden` → `visible` で表示タイミング制御

2. **DOM操作の最小化**
   - `createDocumentFragment()` 使用（未実装）
   - イベント委譲で動的要素対応

3. **Canvas最適化**
   - `position: absolute` でレイアウト影響を分離
   - `overflow: hidden` で再描画範囲制限

### バックエンド

1. **並列実行**
   - `pkg/worker` でマルチコア活用
   - Goルーチンプール

2. **タイムアウト管理**
   - context.WithTimeout()使用
   - 長時間実行の防止

3. **一時ファイル削除**
   - `defer os.Remove()` で確実にクリーンアップ

---

## デバッグ

### フロントエンド

```javascript
// app.js内のDEBUGフラグ
const DEBUG = true;  // ← trueに変更

function debugLog(...args) {
    if (DEBUG && console && console.log) {
        console.log(...args);
    }
}
```

**ログ出力例:**
```
[WebUI] Starting simulation...
[WebUI] Config length: 1234
[WebUI] Sending request to /api/simulate
[WebUI] Response status: 200
[WebUI] Simulation result: {...}
[WebUI] Displaying results...
```

### バックエンド

```go
// ログ出力
log.Printf("simulateHandler: config length=%d", len(payload.Config))
log.Printf("simulateHandler: simulation failed: %v", err)
```

### ブラウザDevTools

- **Network**: API通信確認
- **Console**: JavaScriptエラー確認
- **Elements**: DOM構造確認
- **Performance**: パフォーマンス計測

---

## トラブルシューティング

### エディタが表示されない

**原因:** CodeMirror初期化失敗

**対処:**
```javascript
// app.jsで確認
if (!cmEditor) {
    console.error('CodeMirror not initialized');
}
```

### グラフが表示されない

**原因:** Chart.js読み込み失敗

**対処:**
```javascript
if (typeof Chart === 'undefined') {
    console.error('Chart.js not loaded');
}
```

index.htmlでscriptタグ確認:
```html
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
```

### 日本語が表示されない

**原因:** マッピングファイル未生成

**対処:**
```powershell
go run cmd/generate-webui-mappings/main.go
```

ブラウザコンソールで確認:
```javascript
console.log(window.CHAR_TO_JP);
// Object { xiangling: "香菱", ... }
```

### APIエラー

**原因:** バックエンド未起動

**対処:**
```powershell
go run cmd/gcsim-webui/main.go
```

ブラウザコンソールで確認:
```
Failed to fetch
net::ERR_CONNECTION_REFUSED
```

---

## セキュリティ

### CORS設定

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

本番環境では特定オリジンに制限推奨:
```go
w.Header().Set("Access-Control-Allow-Origin", "https://gcsim-uoc.linole.net")
```

### 一時ファイル

```go
// 0o600 = 所有者のみ読み書き可能
os.WriteFile(tmpFile, []byte(config), 0o600)
defer os.Remove(tmpFile)
```

### 入力検証

- ファイルサイズ制限（未実装）
- 文字列長制限（未実装）
- インジェクション対策（GCSL parserで対応）

---

## 将来の拡張

### 機能追加案

1. **リアルタイムプレビュー**
   - WebSocket使用
   - 設定変更時の即時反映

2. **設定テンプレート**
   - プリセット管理
   - ローカルストレージ保存

3. **結果比較機能**
   - 複数シミュレーション結果の並列表示
   - 差分ハイライト

4. **エクスポート機能**
   - PDF出力
   - PNG画像保存
   - CSV/Excel出力

5. **共有機能**
   - URL共有（設定エンコード）
   - クラウド保存

### 技術改善案

1. **TypeScript移行**
   - 型安全性向上
   - IDE補完強化

2. **フレームワーク導入**
   - React/Vue/Svelte
   - コンポーネント再利用性向上

3. **モジュール分割**
   - Lazy loading
   - Code splitting

4. **テスト追加**
   - Jest (ユニットテスト)
   - Playwright (E2Eテスト)

5. **アクセシビリティ**
   - WAI-ARIA対応
   - キーボード操作改善
   - スクリーンリーダー対応

---

*Last updated: 2025-12-06*
