# 新規武器実装: Athame Artis と The Daybreak Chronicles

**日時**: 2025-12-13T21:49:19+09:00

**実施者**: GitHub Copilot (agent)

**要約**:  
weapons.mdの仕様に基づき、2つの新規★5武器を実装しました:
- **Athame Artis** (ID:90630, Sword) - 爆発CRIT DMGバフとパーティATKバフを提供
- **The Daybreak Chronicles** (ID:90631, Bow) - 戦闘外/戦闘中のDMGボーナスシステムを実装

両武器はHexerei: Secret Rite連携機能を持ち、2人以上のHexereiキャラクターがパーティにいる場合、効果が強化されます。

**変更ファイル**:

1. [pkg/core/keys/weapon.go](pkg/core/keys/weapon.go)
   - weaponNames配列に"athameartis"と"thedaybreakchronicles"を追加
   - 武器enum定数にAthameArtisとTheDaybreakChroniclesを追加

2. [internal/weapons/sword/athame/config.yml](internal/weapons/sword/athame/config.yml)
   - 新規作成: 武器メタデータ (ID:90630, key:athameartis)

3. [internal/weapons/sword/athame/data_gen.textproto](internal/weapons/sword/athame/data_gen.textproto)
   - 新規作成: ★5武器ステータス (Base ATK 46→421, CR 7.2%→33.1%, 301カーブ)

4. [internal/weapons/sword/athame/athame.go](internal/weapons/sword/athame/athame.go)
   - 新規作成: Blade of Daylight Hours効果の実装
   - 爆発CD +16-32%, 爆発ヒット時ATK +20-40% (自身) +16-32% (パーティ) 3秒
   - Hexerei連携で75%効果増強

5. [internal/weapons/sword/athame/athame_gen.go](internal/weapons/sword/athame/athame_gen.go)
   - 新規作成: Data()メソッドとprotoデータ読み込み処理

6. [internal/weapons/bow/daybreak/config.yml](internal/weapons/bow/daybreak/config.yml)
   - 新規作成: 武器メタデータ (ID:90631, key:thedaybreakchronicles)

7. [internal/weapons/bow/daybreak/data_gen.textproto](internal/weapons/bow/daybreak/data_gen.textproto)
   - 新規作成: ★5武器ステータス (Base ATK 48→488, CD 9.6%→44.1%, 302カーブ)

8. [internal/weapons/bow/daybreak/daybreak.go](internal/weapons/bow/daybreak/daybreak.go)
   - 新規作成: Stirring Dawn Breeze効果の実装
   - 戦闘外3秒後に60-120% DMGボーナス蓄積、戦闘中10-20%/秒減衰
   - ヒット時対応DMGタイプに10-20%ボーナス (最大60-120%)
   - Hexereiモードで全DMGタイプに20-40%ボーナス

9. [internal/weapons/bow/daybreak/daybreak_gen.go](internal/weapons/bow/daybreak/daybreak_gen.go)
   - 新規作成: Data()メソッドとprotoデータ読み込み処理

**実行コマンド**:

```powershell
# 武器ビルド検証
go build ./internal/weapons/sword/athame
go build ./internal/weapons/bow/daybreak

# gcsimメインバイナリビルド検証
go build ./cmd/gcsim
```

**コマンド出力（抜粋）**:

すべてのビルドコマンドが成功しました (エラー出力なし)。

**課題 / 残件**:

1. ✅ TheDaybreakChronicles enum定数追加 (完了)
2. ✅ *_gen.goファイル生成 (手動作成完了)
3. ✅ ビルド検証 (完了)
4. ⚠️ 実機テスト未実施 - Hexerei連携とステータス効果の動作確認が必要
5. ⚠️ The Daybreak Chroniclesの戦闘外/戦闘中判定がcombat.goのロジックに依存 - 実際の挙動検証が必要

**TODO 更新**:

manage_todo_listは使用していませんが、主要な実装タスクは完了しています。

**備考**:

- 両武器は★5ステータスパターン (WEAPON_PROTO_DATA_GUIDE.md参照) に準拠
- Athame: ATTACK_301カーブ (Lv90 x2.061), CR substat
- Daybreak: ATTACK_302カーブ (Lv90 x2.465), CD substat
- Hexereiステータスチェック: `char.StatusIsActive("hexerei-character")`
- Athame効果はOnEnemyHitイベントで処理、3秒のバフ期間
- Daybreak効果は180fタスク (戦闘外蓄積) と60fタスク (戦闘中減衰) で処理
- 両武器とも既存の武器実装パターン (alley, mistsplitter等) を踏襲

**参考リンク**:

- [WEAPON_PROTO_DATA_GUIDE.md](docs/WEAPON_PROTO_DATA_GUIDE.md) - 武器データ仕様書
- [weapons.md](weapons.md) - 武器仕様原文
