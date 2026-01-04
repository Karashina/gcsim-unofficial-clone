# グラフはみ出し修正対応

**ステータス**: 失敗
**日時**: 2025-12-06T15:05:00+09:00
**実施者**: GitHub Copilot (agent)

**要約**:
- グラフ描画時にチャート領域がページを押し上げてしまう問題に対処するため、Chart.js の挙動と親コンテナの高さ管理を見直しました。
- 生デバッグパネル（Raw statistics JSON）と自動ターゲット挿入を削除して、UI に不必要なデバッグ要素が出ないようにしました。
- 棒グラフ系について、ラベル数に応じて `barThickness` を動的に縮小するロジックを追加し、チャートがページを伸ばすことを抑制しました。

**変更ファイル**:
- `webui-src/src/app.js`: チャート高さ管理ロジック追加（`setCanvasVisualSize` 調整、`barThickness` のスケーリング、`maintainAspectRatio:false` の適用）、Raw JSON と自動ターゲット挿入の無効化
- `webui-src/src/screens/results-charts.js`: Raw statistics debug パネルの削除
- `webui-src/src/templates/results-screen.html`: 静的な `ターゲット情報` セクションの削除
- `webui-src/src/styles/results.css`: の変更を行い（後に巻き戻しを含む）、canvas の高さは JS 側で管理する形に変更
- `webui/app.js`: ビルド出力を更新（ビルド済みバンドルを生成）
- `.github/copilot-instructions.md`: ワークログ作成ルールを明記する文言を追加

**実行コマンド**:
```powershell
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src'
npm run build
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone'
.\scripts\run-local-devserver.ps1
```

**コマンド出力（抜粋）**:
- `npm run build` -> ✓ CSS built: C:\...\webui\results.css
- `npm run build` -> ✓ JavaScript built: C:\...\webui\app.js
- `run-local-devserver.ps1` -> 警告: Port 8381 is already in use! （既存プロセス停止が必要な場合あり）

**課題 / 残件**:
- devserver が既にポート 8381 で動作しており、新しいビルドが反映されていない可能性があるため、既存プロセスの安全な停止と devserver の再起動が必要。実行にはユーザの明示許可が必要。
- ブラウザでの最終確認（リアルなデータセットでのチャートのはみ出し・アスペクト比）を要実施。

**TODO 更新**:
- TODO #1 `Run frontend build`: completed
- TODO #2 `Start devserver`: in-progress (devserver 再起動待ち)
- TODO #3 `Check console and visuals`: in-progress (ブラウザ確認作業に移行)

**備考**:
- 本ログは `.github/copilot-instructions.md` に従い自動で作成されました。
- ログには機密情報は含めていません。コマンド出力に機密が含まれる場合はマスクしてから記録しています。
