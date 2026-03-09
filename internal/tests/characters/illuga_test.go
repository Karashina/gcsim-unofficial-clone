package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/illuga"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TestIllugaSkillDealsDamage はスキルが岩元素ダメージを与えることを検証する
func TestIllugaSkillDealsDamage(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga: %v", err)
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
	}, "illuga-skill-hits")

	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Illuga, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 60; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 1 {
		t.Fatalf("Illuga Skill should deal at least 1 hit, got %v", hitCount)
	}
}

// TestIllugaBurstInitializesNightingaleStacks は元素爆発がスタックをセットアップすることを検証する
func TestIllugaBurstInitializesNightingaleStacks(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga: %v", err)
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

	// 元素爆発前: スタックは0であるべき
	result, err := c.Player.Chars()[idx].Condition([]string{"nightingale-stacks"})
	if err != nil {
		t.Fatalf("nightingale-stacks condition error: %v", err)
	}
	if stacks, ok := result.(int); ok && stacks != 0 {
		t.Fatalf("nightingale-stacks should be 0 before burst, got %v", stacks)
	}

	// 元素爆発を実行
	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Illuga, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 100; i++ {
		advanceCoreFrame(c)
	}

	// 元素爆発後: スタックが初期化されているべき（基本值21）
	result, _ = c.Player.Chars()[idx].Condition([]string{"nightingale-stacks"})
	stacks, ok := result.(int)
	if !ok {
		t.Fatalf("nightingale-stacks should return int, got %T", result)
	}
	if stacks < 21 {
		t.Fatalf("nightingale-stacks should be at least 21 after burst, got %v", stacks)
	}
}

// TestIllugaOrioleSongCondition はOriole Songのステータス追跡を検証する
func TestIllugaOrioleSongCondition(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga: %v", err)
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

	// 元素爆発前: Oriole Songは非アクティブ(0)であるべき
	result, _ := c.Player.Chars()[idx].Condition([]string{"oriole-song"})
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Oriole Song should be 0 (inactive) before burst")
	}

	// 元素爆発を実行してOriole Songを有効化
	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Illuga, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// 元素爆発後: Oriole Songはアクティブ(1)であるべき
	result, _ = c.Player.Chars()[idx].Condition([]string{"oriole-song"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatalf("Oriole Song should be 1 (active) after burst, got %v (%T)", result, result)
	}
}

// TestIllugaOrioleSongExpires はOriole Songが持続時間後に終了することを検証する
func TestIllugaOrioleSongExpires(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga: %v", err)
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

	// 元素爆発を実行
	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.Illuga, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// 21秒以上進める（1255 + 60フレームバッファ）
	for i := 0; i < 1315; i++ {
		advanceCoreFrame(c)
	}

	// Oriole Songは終了しているべき(0)
	result, _ := c.Player.Chars()[idx].Condition([]string{"oriole-song"})
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Oriole Song should be 0 (expired) after its duration")
	}
}

// TestIllugaC2Setup はC2が正しく初期化されることを検証する
func TestIllugaC2Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 2
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga C2: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Illuga C2: %v", err)
	}
}

// TestIllugaC6Setup はC6が正しく初期化されることを検証する
func TestIllugaC6Setup(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Illuga)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Illuga C6: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, err := c.AddChar(defProfile(keys.TestCharDoNotUse))
		if err != nil {
			t.Fatalf("error adding test char: %v", err)
		}
	}
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core with Illuga C6: %v", err)
	}
}

// TestIllugaAllActionsDoNotPanic は全アクションがパニックしないことを検証する
func TestIllugaAllActionsDoNotPanic(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Illuga)
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
			err := c.Player.Exec(act, keys.Illuga, p)
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
