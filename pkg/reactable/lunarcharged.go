package reactable

import (
	"fmt"
	"sort"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

// HasLCCloud はLC雲が現在アクティブで、かつこれがグローバルなLC雲保持者である場合にtrueを返す。
func (r *Reactable) HasLCCloud() bool {
	if r.core.ActiveLCCloud == nil {
		return false
	}
	if cloud, ok := r.core.ActiveLCCloud.(*Reactable); ok {
		return r.lcCloudActive && r.core.F < r.lcCloudExpiry && cloud == r
	}
	return false
}

var atk = combat.AttackInfo{}

// lcDamageRecord はアクターインデックス、事前計算ダメージ値、および有効期限フレームを格納する。
type lcDamageRecord struct {
	Index  int
	Damage float64
	Expiry int // Frame when this contributor's aura expires
}

// lcDamageResult はLCダメージ計算の結果を格納する。
type lcDamageResult struct {
	FinalDamage    float64
	HighestCR      float64
	HighestCRIndex int
}

// TryAddLC はルナチャージ（LC）ステータスの適用を試み、そのロジックを処理する。
func (r *Reactable) TryAddLC(a *combat.AttackEvent) bool {
	// --- LC雲シングルトン仕様 ---
	// フィールド上に存在できるLC雲は1つだけ。
	// 別のReactableがLCを追加しようとした場合、雲を移動して前のものを無効化する。
	const lcCloudDuration = 360 // 6 seconds = 360 frames
	// 別のLC雲がアクティブなら無効化する
	if r.core.ActiveLCCloud != nil && r.core.ActiveLCCloud != r {
		if prev, ok := r.core.ActiveLCCloud.(*Reactable); ok {
			prev.lcCloudActive = false
		}
	}
	r.core.ActiveLCCloud = r
	r.lcCloudActive = true
	r.lcCloudExpiry = r.core.F + lcCloudDuration
	r.core.Tasks.Add(func() {
		// 期限切れ時にLC雲を無効化
		if r.lcCloudActive && r.core.F >= r.lcCloudExpiry {
			r.lcCloudActive = false
			if r.core.ActiveLCCloud == r {
				r.core.ActiveLCCloud = nil
			}
		}
	}, lcCloudDuration)

	// 新しい反応が発動され、LCティックがアクティブでない場合、貢献者リストをリセット
	if !a.Reacted && r.lcTickSrc == -1 {
		r.lcContributor = make([]int, 0, 4)
		r.lcPrecalcDamages = make([]lcDamageRecord, 0, 4)
		r.lcPrecalcDamagesCRIT = make([]lcDamageRecord, 0, 4)
		r.expiryTaskMap = make(map[int]int)
	}

	atk = combat.AttackInfo{
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.LunarCharged),
		AttackTag:        attacks.AttackTagLCDamage,
		ICDTag:           attacks.ICDTagLCDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		IgnoreDefPercent: 1,
	}

	// ソース元素量が不足、または凍結元素量が存在する場合は中止。
	if a.Info.Durability < ZeroDur || r.Durability[Frozen] > ZeroDur {
		return false
	}

	switch a.Info.Element {
	case attributes.Hydro, attributes.Electro:
		actorIdx := a.Info.ActorIndex
		char := r.core.Player.ByIndex(actorIdx)
		em := char.Stat(attributes.EM)

		// LC作成時、他の元素の最後のアクターも貢献者として追加する。
		if !a.Reacted && r.lcTickSrc == -1 {
			var otherElement attributes.Element
			if a.Info.Element == attributes.Hydro {
				otherElement = attributes.Electro
			} else {
				otherElement = attributes.Hydro
			}
			otherIdx := r.lastEleSource[otherElement]
			if otherIdx != actorIdx && otherIdx >= 0 {
				otherChar := r.core.Player.ByIndex(otherIdx)
				otherEM := otherChar.Stat(attributes.EM)
				found := false
				for _, idx := range r.lcContributor {
					if idx == otherIdx {
						found = true
						break
					}
				}
				if !found {
					r.lcContributor = append(r.lcContributor, otherIdx)
					r.lcPrecalcDamages = append(r.lcPrecalcDamages, lcDamageRecord{
						Index:  otherIdx,
						Damage: calcLunarChargedDmg(otherChar, atk, otherEM),
					})
					r.lcPrecalcDamagesCRIT = append(r.lcPrecalcDamagesCRIT, lcDamageRecord{
						Index:  otherIdx,
						Damage: calcLunarChargedDmgCRIT(otherChar, atk, otherEM),
					})
				}
			}
		}

		// 貢献者削除用のオーラ有効期限計算
		existing := r.Durability[a.Info.Element]
		dur := 0.8 * a.Info.Durability
		if existing > ZeroDur && dur >= existing {
			dr := r.DecayRate[a.Info.Element]
			a.Info.AuraExpiry = r.core.F + int(dur/dr)
		} else {
			a.Info.AuraExpiry = r.core.F + int(6*dur+420)
		}

		// 現在のアクターを貢献者として追加または更新。
		found := -1
		for i, idx := range r.lcContributor {
			if idx == actorIdx {
				found = i
				break
			}
		}
		if found == -1 {
			r.lcContributor = append(r.lcContributor, actorIdx)
			r.lcPrecalcDamages = append(r.lcPrecalcDamages, lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmg(char, atk, em),
				Expiry: a.Info.AuraExpiry,
			})
			r.lcPrecalcDamagesCRIT = append(r.lcPrecalcDamagesCRIT, lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmgCRIT(char, atk, em),
				Expiry: a.Info.AuraExpiry,
			})
		} else {
			r.lcPrecalcDamages[found] = lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmg(char, atk, em),
				Expiry: a.Info.AuraExpiry,
			}
			r.lcPrecalcDamagesCRIT[found] = lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmgCRIT(char, atk, em),
				Expiry: a.Info.AuraExpiry,
			}
		}

		// オーラ有効期限時に貢献者削除をスケジュール
		expiryIndex := actorIdx
		expiryFrame := a.Info.AuraExpiry
		if r.expiryTaskMap == nil {
			r.expiryTaskMap = make(map[int]int)
		}
		r.expiryTaskMap[expiryIndex] = expiryFrame
		r.core.Tasks.Add(func() {
			// この貢献者の最新の有効期限の場合のみ削除
			if r.expiryTaskMap == nil || r.expiryTaskMap[expiryIndex] != expiryFrame {
				return
			}
			// lcContributorから削除
			newContrib := make([]int, 0, len(r.lcContributor))
			for _, idx := range r.lcContributor {
				if idx != expiryIndex {
					newContrib = append(newContrib, idx)
				}
			}
			r.lcContributor = newContrib

			// lcPrecalcDamagesから削除
			newDamages := make([]lcDamageRecord, 0, len(r.lcPrecalcDamages))
			for _, rec := range r.lcPrecalcDamages {
				if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
					newDamages = append(newDamages, rec)
				}
			}
			r.lcPrecalcDamages = newDamages

			// lcPrecalcDamagesCRITから削除
			newDamagesCRIT := make([]lcDamageRecord, 0, len(r.lcPrecalcDamagesCRIT))
			for _, rec := range r.lcPrecalcDamagesCRIT {
				if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
					newDamagesCRIT = append(newDamagesCRIT, rec)
				}
			}
			r.lcPrecalcDamagesCRIT = newDamagesCRIT

			// expiryTaskMapから削除
			delete(r.expiryTaskMap, expiryIndex)
		}, expiryFrame-r.core.F)

		// 水/雷の元素量を適切に付着または補充。
		if a.Info.Element == attributes.Hydro {
			if r.Durability[Electro] < ZeroDur {
				return false
			}
			if !a.Reacted {
				r.attachOrRefillNormalEle(Hydro, a.Info.Durability)
			}
		} else { // Electro
			if r.Durability[Hydro] < ZeroDur {
				return false
			}
			if !a.Reacted {
				r.attachOrRefillNormalEle(Electro, a.Info.Durability)
			}
		}
	default:
		return false
	}

	a.Reacted = true
	// ActorIndexをLCを発動したキャラクター（攻撃者）に設定。
	atk.ActorIndex = a.Info.ActorIndex
	r.core.Events.Emit(event.OnLunarCharged, r.self, a)

	// LCの有効期限をリフレッシュし、削除チェックをスケジュール。
	r.lcActiveExpiry = r.core.F + 360
	r.core.Tasks.Add(func() {
		// まだアクティブで期限切れの場合のみLCを削除。
		if r.lcTickSrc != -1 && r.core.F >= r.lcActiveExpiry {
			r.removeLC()
		}
	}, 360)

	// 各貢献者の個別LCダメージを計算し、攻撃をキュー
	damageList, indexList, isCrit := r.calcLCDamageContrib(r.lcContributor, r.lcPrecalcDamages, r.lcPrecalcDamagesCRIT)
	// 係数割り当てのためにダメージ（とそのインデックス）を降順にソート
	type damageWithIndex struct {
		Dmg  float64
		Idx  int
		Crit bool
	}
	dwiList := make([]damageWithIndex, len(damageList))
	for i := range damageList {
		dwiList[i] = damageWithIndex{Dmg: damageList[i], Idx: indexList[i], Crit: isCrit[i]}
	}
	sort.SliceStable(dwiList, func(i, j int) bool {
		return dwiList[i].Dmg > dwiList[j].Dmg
	})
	// 係数: 1番目=1.0, 2番目=0.5, 3/4番目=1/12, 5番目以降=0
	coefs := []float64{1.0, 0.5, 1.0 / 12.0, 1.0 / 12.0}
	if r.lcTickSrc == -1 {
		r.lcTickSrc = r.core.F
		// Columbina A4: LC作成時の追加攻撃の単一ロール
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
		for i, dwi := range dwiList {
			if i >= len(coefs) {
				break
			}
			coef := coefs[i]
			if coef == 0 {
				continue
			}
			var snap combat.Snapshot
			char := r.core.Player.ByIndex(dwi.Idx)
			snap.Stats[attributes.CR] = char.Stat(attributes.CR)
			snap.Stats[attributes.CD] = 0 // 追加の会心ダメージを防止
			snap.CharLvl = char.Base.Level
			atkCopy := atk
			atkCopy.ActorIndex = dwi.Idx
			atkCopy.FlatDmg = dwi.Dmg * coef
			if i == 0 {
				atkCopy.AttackTag = attacks.AttackTagLCDamage
			} else {
				atkCopy.AttackTag = attacks.AttackTagNone
			}
			r.core.QueueAttackWithSnap(
				atkCopy,
				snap,
				combat.NewSingleTargetHit(r.self.Key()),
				9,
			)
			// Columbina A4: ロール成功時、最大貢献者のみに追加攻撃を適用
			if columbinaExtra && i == 0 && columbinaIdx != -1 {
				colChar := r.core.Player.ByIndex(columbinaIdx)
				atkExtra := atkCopy
				atkExtra.ActorIndex = columbinaIdx
				atkExtra.AttackTag = attacks.AttackTagNone
				var snapExtra combat.Snapshot
				snapExtra.Stats[attributes.CR] = colChar.Stat(attributes.CR)
				snapExtra.Stats[attributes.CD] = 0
				snapExtra.CharLvl = colChar.Base.Level
				r.core.QueueAttackWithSnap(
					atkExtra,
					snapExtra,
					combat.NewSingleTargetHit(r.self.Key()),
					9,
				)
			}
		}
		r.core.Tasks.Add(r.nextTickLC(r.core.F, a.Info.ActorIndex), 70)
		r.core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			n := args[0].(combat.Target)
			ae := args[1].(*combat.AttackEvent)
			dmg := args[2].(float64)
			if n.Key() != r.self.Key() {
				return false
			}
			if ae.Info.AttackTag != attacks.AttackTagLCDamage {
				return false
			}
			if dmg == 0 {
				return false
			}
			if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
				return true
			}
			r.core.Tasks.Add(func() {
				r.wanelc()
			}, 6)
			return false
		}, fmt.Sprintf("lc-%v", r.self.Key()))
	}
	return true
}

// calcLCDamageContrib: contributorごとのダメージ・index・クリティカル判定を返す
func (r *Reactable) calcLCDamageContrib(lcContributor []int, lcPrecalcDamages []lcDamageRecord, lcPrecalcDamagesCRIT []lcDamageRecord) (damageList []float64, indexList []int, isCrit []bool) {
	isCrit = make([]bool, len(lcContributor))
	damageList = make([]float64, len(lcContributor))
	indexList = make([]int, len(lcContributor))
	for i, idx := range lcContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
		var dmg float64
		if isCrit[i] {
			for _, rec := range lcPrecalcDamagesCRIT {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		} else {
			for _, rec := range lcPrecalcDamages {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		}
		damageList[i] = dmg
		indexList[i] = idx
	}
	return
}

// calcLCDamage は貢献者、クリティカル、係数に基づいて最終LCダメージを計算する。
// 最大ダメージを貢献したキャラクターのCRとインデックスを返す。
func (r *Reactable) calcLCDamage(
	lcContributor []int,
	lcPrecalcDamages []lcDamageRecord,
	lcPrecalcDamagesCRIT []lcDamageRecord,
) lcDamageResult {
	// 各貢献者のクリティカルをロール。
	isCrit := make([]bool, len(lcContributor))
	for i, idx := range lcContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
	}

	// クリティカル結果に基づいて各貢献者のダメージを選択。
	damageList := make([]float64, len(lcContributor))
	indexList := make([]int, len(lcContributor))
	for i, idx := range lcContributor {
		var dmg float64
		if isCrit[i] {
			for _, rec := range lcPrecalcDamagesCRIT {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		} else {
			for _, rec := range lcPrecalcDamages {
				if rec.Index == idx {
					dmg = rec.Damage
					break
				}
			}
		}
		damageList[i] = dmg
		indexList[i] = idx
	}

	// 最大ダメージの貢献者インデックスを検索。
	maxIdx := 0
	for i := 1; i < len(damageList); i++ {
		if damageList[i] > damageList[maxIdx] {
			maxIdx = i
		}
	}
	highestCRIndex := indexList[maxIdx]
	highestCR := 0.0
	if len(indexList) > 0 {
		char := r.core.Player.ByIndex(highestCRIndex)
		highestCR = char.Stat(attributes.CR)
	}

	// ダメージ（とそのインデックス）を降順にソート。
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

	// 係数を適用: 1番目=1.0, 2番目=0.5, 3/4番目=1/12, 5番目以降=0。
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

// wanelc はLC減衰用に雷と水の元素量を減少させる。
func (r *Reactable) wanelc() {
	r.Durability[Electro] -= 10
	r.Durability[Electro] = max(0, r.Durability[Electro])
	r.Durability[Hydro] -= 10
	r.Durability[Hydro] = max(0, r.Durability[Hydro])
	r.core.Log.NewEvent("lc wane",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lc").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[Hydro]).
		Write("electro", r.Durability[Electro])
}

// removeLC はLCステータスを削除し、関連イベントの購読を解除する。
func (r *Reactable) removeLC() {
	r.lcTickSrc = -1
	r.lcActiveExpiry = 0
	r.core.Events.Unsubscribe(event.OnEnemyDamage, fmt.Sprintf("lc-%v", r.self.Key()))
	r.core.Events.Unsubscribe(event.OnTick, fmt.Sprintf("lc-tick-%v", r.self.Key()))
	r.core.Log.NewEvent("lc expired",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lc").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[Hydro]).
		Write("electro", r.Durability[Electro])
}

// nextTickLC は共有ダメージ計算ロジックを使用してLC定期ダメージティックを処理する。
func (r *Reactable) nextTickLC(src int, lcActorIndex int) func() {
	return func() {
		// LCがまだ有効で、現在のティックソースと一致し、グローバルLC雲保持者である場合のみティックする。
		if r.lcTickSrc != src || !r.lcCloudActive {
			return
		}
		if r.core.ActiveLCCloud != r {
			return
		}
		// このReactableに最も近い敵を検索
		closest := r.core.Combat.ClosestEnemy(r.self.Pos())
		if closest == nil {
			// ターゲット可能な敵がいない
			r.core.Tasks.Add(r.nextTickLC(src, lcActorIndex), 122+r.core.Rand.Intn(9))
			return
		}
		// 各貢献者の個別LCダメージを計算し、攻撃をキュー（ティック）
		damageList, indexList, isCrit := r.calcLCDamageContrib(r.lcContributor, r.lcPrecalcDamages, r.lcPrecalcDamagesCRIT)
		type damageWithIndex struct {
			Dmg  float64
			Idx  int
			Crit bool
		}
		dwiList := make([]damageWithIndex, len(damageList))
		for i := range damageList {
			dwiList[i] = damageWithIndex{Dmg: damageList[i], Idx: indexList[i], Crit: isCrit[i]}
		}
		sort.SliceStable(dwiList, func(i, j int) bool {
			return dwiList[i].Dmg > dwiList[j].Dmg
		})
		coefs := []float64{1.0, 0.5, 1.0 / 12.0, 1.0 / 12.0}
		// Columbina A4: LCティック時の追加攻撃の単一ロール
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
		for i, dwi := range dwiList {
			if i >= len(coefs) {
				break
			}
			coef := coefs[i]
			if coef == 0 {
				continue
			}
			var snap combat.Snapshot
			char := r.core.Player.ByIndex(dwi.Idx)
			snap.Stats[attributes.CR] = char.Stat(attributes.CR)
			snap.Stats[attributes.CD] = 0
			snap.CharLvl = char.Base.Level
			atkCopy := atk
			atkCopy.ActorIndex = dwi.Idx
			atkCopy.FlatDmg = dwi.Dmg * coef
			if i == 0 {
				atkCopy.AttackTag = attacks.AttackTagLCDamage
			} else {
				atkCopy.AttackTag = attacks.AttackTagNone
			}
			r.core.QueueAttackWithSnap(
				atkCopy,
				snap,
				combat.NewSingleTargetHit(closest.Key()),
				9,
			)
			if columbinaExtra && i == 0 && columbinaIdx != -1 {
				colChar := r.core.Player.ByIndex(columbinaIdx)
				atkExtra := atkCopy
				atkExtra.ActorIndex = columbinaIdx
				atkExtra.AttackTag = attacks.AttackTagNone
				var snapExtra combat.Snapshot
				snapExtra.Stats[attributes.CR] = colChar.Stat(attributes.CR)
				snapExtra.Stats[attributes.CD] = 0
				snapExtra.CharLvl = colChar.Base.Level
				r.core.QueueAttackWithSnap(
					atkExtra,
					snapExtra,
					combat.NewSingleTargetHit(closest.Key()),
					9,
				)
			}
		}
		// 次のLCティックをスケジュール。
		r.core.Tasks.Add(r.nextTickLC(src, lcActorIndex), 122+r.core.Rand.Intn(9))
	}
}
