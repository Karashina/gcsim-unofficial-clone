# AdjustAllChartInsets: scale multiple charts to fit viewport

**ステータス**: 失敗
**日時**: 2025-12-06T15:18:00+09:00

**実施者**: GitHub Copilot (agent)

**要約**:
- 単一チャートの高さ制御だけでは、複数のチャートの合計高さがページを押し出す問題を解決できないため、`adjustAllChartInsets` を修正しました。
- 全チャートの `required` 高さを最初に収集し、合計がビューポートの保守的な割合（70%）を超える場合は比率で縮小して各チャートの親コンテナに適用するようにしました。
- これによりチャート群が集積してページ全体を押し出す事象を防ぎます。デバッグログを追加して合計値・スケールを出力します。

**変更ファイル**:
- `webui-src/src/app.js`: `adjustAllChartInsets` を書き換え、collect/scale/apply の 2 パス方式に変更。

**実行コマンド**:
```powershell
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src'; npm run build
Set-Location -Path 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone'; .\scripts\run-local-devserver.ps1
```

**コマンド出力（抜粋）**:
- ビルド成功。`webui/app.js` に変更が反映されています。

**課題 / 残件**:
- ユーザー側で再度ページをリロードして挙動を確認してください。コンソールの `[WebUI][Sizing] adjustAllChartInsets totals` 行（`totalRequired`, `allowedTotal`, `scale`）を送っていただけるとより改善が可能です。

**TODO 更新**:
- ID 4: in-progress
- ID 5: completed (this change)

**備考**:
- 将来的にチャートごとの優先順位（例: char-dps は固定で大きめ、その他は縮小）を導入すると見やすさを保ちつつ全体高さを管理できます。
