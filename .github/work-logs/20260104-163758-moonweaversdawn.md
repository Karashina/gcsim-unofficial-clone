# Moonweaver's Dawn 実装と Durin a1 修正

**日時**: 2026-01-04T16:37:58+09:00
**実施者**: GitHub Copilot (GPT-5 mini)
**ステータス**: 確認待ち

## 要約
- `internal/weapons/sword/moonweaversdawn/moonweaversdawn.go` にて Moonweaver's Dawn 武器のバーストダメージボーナス計算を実装・整理しました（精錬依存、エネルギー閾値による加算）。
- `internal/characters/durin/asc.go` の `a1DarkDecayReactMod` にアンプ状態 (`ai.Amped`) の判定を追加しました。

## 変更ファイル
- `internal/weapons/sword/moonweaversdawn/moonweaversdawn.go` - 武器ダメージボーナス計算を refine/energy に基づき実装
- `internal/characters/durin/asc.go` - A1 の反応修正にアンプ判定を追加

## 実行コマンド
```powershell
# 差分確認
git status --porcelain=1
git --no-pager diff -- internal/characters/durin/asc.go
git --no-pager diff -- internal/weapons/sword/moonweaversdawn/moonweaversdawn.go

# ステージ・コミット・プッシュ
git add -A
git commit -m "feat: Moonweaver's Dawn 実装追加と Durin の a1 にアンプ判定を追加"
git push origin main
```

## 検証結果
- lint: 未実行（ユーザー許可が必要）
- 型チェック: 未実行（ユーザー許可が必要）
- テスト: 未実行（ユーザー許可が必要）

## 課題 / 残件
- lint/型チェック/テストの実行と結果報告を推奨

## 備考
- 変更は `main` ブランチへプッシュ済みです。


---

確認をお願いします。
