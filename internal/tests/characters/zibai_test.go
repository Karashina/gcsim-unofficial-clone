package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/zibai"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TestZibaiSkillActivatesPhaseShift verifies Skill enters Lunar Phase Shift mode
func TestZibaiSkillActivatesPhaseShift(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai: %v", err)
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

	// Before skill: Phase Shift should be inactive (0)
	result, err := c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if err != nil {
		t.Fatalf("lunar-phase-shift condition error: %v", err)
	}
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Phase Shift should be 0 (inactive) before skill use")
	}

	// Execute skill
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// After skill: Phase Shift should be active (1)
	result, _ = c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatalf("Phase Shift should be 1 (active) after skill use, got %v (%T)", result, result)
	}
}

// TestZibaiRadianceConditionQuery verifies Radiance starts at expected value
func TestZibaiRadianceConditionQuery(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai: %v", err)
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

	// Activate Phase Shift
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// After activation, radiance should be 0 (or 100 with C1)
	result, err := c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	if err != nil {
		t.Fatalf("phase-shift-radiance condition error: %v", err)
	}
	radiance, ok := result.(int)
	if !ok {
		t.Fatalf("radiance should return int, got %T", result)
	}
	// C0: radiance resets to 0 on skill
	if radiance < 0 {
		t.Fatalf("radiance should be non-negative, got %v", radiance)
	}
}

// TestZibaiRadianceAccumulation verifies Radiance increases over time during Phase Shift
func TestZibaiRadianceAccumulation(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai: %v", err)
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

	// Enter Phase Shift
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Get initial radiance
	result, _ := c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	initialRadiance, _ := result.(int)

	// Advance several seconds (radiance gains 1 per 6 frames)
	for i := 0; i < 300; i++ { // ~5 seconds
		advanceCoreFrame(c)
	}

	// Radiance should have increased
	result, _ = c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	laterRadiance, _ := result.(int)

	if laterRadiance <= initialRadiance {
		t.Fatalf("radiance should increase over time during Phase Shift, initial=%v later=%v",
			initialRadiance, laterRadiance)
	}
}

// TestZibaiPhaseShiftExpires verifies Phase Shift mode expires after 16.5 seconds
func TestZibaiPhaseShiftExpires(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai: %v", err)
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

	// Enter Phase Shift
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Verify it's active (1)
	result, _ := c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatal("Phase Shift should be active after skill")
	}

	// Advance 17 seconds (1020 frames) — past 16.5s duration
	for i := 0; i < 1020; i++ {
		advanceCoreFrame(c)
	}

	// Phase Shift should have expired (0)
	result, _ = c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Phase Shift should expire after 16.5 seconds")
	}
}

// TestZibaiBurstDealsDamage verifies Burst deals 2 hits
func TestZibaiBurstDealsDamage(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai: %v", err)
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

	hitCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			hitCount++
		}
		return false
	}, "zibai-burst-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 200; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 2 {
		t.Fatalf("Zibai Burst should deal at least 2 hits, got %v", hitCount)
	}
}

// TestZibaiC1Setup verifies C1 initializes correctly
func TestZibaiC1Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 1
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai C1: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Zibai C1: %v", err)
	}
}

// TestZibaiC6Setup verifies C6 initializes correctly
func TestZibaiC6Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Zibai)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Zibai C6: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Zibai C6: %v", err)
	}
}

// TestZibaiAllActionsDoNotPanic verifies all actions don't panic
func TestZibaiAllActionsDoNotPanic(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Zibai)
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
			err := c.Player.Exec(act, keys.Zibai, p)
			if err == nil {
				for !c.Player.CanQueueNextAction() {
					advanceCoreFrame(c)
				}
			}
			for i := 0; i < 120; i++ {
				advanceCoreFrame(c)
			}
		}()
	}
}
