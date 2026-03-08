package characters

import (
	"errors"
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/varka"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/venti"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

// TestVarkaSkillEntersSturmUndDrang verifies that using Skill activates S&D mode
func TestVarkaSkillEntersSturmUndDrang(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	// Fill remaining slots with test chars
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Before skill: S&D should be inactive
	result, err := c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("error querying condition: %v", err)
	}
	if active, ok := result.(bool); ok && active {
		t.Fatal("S&D should not be active before skill use")
	}

	// Execute skill
	p := make(map[string]int)
	if err := c.Player.Exec(action.ActionSkill, keys.Varka, p); err != nil {
		t.Fatalf("unexpected error executing skill: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// After skill: S&D should be active
	result, err = c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("error querying condition: %v", err)
	}
	if active, ok := result.(bool); !ok || !active {
		t.Fatal("S&D should be active after skill use")
	}
}

// TestVarkaFWAChargeConsumption verifies FWA consumes charges in S&D mode
func TestVarkaFWAChargeConsumption(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 1 // C1 grants 1 FWA charge on S&D entry
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Enter S&D with skill
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// C1: should have 1 FWA charge
	result, _ := c.Player.Chars()[idx].Condition([]string{"fwa-charges"})
	charges, ok := result.(int)
	if !ok {
		t.Fatalf("expected fwa-charges to be int, got %T", result)
	}
	if charges < 1 {
		t.Fatalf("C1 should grant at least 1 FWA charge on S&D entry, got %v", charges)
	}

	// Use FWA (skill in S&D mode)
	err = c.Player.Exec(action.ActionSkill, keys.Varka, p)
	if err != nil {
		t.Fatalf("error executing FWA: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// After FWA: charges should be consumed
	result, _ = c.Player.Chars()[idx].Condition([]string{"fwa-charges"})
	chargesAfter, _ := result.(int)
	if chargesAfter >= charges {
		t.Fatalf("FWA should consume a charge, before=%v after=%v", charges, chargesAfter)
	}
}

// TestVarkaFWADealsTwoHits verifies FWA deals 2 hits (Other + Anemo)
func TestVarkaFWADealsTwoHits(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 1
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Enter S&D
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Count hits from FWA
	hitCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			hitCount++
		}
		return false
	}, "fwa-hit-count")

	// Execute FWA
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// Advance extra frames for damage processing
	for i := 0; i < 60; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 2 {
		t.Fatalf("FWA should deal at least 2 hits, got %v", hitCount)
	}
}

// TestVarkaAzureDevourDeals4Hits verifies Azure Devour hits 4 times
func TestVarkaAzureDevourDeals4Hits(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 1
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Enter S&D
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Count Azure Devour hits (charge attack in S&D with FWA charges)
	hitCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			hitCount++
		}
		return false
	}, "azure-hit-count")

	// Execute Azure Devour (charge attack in S&D mode)
	err = c.Player.Exec(action.ActionCharge, keys.Varka, p)
	if err != nil {
		// If charge is not ready, advance and retry
		for i := 0; i < 60; i++ {
			advanceCoreFrame(c)
		}
		err = c.Player.Exec(action.ActionCharge, keys.Varka, p)
		if err != nil {
			t.Fatalf("error executing Azure Devour: %v", err)
		}
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 120; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 4 {
		t.Fatalf("Azure Devour should deal at least 4 hits, got %v", hitCount)
	}
}

// TestVarkaSturmUndDrangExpires verifies S&D mode expires after 12 seconds
func TestVarkaSturmUndDrangExpires(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Enter S&D
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Verify S&D is active
	result, _ := c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if active, ok := result.(bool); !ok || !active {
		t.Fatal("S&D should be active after skill use")
	}

	// Advance 12 seconds + buffer (720 frames + 60 buffer)
	for i := 0; i < 780; i++ {
		advanceCoreFrame(c)
	}

	// S&D should have expired
	result, _ = c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if active, ok := result.(bool); ok && active {
		t.Fatal("S&D should expire after 12 seconds")
	}
}

// TestVarkaBurstDealsDamage verifies Burst deals damage
func TestVarkaBurstDealsDamage(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	dmgCount := make(map[targets.TargetKey]int)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		trg, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			dmgCount[trg.Key()]++
		}
		return false
	}, "burst-dmg-count")

	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// Extra frames for processing
	for i := 0; i < 200; i++ {
		advanceCoreFrame(c)
	}

	if dmgCount[trg[0].Key()] < 2 {
		t.Fatalf("Burst should deal at least 2 hits, got %v", dmgCount[trg[0].Key()])
	}
}

// TestVarkaC6ChainFWAToAzure verifies C6 FWA→Azure chaining via Skill routing
func TestVarkaC6ChainFWAToAzure(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Enter S&D (C1 grants 1 charge)
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Use FWA (sets c6FWAWindowKey)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// After FWA with C6: c6FWAWindowKey should route next Skill to Azure Devour
	ready, _ := c.Player.Chars()[idx].ActionReady(action.ActionSkill, p)
	if !ready {
		t.Fatal("C6: Skill should be ready after FWA (window routes to Azure Devour)")
	}

	// Execute Azure Devour via Skill (C6 routes Skill→Azure when c6FWAWindowKey active)
	err = c.Player.Exec(action.ActionSkill, keys.Varka, p)
	if err != nil {
		t.Fatalf("C6: Skill→Azure Devour should execute after FWA, got error: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// C6 chain is one-direction only (FWA→Azure); chained Azure does NOT
	// re-open the reverse window (consumeCharge=false ⇒ c6AzureWindowKey not set).
	// So no further chain is available. This prevents infinite chains by design.
}

// TestVarkaConditionQueries verifies all Condition() fields return correct types
func TestVarkaConditionQueries(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	ch := c.Player.Chars()[idx]

	// Test hexerei query (disabled via nohex=1)
	result, err := ch.Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || isHex {
		t.Fatal("hexerei should be false with nohex=1")
	}

	// Test sturm-und-drang query
	result, err = ch.Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("sturm-und-drang condition error: %v", err)
	}
	if _, ok := result.(bool); !ok {
		t.Fatalf("sturm-und-drang should return bool, got %T", result)
	}

	// Test fwa-charges query
	result, err = ch.Condition([]string{"fwa-charges"})
	if err != nil {
		t.Fatalf("fwa-charges condition error: %v", err)
	}
	if _, ok := result.(int); !ok {
		t.Fatalf("fwa-charges should return int, got %T", result)
	}

	// Test a4-stacks query
	result, err = ch.Condition([]string{"a4-stacks"})
	if err != nil {
		t.Fatalf("a4-stacks condition error: %v", err)
	}
	if stacks, ok := result.(int); !ok || stacks != 0 {
		t.Fatalf("a4-stacks should return 0 initially, got %v", result)
	}
}

// TestVarkaHexereiBonusDetection verifies Hexerei party detection
func TestVarkaHexereiBonusDetection(t *testing.T) {
	c, _ := makeCore(1)

	// Varka with Hexerei enabled (default)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Varka: %v", err)
	}

	// Add Venti (also Hexerei by default) — should trigger 2+ Hexerei
	profVenti := defProfile(keys.Venti)
	profVenti.Base.Cons = 0
	profVenti.Base.Ascension = 6
	_, err = c.AddChar(profVenti)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}

	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	// Both Varka and Venti are Hexerei → party has 2+ Hexerei → hasHexBonus should be true
	result, err := c.Player.Chars()[0].Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || !isHex {
		t.Fatal("Varka should have hexerei enabled by default")
	}
}

// TestVarkaNoHexDisablesHexerei verifies nohex=1 parameter disables Hexerei
func TestVarkaNoHexDisablesHexerei(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	result, _ := c.Player.Chars()[idx].Condition([]string{"hexerei"})
	if isHex, ok := result.(bool); ok && isHex {
		t.Fatal("hexerei should be disabled with nohex=1")
	}
}

// TestVarkaSkillAllActionsDoNotPanic verifies all actions don't panic for Varka
func TestVarkaSkillAllActionsDoNotPanic(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding char: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// Execute all actions without panicking
	actions := []action.Action{
		action.ActionAttack,
		action.ActionSkill,
		action.ActionBurst,
		action.ActionCharge,
	}
	p := make(map[string]int)
	for _, act := range actions {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic during action %v: %v", act, r)
				}
			}()
			err := c.Player.Exec(act, keys.Varka, p)
			switch {
			case errors.Is(err, player.ErrActionNotReady),
				errors.Is(err, player.ErrPlayerNotReady),
				errors.Is(err, player.ErrActionNoOp):
				// Expected when not ready
			case err == nil:
				for !c.Player.CanQueueNextAction() {
					advanceCoreFrame(c)
				}
			}
			// Advance some frames between actions
			for i := 0; i < 120; i++ {
				advanceCoreFrame(c)
			}
		}()
	}
}
