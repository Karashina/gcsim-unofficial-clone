package reactable

import (
	"fmt"
	"sort"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/reactions"
)

var atk = combat.AttackInfo{}

// lcDamageRecord stores the actor index and precomputed damage value.
type lcDamageRecord struct {
	Index  int
	Damage float64
}

// lcDamageResult stores the result of the LC damage calculation.
type lcDamageResult struct {
	FinalDamage    float64
	HighestCR      float64
	HighestCRIndex int
}

// TryAddLC attempts to apply the Lunar Charged (LC) status and handles its logic.
func (r *Reactable) TryAddLC(a *combat.AttackEvent) bool {
	// Reset contributor lists when:
	// - A new reaction is triggered and LC tick is not active, or
	// - Hydro or Electro durability is below ZeroDur
	if (!a.Reacted && r.lcTickSrc == -1) ||
		(r.Durability[Hydro] < ZeroDur || r.Durability[Electro] < ZeroDur) {
		r.lcContributor = make([]int, 0, 4)
		r.lcPrecalcDamages = make([]lcDamageRecord, 0, 4)
		r.lcPrecalcDamagesCRIT = make([]lcDamageRecord, 0, 4)
	}

	atk = combat.AttackInfo{
		// ActorIndex will be set after contributors are determined.
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.LunarCharged),
		AttackTag:        attacks.AttackTagLCDamage,
		ICDTag:           attacks.ICDTagLCDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		IgnoreDefPercent: 1,
	}

	// Abort if source durability is insufficient or if frozen durability exists.
	if a.Info.Durability < ZeroDur || r.Durability[Frozen] > ZeroDur {
		return false
	}

	switch a.Info.Element {
	case attributes.Hydro, attributes.Electro:
		actorIdx := a.Info.ActorIndex
		char := r.core.Player.ByIndex(actorIdx)
		em := char.Stat(attributes.EM)

		// On LC creation, also add the last actor of the other element as a contributor.
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

		// Add or update the current actor as a contributor.
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
			})
			r.lcPrecalcDamagesCRIT = append(r.lcPrecalcDamagesCRIT, lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmgCRIT(char, atk, em),
			})
		} else {
			r.lcPrecalcDamages[found] = lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmg(char, atk, em),
			}
			r.lcPrecalcDamagesCRIT[found] = lcDamageRecord{
				Index:  actorIdx,
				Damage: calcLunarChargedDmgCRIT(char, atk, em),
			}
		}

		// Attach or refill Hydro/Electro durability as appropriate.
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
	r.core.Events.Emit(event.OnElectroCharged, r.self, a)

	// Refresh LC expiration time and schedule removal check.
	r.lcActiveExpiry = r.core.F + 360
	r.core.Tasks.Add(func() {
		// Remove LC only if still active and expired.
		if r.lcTickSrc != -1 && r.core.F >= r.lcActiveExpiry {
			r.removeLC()
		}
	}, 360)

	// Calculate final LC damage and get highest CR.
	damageResult := r.calcLCDamage(r.lcContributor, r.lcPrecalcDamages, r.lcPrecalcDamagesCRIT)
	atk.FlatDmg = damageResult.FinalDamage

	// Set ActorIndex to the last added contributor.
	if len(r.lcContributor) > 0 {
		atk.ActorIndex = r.lcContributor[len(r.lcContributor)-1]
	}

	// If LC tick is not active, start it and queue attack with calculated damage.
	if r.lcTickSrc == -1 {
		r.lcTickSrc = r.core.F
		var snap combat.Snapshot
		snap.Stats[attributes.CR] = damageResult.HighestCR
		snap.Stats[attributes.CD] = 0 // Prevent additional Crit DMG
		snap.CharLvl = r.core.Player.ByIndex(damageResult.HighestCRIndex).Base.Level
		r.core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewSingleTargetHit(r.self.Key()),
			9,
		)

		r.core.Tasks.Add(r.nextTickLC(r.core.F), 70)
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

// calcLCDamage calculates the final LC damage based on contributors, crits, and coefficients.
// Returns the CR and index of the character who contributed the highest damage.
func (r *Reactable) calcLCDamage(
	lcContributor []int,
	lcPrecalcDamages []lcDamageRecord,
	lcPrecalcDamagesCRIT []lcDamageRecord,
) lcDamageResult {
	// Roll crit for each contributor.
	isCrit := make([]bool, len(lcContributor))
	for i, idx := range lcContributor {
		char := r.core.Player.ByIndex(idx)
		cr := char.Stat(attributes.CR)
		randVal := r.core.Rand.Float64()
		isCrit[i] = randVal < cr
	}

	// Select each contributor's damage based on crit result.
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

	// Find the index of the contributor with the highest damage.
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

	// Sort damages (and their indices) in descending order.
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

	// Apply coefficients: 1st=1.0, 2nd=0.5, 3rd/4th=1/12, 5th+=0.
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

// wanelc reduces both Electro and Hydro durability for LC wane.
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

// removeLC removes LC status and unsubscribes related events.
func (r *Reactable) removeLC() {
	r.lcTickSrc = -1
	r.lcActiveExpiry = 0
	r.core.Events.Unsubscribe(event.OnEnemyDamage, fmt.Sprintf("lc-%v", r.self.Key()))
	r.core.Log.NewEvent("lc expired",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lc").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[Hydro]).
		Write("electro", r.Durability[Electro])
}

// nextTickLC handles LC periodic damage ticks using the shared damage calculation logic.
func (r *Reactable) nextTickLC(src int) func() {
	return func() {
		// Only tick if LC is still valid and matches the current tick source.
		if r.lcTickSrc != src {
			return
		}
		// Calculate final LC damage using the latest LC state.
		damageResult := r.calcLCDamage(r.lcContributor, r.lcPrecalcDamages, r.lcPrecalcDamagesCRIT)
		atk.FlatDmg = damageResult.FinalDamage

		// Set ActorIndex to the last added contributor.
		if len(r.lcContributor) > 0 {
			atk.ActorIndex = r.lcContributor[len(r.lcContributor)-1]
		}

		var snap combat.Snapshot
		snap.Stats[attributes.CR] = damageResult.HighestCR
		snap.Stats[attributes.CD] = 0 // Prevent additional Crit DMG
		snap.CharLvl = r.core.Player.ByIndex(damageResult.HighestCRIndex).Base.Level
		r.core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewSingleTargetHit(r.self.Key()),
			9,
		)
		// Schedule the next LC tick.
		r.core.Tasks.Add(r.nextTickLC(src), 122+r.core.Rand.Intn(9))
	}
}
