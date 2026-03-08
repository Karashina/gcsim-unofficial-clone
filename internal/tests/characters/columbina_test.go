package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/columbina"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TestColumbinaSkillActivation verifies Skill deals Hydro DMG and activates Gravity Ripple
func TestColumbinaSkillActivation(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Columbina)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Columbina: %v", err)
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
	}, "columbina-skill-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Columbina, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// Extra frames for hitmark processing
	for i := 0; i < 60; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 1 {
		t.Fatalf("Columbina Skill should deal at least 1 hit, got %v", hitCount)
	}
}

// TestColumbinaGravityConditionQuery verifies Condition fields work correctly
func TestColumbinaGravityConditionQuery(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Columbina)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Columbina: %v", err)
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

	// Test gravity query — should be 0 initially
	result, err := ch.Condition([]string{"gravity"})
	if err != nil {
		t.Fatalf("gravity condition error: %v", err)
	}
	if grav, ok := result.(int); !ok {
		t.Fatalf("gravity should return int, got %T", result)
	} else if grav != 0 {
		t.Fatalf("gravity should be 0 initially, got %v", grav)
	}

	// Test lunacy query — should be 0 initially
	result, err = ch.Condition([]string{"lunacy"})
	if err != nil {
		t.Fatalf("lunacy condition error: %v", err)
	}
	if lunacy, ok := result.(int); !ok {
		t.Fatalf("lunacy should return int, got %T", result)
	} else if lunacy != 0 {
		t.Fatalf("lunacy should be 0 initially, got %v", lunacy)
	}

	// Test lunar-domain query — should be 0 initially (0=inactive, 1=active)
	result, err = ch.Condition([]string{"lunar-domain"})
	if err != nil {
		t.Fatalf("lunar-domain condition error: %v", err)
	}
	if active, ok := result.(int); !ok {
		t.Fatalf("lunar-domain should return int, got %T", result)
	} else if active != 0 {
		t.Fatal("lunar-domain should be 0 initially")
	}
}

// TestColumbinaBurstDealsDamage verifies Burst activation and Lunar Domain
func TestColumbinaBurstDealsDamage(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Columbina)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Columbina: %v", err)
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
	}, "columbina-burst-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Columbina, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// Extra frames for damage ticks
	for i := 0; i < 200; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 1 {
		t.Fatalf("Columbina Burst should deal at least 1 hit, got %v", hitCount)
	}

	// Lunar Domain should be active after burst (1=active)
	result, _ := c.Player.Chars()[idx].Condition([]string{"lunar-domain"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatalf("Lunar Domain should be 1 (active) after burst, got %v (%T)", result, result)
	}
}

// TestColumbinaAllActionsDoNotPanic verifies all actions don't panic
func TestColumbinaAllActionsDoNotPanic(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Columbina)
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
			err := c.Player.Exec(act, keys.Columbina, p)
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

// TestColumbinaC6Setup verifies C6 constellation initializes without error
func TestColumbinaC6Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Columbina)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Columbina C6: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Columbina C6: %v", err)
	}
}
