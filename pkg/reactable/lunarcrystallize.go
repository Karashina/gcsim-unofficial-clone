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

// tryLunarCrystallize handles Lunar-Crystallize reaction (Geo + Hydro with LCrs-Key)
// Lunar-Crystallize creates 3 Moondrifts that deal Geo DMG after 3 triggers
func (r *Reactable) tryLunarCrystallize(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}

	// Lunar-Crystallize base coefficient is 0.96 (not 1.8 like Lunar-Charged)
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

	// Reset contributor lists when a new reaction is triggered and LCrs tick is not active
	if !a.Reacted && r.lcrsTickSrc == -1 {
		r.lcrsContributor = make([]int, 0, 4)
		r.lcrsPrecalcDamages = make([]lcDamageRecord, 0, 4)
		r.lcrsPrecalcDamagesCRIT = make([]lcDamageRecord, 0, 4)
		r.lcrsExpiryTaskMap = make(map[int]int)
		r.lcrsTriggerCount = 0
	}

	// Record the current actor as a contributor
	found := -1
	for i, idx := range r.lcrsContributor {
		if idx == actorIdx {
			found = i
			break
		}
	}

	// Aura expiry calculation
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

	// Schedule contributor removal at aura expiry
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
		// Remove from lcrsContributor
		newContrib := make([]int, 0, len(r.lcrsContributor))
		for _, idx := range r.lcrsContributor {
			if idx != expiryIndex {
				newContrib = append(newContrib, idx)
			}
		}
		r.lcrsContributor = newContrib

		// Remove from lcrsPrecalcDamages
		newDamages := make([]lcDamageRecord, 0, len(r.lcrsPrecalcDamages))
		for _, rec := range r.lcrsPrecalcDamages {
			if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
				newDamages = append(newDamages, rec)
			}
		}
		r.lcrsPrecalcDamages = newDamages

		// Remove from lcrsPrecalcDamagesCRIT
		newDamagesCRIT := make([]lcDamageRecord, 0, len(r.lcrsPrecalcDamagesCRIT))
		for _, rec := range r.lcrsPrecalcDamagesCRIT {
			if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
				newDamagesCRIT = append(newDamagesCRIT, rec)
			}
		}
		r.lcrsPrecalcDamagesCRIT = newDamagesCRIT

		delete(r.lcrsExpiryTaskMap, expiryIndex)
	}, expiryFrame-r.core.F)

	// Reduce Hydro durability
	r.reduce(attributes.Hydro, a.Info.Durability, 0.5)
	a.Info.Durability = 0
	a.Reacted = true

	lcrsAtk.ActorIndex = actorIdx
	r.core.Events.Emit(event.OnLunarCrystallize, r.self, a)

	// Increment trigger count
	r.lcrsTriggerCount++

	// Refresh LCrs expiration time
	r.lcrsActiveExpiry = r.core.F + 360
	r.core.Tasks.Add(func() {
		if r.lcrsTickSrc != -1 && r.core.F >= r.lcrsActiveExpiry {
			r.removeLCrs()
		}
	}, 360)

	// Calculate individual LCrs damages and queue attacks per contributor
	damageList, indexList, isCrit := r.calcLCrsDamageContrib(r.lcrsContributor, r.lcrsPrecalcDamages, r.lcrsPrecalcDamagesCRIT)
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
	if r.lcrsTickSrc == -1 {
		r.lcrsTickSrc = r.core.F
		// Columbina A4: single-roll for extra attack on LCrs creation
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
			snap.Stats[attributes.CD] = 0 // Prevent additional Crit DMG
			snap.CharLvl = char.Base.Level
			atkCopy := lcrsAtk
			atkCopy.ActorIndex = dwi.Idx
			atkCopy.FlatDmg = dwi.Dmg * coef
			if i == 0 {
				atkCopy.AttackTag = attacks.AttackTagLCrsDamage
			} else {
				atkCopy.AttackTag = attacks.AttackTagNone
			}
			r.core.QueueAttackWithSnap(
				atkCopy,
				snap,
				combat.NewSingleTargetHit(r.self.Key()),
				9,
			)
			if columbinaExtra && i == 0 {
				if columbinaIdx != -1 {
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
		}
	}
	return true
}

// calcLCrsDamageContrib: contributorごとのダメージ・index・クリティカル判定を返す
func (r *Reactable) calcLCrsDamageContrib(lcrsContributor []int, lcrsPrecalcDamages []lcDamageRecord, lcrsPrecalcDamagesCRIT []lcDamageRecord) (damageList []float64, indexList []int, isCrit []bool) {
	isCrit = make([]bool, len(lcrsContributor))
	damageList = make([]float64, len(lcrsContributor))
	indexList = make([]int, len(lcrsContributor))
	for i, idx := range lcrsContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
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
	return
}

// calcLunarCrystallizeDmg calculates Lunar-Crystallize damage (base coefficient 0.96)
func calcLunarCrystallizeDmg(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 99 {
		lvl = 99
	}
	if lvl < 0 {
		lvl = 0
	}
	base := 0.96 * reactionLvlBase[lvl] * (1 + char.LCrsBaseReactBonus(atk))
	bonus := (16 * em) / (2000 + em)
	return base * (1 + bonus + char.LCrsReactBonus(atk)) * (1 + char.ElevationBonus(atk))
}

// calcLunarCrystallizeDmgCRIT calculates Lunar-Crystallize damage with crit
func calcLunarCrystallizeDmgCRIT(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 99 {
		lvl = 99
	}
	if lvl < 0 {
		lvl = 0
	}
	base := 0.96 * reactionLvlBase[lvl] * (1 + char.LCrsBaseReactBonus(atk))
	bonus := (16 * em) / (2000 + em)
	cd := char.Stat(attributes.CD)
	return base * (1 + bonus + char.LCrsReactBonus(atk)) * (1 + cd) * (1 + char.ElevationBonus(atk))
}

// calcLCrsDamage calculates the final Lunar-Crystallize damage based on contributors
func (r *Reactable) calcLCrsDamage(
	lcrsContributor []int,
	lcrsPrecalcDamages []lcDamageRecord,
	lcrsPrecalcDamagesCRIT []lcDamageRecord,
) lcDamageResult {
	// Roll crit for each contributor
	isCrit := make([]bool, len(lcrsContributor))
	for i, idx := range lcrsContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
	}

	// Select each contributor's damage based on crit result
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

	// Find the index of the contributor with the highest damage
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

	// Apply coefficients: 1st=1.0, 2nd=0.5, 3rd/4th=1/12, 5th+=0
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

// fireMoondriftHarmony fires 3 projectiles from Moondrifts dealing Geo DMG
func (r *Reactable) fireMoondriftHarmony(lcrsAtk combat.AttackInfo, damageResult lcDamageResult, actorIdx int) {
	// Fire 3 instances of Geo DMG
	// Columbina A4: single-roll extra charge across the 3 moondrift projectiles
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
		delay := 10 + i*5 // Stagger the hits slightly
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
			// Only the first moondrift projectile should carry the reaction tag to avoid duplicate triggers
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

// removeLCrs removes Lunar-Crystallize status
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
