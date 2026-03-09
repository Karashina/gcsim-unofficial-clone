package reactable

import (
	"sort"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

// tryLunarCrystallize はルナ結晶化反応を処理する（岩 + 水 + LCrs-Key）
// ルナ結晶化は3回のトリガー後に岩元素ダメージを与える3つの月漂いを生成する
func (r *Reactable) tryLunarCrystallize(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}

	// ルナ結晶化の基本係数は0.96
	lcrsAtk := combat.AttackInfo{
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.LunarCrystallize),
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagLCrsDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		IgnoreDefPercent: 1,
	}

	actorIdx := a.Info.ActorIndex
	char := r.core.Player.ByIndex(actorIdx)
	em := char.Stat(attributes.EM)

	// アクティブなLCrsセッションがない場合、貢献者リストをリセット
	if r.lcrsTickSrc == -1 {
		r.lcrsContributor = make([]int, 0, 4)
		r.lcrsPrecalcDamages = make([]lcDamageRecord, 0, 4)
		r.lcrsPrecalcDamagesCRIT = make([]lcDamageRecord, 0, 4)
		r.lcrsExpiryTaskMap = make(map[int]int)
		r.lcrsTriggerCount = 0
		r.lcrsTickSrc = r.core.F
	}

	// 現在のアクターを貢献者として記録
	found := -1
	for i, idx := range r.lcrsContributor {
		if idx == actorIdx {
			found = i
			break
		}
	}

	// オーラ有効期限計算
	existing := r.Durability[Hydro]
	dur := 0.8 * a.Info.Durability
	if existing > ZeroDur && dur >= existing {
		dr := r.DecayRate[Hydro]
		a.Info.AuraExpiry = r.core.F + int(dur/dr)
	} else {
		a.Info.AuraExpiry = r.core.F + int(6*dur+420)
	}

	if found == -1 {
		r.lcrsContributor = append(r.lcrsContributor, actorIdx)
		r.lcrsPrecalcDamages = append(r.lcrsPrecalcDamages, lcDamageRecord{
			Index:  actorIdx,
			Damage: calcLunarCrystallizeDmg(char, lcrsAtk, em),
			Expiry: a.Info.AuraExpiry,
		})
		r.lcrsPrecalcDamagesCRIT = append(r.lcrsPrecalcDamagesCRIT, lcDamageRecord{
			Index:  actorIdx,
			Damage: calcLunarCrystallizeDmgCRIT(char, lcrsAtk, em),
			Expiry: a.Info.AuraExpiry,
		})
	} else {
		r.lcrsPrecalcDamages[found] = lcDamageRecord{
			Index:  actorIdx,
			Damage: calcLunarCrystallizeDmg(char, lcrsAtk, em),
			Expiry: a.Info.AuraExpiry,
		}
		r.lcrsPrecalcDamagesCRIT[found] = lcDamageRecord{
			Index:  actorIdx,
			Damage: calcLunarCrystallizeDmgCRIT(char, lcrsAtk, em),
			Expiry: a.Info.AuraExpiry,
		}
	}

	// オーラ有効期限時に貢献者削除をスケジュール
	expiryIndex := actorIdx
	expiryFrame := a.Info.AuraExpiry
	if r.lcrsExpiryTaskMap == nil {
		r.lcrsExpiryTaskMap = make(map[int]int)
	}
	r.lcrsExpiryTaskMap[expiryIndex] = expiryFrame
	r.core.Tasks.Add(func() {
		if r.lcrsExpiryTaskMap == nil || r.lcrsExpiryTaskMap[expiryIndex] != expiryFrame {
			return
		}
		// lcrsContributorから削除
		newContrib := make([]int, 0, len(r.lcrsContributor))
		for _, idx := range r.lcrsContributor {
			if idx != expiryIndex {
				newContrib = append(newContrib, idx)
			}
		}
		r.lcrsContributor = newContrib

		// lcrsPrecalcDamagesから削除
		newDamages := make([]lcDamageRecord, 0, len(r.lcrsPrecalcDamages))
		for _, rec := range r.lcrsPrecalcDamages {
			if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
				newDamages = append(newDamages, rec)
			}
		}
		r.lcrsPrecalcDamages = newDamages

		// lcrsPrecalcDamagesCRITから削除
		newDamagesCRIT := make([]lcDamageRecord, 0, len(r.lcrsPrecalcDamagesCRIT))
		for _, rec := range r.lcrsPrecalcDamagesCRIT {
			if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
				newDamagesCRIT = append(newDamagesCRIT, rec)
			}
		}
		r.lcrsPrecalcDamagesCRIT = newDamagesCRIT

		delete(r.lcrsExpiryTaskMap, expiryIndex)
	}, expiryFrame-r.core.F)

	// 水元素量を減少
	r.reduce(attributes.Hydro, a.Info.Durability, 0.5)
	a.Info.Durability = 0
	a.Reacted = true

	lcrsAtk.ActorIndex = actorIdx
	r.core.Events.Emit(event.OnLunarCrystallize, r.self, a)

	// トリガーカウントをインクリメント（R-8: トリガーを蓄積）
	r.lcrsTriggerCount++

	// LCrsセッションの有効期限をリフレッシュ
	r.lcrsActiveExpiry = r.core.F + 360
	r.core.Tasks.Add(func() {
		if r.lcrsTickSrc != -1 && r.core.F >= r.lcrsActiveExpiry {
			r.removeLCrs()
		}
	}, 360)

	// R-8: 3回のトリガー後にのみ月漂いを発射
	if r.lcrsTriggerCount >= 3 {
		// R-9: calcLCrsDamage が合算ダメージ + 最高CR貢献者の追跡を処理
		damageResult := r.calcLCrsDamage(
			r.lcrsContributor,
			r.lcrsPrecalcDamages,
			r.lcrsPrecalcDamagesCRIT,
		)
		// 3つの月漂い弾を合算FinalDamageで発射
		r.fireMoondriftHarmony(lcrsAtk, damageResult, actorIdx)
		// 次のサイクル用にトリガーカウントをリセット
		r.lcrsTriggerCount = 0
	}

	return true
}

// calcLunarCrystallizeDmg はルナ結晶化ダメージを計算する（基本係数 0.96）
func calcLunarCrystallizeDmg(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 99 {
		lvl = 99
	}
	if lvl < 0 {
		lvl = 0
	}
	base := 0.96 * reactionLvlBase[lvl] * (1 + char.LCrsBaseReactBonus(atk))
	bonus := (6 * em) / (2000 + em) // R-7修正: 以前は 16*em
	return base * (1 + bonus + char.LCrsReactBonus(atk)) * (1 + char.ElevationBonus(atk))
}

// calcLunarCrystallizeDmgCRIT はクリティカル付きのルナ結晶化ダメージを計算する
func calcLunarCrystallizeDmgCRIT(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 99 {
		lvl = 99
	}
	if lvl < 0 {
		lvl = 0
	}
	base := 0.96 * reactionLvlBase[lvl] * (1 + char.LCrsBaseReactBonus(atk))
	bonus := (6 * em) / (2000 + em) // R-7修正: 以前は 16*em
	cd := char.Stat(attributes.CD)
	return base * (1 + bonus + char.LCrsReactBonus(atk)) * (1 + cd) * (1 + char.ElevationBonus(atk))
}

// calcLCrsDamage は貢献者に基づいて最終ルナ結晶化ダメージを計算する
func (r *Reactable) calcLCrsDamage(
	lcrsContributor []int,
	lcrsPrecalcDamages []lcDamageRecord,
	lcrsPrecalcDamagesCRIT []lcDamageRecord,
) lcDamageResult {
	// 各貢献者のクリティカルをロール
	isCrit := make([]bool, len(lcrsContributor))
	for i, idx := range lcrsContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
	}

	// クリティカル結果に基づいて各貢献者のダメージを選択
	damageList := make([]float64, len(lcrsContributor))
	indexList := make([]int, len(lcrsContributor))
	for i, idx := range lcrsContributor {
		var dmg float64
		if isCrit[i] {
			for _, rec := range lcrsPrecalcDamagesCRIT {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		} else {
			for _, rec := range lcrsPrecalcDamages {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		}
		damageList[i] = dmg
		indexList[i] = idx
	}

	// 最大ダメージの貢献者インデックスを検索
	maxIdx := 0
	for i := 1; i < len(damageList); i++ {
		if damageList[i] > damageList[maxIdx] {
			maxIdx = i
		}
	}
	highestCRIndex := 0
	if len(indexList) > 0 {
		highestCRIndex = indexList[maxIdx]
	}
	highestCR := 0.0
	if len(indexList) > 0 {
		char := r.core.Player.ByIndex(highestCRIndex)
		highestCR = char.Stat(attributes.CR)
	}

	// 係数を適用: 1番目=1.0, 2番目=0.5, 3/4番目=1/12, 5番目以降=0
	type damageWithIndex struct {
		Dmg float64
		Idx int
	}
	dwiList := make([]damageWithIndex, len(damageList))
	for i := range damageList {
		dwiList[i] = damageWithIndex{Dmg: damageList[i], Idx: indexList[i]}
	}
	sort.SliceStable(dwiList, func(i, j int) bool {
		return dwiList[i].Dmg > dwiList[j].Dmg
	})

	finalDamage := 0.0
	for i := 0; i < len(dwiList); i++ {
		var coef float64
		switch i {
		case 0:
			coef = 1.0
		case 1:
			coef = 0.5
		case 2, 3:
			coef = 1.0 / 12.0
		default:
			coef = 0
		}
		finalDamage += dwiList[i].Dmg * coef
	}

	return lcDamageResult{
		FinalDamage:    finalDamage,
		HighestCR:      highestCR,
		HighestCRIndex: highestCRIndex,
	}
}

// fireMoondriftHarmony は月漂いから3つの岩元素ダメージ弾を発射する
func (r *Reactable) fireMoondriftHarmony(lcrsAtk combat.AttackInfo, damageResult lcDamageResult, actorIdx int) {
	// 3回の岩元素ダメージを発射
	// Columbina A4: 3つの月漂い弾全体での単一ロール追加チャージ
	columbinaExtra := false
	columbinaIdx := -1
	for idx, p := range r.core.Player.Chars() {
		if p.Base.Key == keys.Columbina {
			columbinaIdx = idx
			break
		}
	}
	if columbinaIdx != -1 {
		columbinaExtra = r.core.Rand.Float64() < 0.33
	}
	for i := 0; i < 3; i++ {
		delay := 10 + i*5 // ヒットを少しずらす
		idx := i
		r.core.Tasks.Add(func() {
			closest := r.core.Combat.ClosestEnemy(r.self.Pos())
			if closest == nil {
				return
			}
			var snap combat.Snapshot
			char := r.core.Player.ByIndex(actorIdx)
			snap.Stats[attributes.CR] = char.Stat(attributes.CR)
			snap.Stats[attributes.CD] = char.Stat(attributes.CD)
			snap.CharLvl = char.Base.Level
			lcrsAtk.FlatDmg = damageResult.FinalDamage
			// 重複トリガーを避けるため、最初の月漂い弾のみが反応タグを携帯する
			lcrsAtkCopy := lcrsAtk
			if idx == 0 {
				lcrsAtkCopy.AttackTag = attacks.AttackTagLCrsDamage
			} else {
				lcrsAtkCopy.AttackTag = attacks.AttackTagNone
			}
			lcrsAtkCopy.ActorIndex = actorIdx
			r.core.QueueAttackWithSnap(
				lcrsAtkCopy,
				snap,
				combat.NewSingleTargetHit(closest.Key()),
				0,
			)
			if columbinaExtra && idx == 0 && columbinaIdx != -1 {
				colChar := r.core.Player.ByIndex(columbinaIdx)
				atkExtra := lcrsAtkCopy
				atkExtra.ActorIndex = columbinaIdx
				atkExtra.AttackTag = attacks.AttackTagNone
				var snapExtra combat.Snapshot
				snapExtra.Stats[attributes.CR] = colChar.Stat(attributes.CR)
				snapExtra.Stats[attributes.CD] = colChar.Stat(attributes.CD)
				snapExtra.CharLvl = colChar.Base.Level
				r.core.QueueAttackWithSnap(
					atkExtra,
					snapExtra,
					combat.NewSingleTargetHit(closest.Key()),
					0,
				)
			}
		}, delay)
	}

	r.core.Log.NewEvent("moondrift harmony triggered",
		glog.LogElementEvent,
		actorIdx,
	).
		Write("aura", "lcrs").
		Write("target", r.self.Key())
}

// removeLCrs はルナ結晶化ステータスを削除する
func (r *Reactable) removeLCrs() {
	r.lcrsTickSrc = -1
	r.lcrsActiveExpiry = 0
	r.lcrsTriggerCount = 0
	r.core.Log.NewEvent("lcrs expired",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lcrs").
		Write("target", r.self.Key())
}
