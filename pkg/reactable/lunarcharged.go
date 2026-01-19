package reactable

import (
	"fmt"
	"sort"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

// HasLCCloud returns true if LC Cloud is currently active and this is the global LC Cloud holder.
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

// lcDamageRecord stores the actor index, precomputed damage value, and expiry frame.
type lcDamageRecord struct {
	Index  int
	Damage float64
	Expiry int // Frame when this contributor's aura expires
}

// lcDamageResult stores the result of the LC damage calculation.
type lcDamageResult struct {
	FinalDamage    float64
	HighestCR      float64
	HighestCRIndex int
}

// TryAddLC attempts to apply the Lunar Charged (LC) status and handles its logic.
func (r *Reactable) TryAddLC(a *combat.AttackEvent) bool {
	// --- LC Cloud singleton spec ---
	// Only one LC Cloud can exist on the field at a time.
	// If another Reactable tries to add LC, transfer the cloud to it and deactivate the previous one.
	const lcCloudDuration = 360 // 6 seconds = 360 frames
	// If another LC Cloud is active, deactivate it
	if r.core.ActiveLCCloud != nil && r.core.ActiveLCCloud != r {
		if prev, ok := r.core.ActiveLCCloud.(*Reactable); ok {
			prev.lcCloudActive = false
		}
	}
	r.core.ActiveLCCloud = r
	r.lcCloudActive = true
	r.lcCloudExpiry = r.core.F + lcCloudDuration
	r.core.Tasks.Add(func() {
		// Deactivate LC Cloud when expired
		if r.lcCloudActive && r.core.F >= r.lcCloudExpiry {
			r.lcCloudActive = false
			if r.core.ActiveLCCloud == r {
				r.core.ActiveLCCloud = nil
			}
		}
	}, lcCloudDuration)

	// Reset contributor lists when a new reaction is triggered and LC tick is not active
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

		// Aura expiry calculation for contributor removal
		existing := r.Durability[a.Info.Element]
		dur := 0.8 * a.Info.Durability
		if existing > ZeroDur && dur >= existing {
			dr := r.DecayRate[a.Info.Element]
			a.Info.AuraExpiry = r.core.F + int(dur/dr)
		} else {
			a.Info.AuraExpiry = r.core.F + int(6*dur+420)
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

		// Schedule contributor removal at aura expiry
		expiryIndex := actorIdx
		expiryFrame := a.Info.AuraExpiry
		if r.expiryTaskMap == nil {
			r.expiryTaskMap = make(map[int]int)
		}
		r.expiryTaskMap[expiryIndex] = expiryFrame
		r.core.Tasks.Add(func() {
			// Only remove if this is the latest expiry for the contributor
			if r.expiryTaskMap == nil || r.expiryTaskMap[expiryIndex] != expiryFrame {
				return
			}
			// Remove from lcContributor
			newContrib := make([]int, 0, len(r.lcContributor))
			for _, idx := range r.lcContributor {
				if idx != expiryIndex {
					newContrib = append(newContrib, idx)
				}
			}
			r.lcContributor = newContrib

			// Remove from lcPrecalcDamages
			newDamages := make([]lcDamageRecord, 0, len(r.lcPrecalcDamages))
			for _, rec := range r.lcPrecalcDamages {
				if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
					newDamages = append(newDamages, rec)
				}
			}
			r.lcPrecalcDamages = newDamages

			// Remove from lcPrecalcDamagesCRIT
			newDamagesCRIT := make([]lcDamageRecord, 0, len(r.lcPrecalcDamagesCRIT))
			for _, rec := range r.lcPrecalcDamagesCRIT {
				if rec.Index != expiryIndex || rec.Expiry != expiryFrame {
					newDamagesCRIT = append(newDamagesCRIT, rec)
				}
			}
			r.lcPrecalcDamagesCRIT = newDamagesCRIT

			// Remove from expiryTaskMap
			delete(r.expiryTaskMap, expiryIndex)
		}, expiryFrame-r.core.F)

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
	// Set ActorIndex to the character who triggered LC (the attacker).
	atk.ActorIndex = a.Info.ActorIndex
	r.core.Events.Emit(event.OnLunarCharged, r.self, a)

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
		if r.core.Player.ByIndex(a.Info.ActorIndex).StatusIsActive("law-of-new-moon") && r.core.Rand.Float64() < 0.33 {
			r.core.QueueAttackWithSnap(
				atk,
				snap,
				combat.NewSingleTargetHit(r.self.Key()),
				9,
			)
		} // Double LC tick with 33% chance for extra attack if Columbina A4 is active

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

// nextTickLC handles LC periodic damage ticks using the shared damage calculation logic.
func (r *Reactable) nextTickLC(src int, lcActorIndex int) func() {
	return func() {
		// Only tick if LC is still valid, matches the current tick source, and is the global LC Cloud holder.
		if r.lcTickSrc != src || !r.lcCloudActive {
			return
		}
		if r.core.ActiveLCCloud != r {
			return
		}
		// Find the closest enemy to this reactable
		closest := r.core.Combat.ClosestEnemy(r.self.Pos())
		if closest == nil {
			// No valid enemy to target
			r.core.Tasks.Add(r.nextTickLC(src, lcActorIndex), 122+r.core.Rand.Intn(9))
			return
		}
		// Calculate final LC damage using the latest LC state.
		damageResult := r.calcLCDamage(r.lcContributor, r.lcPrecalcDamages, r.lcPrecalcDamagesCRIT)
		atk.FlatDmg = damageResult.FinalDamage

		// Set ActorIndex to the character who triggered LC (the original attacker).
		atk.ActorIndex = lcActorIndex

		var snap combat.Snapshot
		snap.Stats[attributes.CR] = damageResult.HighestCR
		snap.Stats[attributes.CD] = 0 // Prevent additional Crit DMG
		snap.CharLvl = r.core.Player.ByIndex(damageResult.HighestCRIndex).Base.Level
		r.core.QueueAttackWithSnap(
			atk,
			snap,
			combat.NewSingleTargetHit(closest.Key()),
			9,
		)
		if r.core.Player.ByIndex(lcActorIndex).StatusIsActive("law-of-new-moon") && r.core.Rand.Float64() < 0.33 {
			r.core.QueueAttackWithSnap(
				atk,
				snap,
				combat.NewSingleTargetHit(closest.Key()),
				9,
			)
		} // Double LC tick with 33% chance for extra attack if Columbina A4 is active
		// Schedule the next LC tick.
		r.core.Tasks.Add(r.nextTickLC(src, lcActorIndex), 122+r.core.Rand.Intn(9))
	}
}
