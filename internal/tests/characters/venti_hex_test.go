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

// TestVentiSkillDealsDamage はVentiのスキル（Skyward Sonnet）が風元素ダメージを与えることを検証する
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

// TestVentiBurstCreatesBurstEye は元素爆発がWind's Grand Odeの目を生成することを検証する
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

	// 元素爆発のtickのため数秒分進める
	for i := 0; i < 600; i++ {
		advanceCoreFrame(c)
	}

	// 元素爆発は持続時間中に複数のtickヒットを与えるべき
	if hitCount < 5 {
		t.Fatalf("Venti Burst should deal multiple tick hits, got %v", hitCount)
	}
}

// TestVentiHexereiCondition はhexereiのConditionクエリを検証する
func TestVentiHexereiCondition(t *testing.T) {
	c, _ := makeCore(1)

	// Hexerei有効のVenti（デフォルト）
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

	// デフォルトでHexereiはtrueであるべき
	result, err := c.Player.Chars()[idx].Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || !isHex {
		t.Fatal("Venti should have hexerei=true by default")
	}
}

// TestVentiNoHexDisablesHexerei はnohex=1がHexereiを無効化することを検証する
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

// TestVentiC6Setup はC6がエラーなく初期化されることを検証する
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

// TestVentiAllActionsDoNotPanic は全アクションがパニックしないことを検証する
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
