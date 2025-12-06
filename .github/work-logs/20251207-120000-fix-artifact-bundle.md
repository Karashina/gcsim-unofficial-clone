 # Fix artifact UI not present in built bundle

 **日時**: 2025-12-07T12:00:00+09:00

 **実施者**: GitHub Copilot (agent)

 **要約**:
 - `webui-src/src/screens/results-characters.js` にコミット済みの「聖遺物をアイコン＋ツールチップで表示する」変更が、配信されるバンドル `webui/app.js` に反映されていなかったため、バンドルのエントリに同等の変更を適用しました。
 - 原因は、ビルド済み `webui/app.js` が `webui-src/src/app.js` の内蔵 `displayCharacters` 実装を使用しており、`webui-src/src/screens/results-characters.js` 側の変更がバンドルに取り込まれていなかったためです。

 **変更ファイル**:
 - `webui-src/src/app.js`: チップ表現からアイコン＋ツールチップ表現へ置換（`setsBadgesHTML` 部分の置換） — 直接バンドルに取り込まれるエントリを修正

 **実行コマンド**:
 - ソース編集（agent による apply_patch）
 - フロントエンドの再ビルド（手動）:
 ```powershell
 Set-Location 'C:\Users\linol\Documents\Gitrepos\gcsim-unofficial-clone\webui-src'
 npm run build
 ```

 **コマンド出力（抜粋）**:
 - エディットはワークスペースに適用済み（agent の apply_patch）。
 - `npm run build` はユーザー環境で実行済み/実行予定。ビルド後に `webui/app.js` に `artifact-icon` が含まれることを確認してください。

 **課題 / 残件**:
 1. ユーザー側で `npm run build` を実行し、`webui/app.js` に `artifact-icon` マークアップが含まれていることを確認してください。
 2. devserver を再起動（`.\scripts\run-local-devserver.ps1`）して、ブラウザでアイコン表示とツールチップが出るか検証してください。

 **TODO 更新**:
 - ID 1: completed
 - ID 2: completed
 - ID 3: completed
 - ID 4: not-started
 - ID 5: not-started

 **備考**:
 - 今回の修正はビルド出力を直接変更するものではなく、ソースエントリの修正です。ビルドを実行していないと配信ファイルには反映されません。
 - ビルド後にも反映されない場合は、どのファイルがバンドルエントリとして使われているか（`webui-src/src/app.js` が正しいエントリであるか）を再確認します。
