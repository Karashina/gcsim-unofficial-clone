# Fix chart overflow: cap source/stacked DPS chart heights

**ステータス**: 失敗
**日時**: 2025-12-06T15:10:00+09:00
**実施者**: GitHub Copilot (agent)

**要約**:
- ブラウザコンソールの診断ログから、`source-dps-chart` / 能力別スタックチャートが行数に応じて高さを大きく取りすぎ、ページを押し出していることを確認。
- `webui-src/src/app.js` 内のチャート高さ計算（`createStackedAbilitiesChart` と `createBarChart`）を修正し、ビューポートに収まるよう上限（おおむねビューポートの75%）を設定。行数が多い場合は barThickness を動的に算出して収めるように変更した。
- 変更をビルドに反映させる準備を行い、続く検証としてフロントエンド再ビルドとブラウザでの確認を依頼。

**変更ファイル**:
- `webui-src/src/app.js`: チャートの高さ計算ロジックを修正（`createStackedAbilitiesChart` と `createBarChart` の desiredHeight 算出を viewport-aware に変更）。

**実行コマンド**:
```powershell
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src'; npm run build
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone'; .\scripts\run-local-devserver.ps1
``` 

**コマンド出力（抜粋）**:
- フロントエンドビルドは成功し `webui/app.js` と `webui/results.css` が更新されました。
- ローカル devserver を起動済み（必要な場合は既存のプロセスを停止して再起動しました）。

**課題 / 残件**:
- ユーザー側でブラウザをリロードして、該当ページ（結果画面）でチャートが正しく収まるか確認してください。
- もし依然としてチャートがページを押し出す場合は、ブラウザコンソールから `[WebUI][Sizing]` プレフィックスの付いた診断ログをコピーして提供してください（該当ログはチャートごとの desiredHeight/cappedHeight/parentInlineHeight 等を示します）。

**TODO 更新**:
- ID 1: completed
- ID 2: completed
- ID 3: completed
- ID 4: in-progress

**備考**:
- 変更は「高さを無制限に伸ばす」のではなく「表示可能領域内に収める」方針です。大量の行がある場合は、追加オプション（モーダルで全件表示、ページネーション、折りたたみ）を提案できます。
