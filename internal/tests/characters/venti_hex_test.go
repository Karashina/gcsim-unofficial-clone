package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/venti"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TestVentiSkillDealsDamage verifies Venti Skill (Skyward Sonnet) deals Anemo DMG
func TestVentiSkillDealsDamage(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Venti)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
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
	}, "venti-skill-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Venti, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 120; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 1 {
		t.Fatalf("Venti Skill should deal at least 1 hit, got %v", hitCount)
	}
}

// TestVentiBurstCreatesBurstEye verifies Burst spawns the Wind's Grand Ode eye
func TestVentiBurstCreatesBurstEye(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Venti)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
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
	}, "venti-burst-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Venti, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Advance several seconds for burst ticks
	for i := 0; i < 600; i++ {
		advanceCoreFrame(c)
	}

	// Burst should deal multiple tick hits over its duration
	if hitCount < 5 {
		t.Fatalf("Venti Burst should deal multiple tick hits, got %v", hitCount)
	}
}

// TestVentiHexereiCondition verifies hexerei condition query
func TestVentiHexereiCondition(t *testing.T) {
	c, _ := makeCore(1)

	// Venti with Hexerei enabled (default)
	prof := defProfile(keys.Venti)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	// Hexerei should be true by default
	result, err := c.Player.Chars()[idx].Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || !isHex {
		t.Fatal("Venti should have hexerei=true by default")
	}
}

// TestVentiNoHexDisablesHexerei verifies nohex=1 disables Hexerei
func TestVentiNoHexDisablesHexerei(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Venti)
	prof.Params["nohex"] = 1
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	result, _ := c.Player.Chars()[idx].Condition([]string{"hexerei"})
	if isHex, ok := result.(bool); ok && isHex {
		t.Fatal("Venti hexerei should be false with nohex=1")
	}
}

// TestVentiC6Setup verifies C6 initializes without error
func TestVentiC6Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Venti)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti C6: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Venti C6: %v", err)
	}
}

// TestVentiAllActionsDoNotPanic verifies all actions don't panic
func TestVentiAllActionsDoNotPanic(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Venti)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Venti: %v", err)
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
			err := c.Player.Exec(act, keys.Venti, p)
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
