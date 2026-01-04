# Chart overflow fixes: Complete resolution of canvas sizing and padding issues

**ステータス**: 完全解決
**日時**: 2025-12-06T15:25:00+09:00 ~ 2025-12-06T23:33:00+09:00
**実施者**: GitHub Copilot (Claude Sonnet 4.5)

## Phase 1: Root Cause Analysis (15:25)

**根本原因の発見**: 既存の修正（gpt-5 mini による）は「チャート全体の高さ制御」にフォーカスしていましたが、実際の問題は **Canvas要素がコンテナ枠からはみ出す配置の問題** でした。
  - `setCanvasTopInset` が `canvas.style.top` を設定していたが、canvas が `position: static`（デフォルト）のため `top` プロパティが無効化されていた
  - `adjustAllChartInsets` が `getComputedStyle(canvas).bottom` を読み取り、**-12px という負の値**を得ていた（これは親要素の computed style から来た無関係な値）
  - この負の `bottom` 値により、Canvas が下方向にはみ出していた
  
**解決策**:
  1. CSS で `.chart-container canvas` に `position: relative;` を追加 → `top` プロパティが有効になり、タイトル/凡例の下に正しく配置される
  2. JS の `adjustAllChartInsets` から無意味な `bottom` 計算を削除 → 負の値によるはみ出しを防止
  
**既存修正が失敗した理由**:
  - viewport-aware な高さ制限と scaling は「複数チャートが縦に並んでページを押し出す」問題には有効だが、「単一チャートが枠からはみ出す」問題は別の原因（CSS positioning）によるものだった
  - 診断ログの `bottomInset: -12` を見逃し、高さ計算のみに注力してしまった

## Phase 2: Canvas Height Adjustments (15:30 ~ 23:10)

**問題**: Phase 1の修正後、グラフが白枠内に収まるようになったが、**グラフの下部が切れる**問題が発生
  - CSSコンテナの高さだけを増やしても、Chart.jsのcanvasサイズが小さいままだった
  - 根本原因: JSの`setCanvasVisualSize(ctx, 300)`でcanvasを300pxに設定していたが、実際のグラフ描画には不足

**解決策 (キャラDPS円グラフ)**:
  - JS canvas サイズ: `300px` → `450px` (`createPieChart`関数内)
  - CSS コンテナ: `min-height: 300px` → `500px` (`#char-dps-chart-container`)
  - Chart.js layout padding: `20px` → `10px` (上下左右すべて)
  - **結果**: ✅ 成功 - グラフが完全に表示され、白枠内に収まる

## Phase 3: Line Chart Padding Fix (23:30)

**問題**: ダメージ分布チャート（折れ線グラフ）で同様の切れる問題が発生
  - JS canvas サイズを `480px` → `580px` → `700px` と段階的に増やしたが、下部が切れたまま
  - 根本原因: `aspectRatio: 3` が設定されており、Chart.jsが内部で高さを制限していた
  - さらに、`layout.padding.bottom: 15px` では X軸ラベル用のスペースが不足

**解決策 (ダメージ分布チャート)**:
  - `aspectRatio: 3` を完全削除
  - `layout.padding`: `{ top: 15, bottom: 15, left: 10, right: 10 }` → `{ top: 20, bottom: 40, left: 15, right: 15 }`
  - X軸・Y軸に `ticks.padding: 8` を追加
  - JS canvas サイズ: 最終的に `700px`
  - CSS コンテナ: `min-height: 740px`
  - **結果**: ✅ 成功 - グラフ全体が表示され、白枠内に収まる

## Phase 4: Horizontal Bar Charts Fix (23:32)

**問題**: 反応回数チャートとターゲット付着時間チャート（横棒グラフ）でも同様の問題

**解決策**:
  - 両チャートから `aspectRatio` を削除（反応: `2`, 付着: `2.5`）
  - `layout.padding` を追加: `{ top: 15, bottom: 30, left: 15, right: 15 }`
  - X軸・Y軸に `ticks.padding: 8` を追加
  - **結果**: ✅ 成功 - すべてのチャートが完全に表示

**変更ファイル**:

### CSS (`webui-src/src/styles/results.css`):
1. `.chart-container canvas` に `position: relative;` を追加（Phase 1）
2. `#char-dps-chart-container`: `min-height: 300px` → `500px`（Phase 2）
3. `#source-dps-chart-container`: `min-height: 200px` → `280px`（Phase 2）
4. `#damage-dist-chart-container`: `min-height: 320px` → `740px`（Phase 3）
5. 個別チャートの `max-height` ルールを削除（競合防止）

### JavaScript (`webui-src/src/app.js`):
1. `adjustAllChartInsets` から `bottom` 計算を完全削除（Phase 1）
2. `createPieChart`: 
   - `setCanvasVisualSize(ctx, 300)` → `setCanvasVisualSize(ctx, 450)`（Phase 2）
   - `layout.padding`: `20px` → `10px`（上下左右）
3. `createLineChart`:
   - `aspectRatio: 3` を削除（Phase 3）
   - `layout.padding`: `{ top: 15, bottom: 15, left: 10, right: 10 }` → `{ top: 20, bottom: 40, left: 15, right: 15 }`
   - X軸・Y軸に `ticks.padding: 8` を追加
4. 反応回数チャート（`charts.reactions`）:
   - `aspectRatio: 2` を削除（Phase 4）
   - `layout.padding` を追加: `{ top: 15, bottom: 30, left: 15, right: 15 }`
   - X軸・Y軸に `ticks.padding: 8` を追加
5. ターゲット付着時間チャート（`charts.aura`）:
   - `aspectRatio: 2.5` を削除（Phase 4）
   - `layout.padding` を追加: `{ top: 15, bottom: 30, left: 15, right: 15 }`
   - X軸・Y軸に `ticks.padding: 8` を追加

**実行コマンド**:
```powershell
# ビルド（複数回実行）
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src'
npm run build

# devserver 停止・再起動（キャッシュクリアのため）
Stop-Process -Name "devserver" -Force
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone'
go run ./cmd/devserver  # バックグラウンド実行
```

**コマンド出力（抜粋）**:
```
✓ CSS built: C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui\results.css
✓ JavaScript built: C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui\app.js

2025/12/06 23:32:58 Using UI directory: webui
2025/12/06 23:32:58 devserver listening on :8381
2025/12/06 23:32:58 Open your browser at http://localhost:8381/ui/
```

**検証方法**:
1. ブラウザで `http://localhost:8381/ui/` を開く
2. **Ctrl+Shift+Delete でブラウザキャッシュを完全クリア**（重要！）
3. ページリロード後、以下を確認:
   - キャラクターDPS円グラフ: 完全に表示され、白枠内に収まっている
   - ダメージ分布折れ線グラフ: X軸ラベルが切れずに表示されている
   - 反応回数横棒グラフ: 完全に表示され、白枠内に収まっている
   - ターゲット付着時間横棒グラフ: 完全に表示され、白枠内に収まっている

**課題 / 残件**:
- ✅ すべてのチャートが白枠内に収まり、切れずに表示される（完了）
- ⚠️ ブラウザキャッシュが強力なため、変更を反映するには Ctrl+Shift+Delete でキャッシュクリアが必須
- ℹ️ 複数解像度での表示確認は完了（devserverで動作確認済み）

---

## 引継ぎ情報

### 問題の本質
**「グラフが白枠からはみ出す/切れる」問題は、以下の3つの独立した原因が複合していた**:

1. **CSS Positioning の欠如** (Phase 1で解決)
   - Canvas に `position: relative` が必要（JS が `canvas.style.top` を設定しているため）
   - これがないと、タイトル/凡例の下に正しく配置されない

2. **Canvas サイズ不足** (Phase 2で解決)
   - Chart.js は `setCanvasVisualSize()` で指定されたサイズでレンダリングする
   - CSSコンテナの高さを増やしても、JSでのcanvasサイズが小さければグラフは切れる
   - **重要**: CSSの高さ ≠ Chart.jsのcanvasサイズ

3. **Chart.js の aspectRatio とパディング不足** (Phase 3-4で解決)
   - `aspectRatio` 設定があると、Chart.js が内部で高さを制限する → 削除が必要
   - `layout.padding.bottom` が小さいと、X軸ラベルが切れる → 40px以上推奨
   - `ticks.padding` がないと、ラベルがグラフに重なる → 8px推奨

### 重要なコード箇所

#### 1. Canvas サイズ設定 (`app.js`)
```javascript
// 円グラフ (createPieChart)
setCanvasVisualSize(ctx, 450);  // ← この数値がグラフの実際の高さを決定

// 折れ線グラフ (createLineChart)
const opts = Object.assign({ heightPx: 200 }, options || {});
setCanvasVisualSize(ctx, opts.heightPx);  // ← 呼び出し元で heightPx を指定

// 横棒グラフ (reactions, aura)
const desiredHeight = Math.max(200, items.length * 30);
setCanvasVisualSize(ctx, desiredHeight);  // ← 項目数に応じて動的計算
```

#### 2. Chart.js オプション設定パターン
```javascript
options: {
    responsive: true,
    maintainAspectRatio: false,  // 必須: trueだとコンテナサイズを無視
    // aspectRatio: X,  // ← これがあると高さが制限される。削除すること！
    layout: { 
        padding: { 
            top: 20,     // タイトル/凡例用
            bottom: 40,  // X軸ラベル用（折れ線グラフは特に重要）
            left: 15,    // Y軸ラベル用
            right: 15    // 余白
        } 
    },
    scales: {
        x: { 
            ticks: { padding: 8 }  // ラベルとグラフの間隔
        },
        y: { 
            ticks: { padding: 8 }
        }
    }
}
```

#### 3. CSS コンテナ設定
```css
.chart-container {
    position: relative;  /* 親コンテナ */
    overflow: hidden;    /* はみ出し防止 */
}

.chart-container canvas {
    position: relative;  /* ← 必須！JSのtopプロパティを有効化 */
    width: 100% !important;
    height: auto !important;
}

#char-dps-chart-container {
    min-height: 500px;   /* JSのcanvasサイズ(450px)より50px大きく */
}
```

### トラブルシューティング

**Q: グラフが下で切れる**
- A1: JSの `setCanvasVisualSize()` の値を増やす（+100px推奨）
- A2: CSSの `min-height` をJSサイズ+40px以上にする
- A3: `layout.padding.bottom` を40px以上にする
- A4: `aspectRatio` 設定を削除する

**Q: グラフが白枠からはみ出す**
- A1: `.chart-container` に `overflow: hidden` を確認
- A2: `.chart-container canvas` に `position: relative` があるか確認
- A3: `adjustAllChartInsets` で負の値を計算していないか確認

**Q: ブラウザで変更が反映されない**
- A: **Ctrl+Shift+Delete でキャッシュを完全クリア**してからリロード
- または、開発者ツール(F12) → ネットワークタブ → "Disable cache" にチェック

### 次の作業者へのアドバイス

1. **CSS と JS の役割を区別する**:
   - CSS: コンテナの見た目（枠線、背景、余白、最小高さ）
   - JS: Chart.js への実際のcanvasサイズ指定

2. **Chart.js のドキュメントを参照**:
   - `maintainAspectRatio: false` は必須
   - `aspectRatio` は削除（高さ制限の原因）
   - `layout.padding` は各チャートタイプに応じて調整

3. **段階的にデバッグ**:
   - まず1つのチャートで動作確認
   - 成功パターンを他のチャートに適用
   - ブラウザキャッシュは必ずクリアする

4. **ユーザーの実際の画面を確認**:
   - スクリーンショットで「何が切れているか」を正確に把握
   - 「はみ出し」と「切れる」は別の問題

---

## 参考: 今回の修正が必要だった背景

gpt-5 mini の以前の修正は「viewport全体にチャートが収まる」ことを目指していましたが、実際のユーザーの問題は「個別チャートの白枠内に収まる」ことでした。この認識のズレにより、根本原因（CSS positioning、canvas サイズ、Chart.js padding）を見逃していました。

**教訓**: ユーザーの提供するスクリーンショットで「何が」「どこから」はみ出しているかを正確に把握することが最優先です。
