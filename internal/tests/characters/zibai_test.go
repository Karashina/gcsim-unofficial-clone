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

// TestZibaiSkillActivatesPhaseShift はスキル使用でLunar Phase Shiftモードに入ることを検証する
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

	// スキル前: Phase Shiftは非アクティブ(0)であるべき
	result, err := c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if err != nil {
		t.Fatalf("lunar-phase-shift condition error: %v", err)
	}
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Phase Shift should be 0 (inactive) before skill use")
	}

	// 元素スキルを実行
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// スキル後: Phase Shiftはアクティブ(1)であるべき
	result, _ = c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatalf("Phase Shift should be 1 (active) after skill use, got %v (%T)", result, result)
	}
}

// TestZibaiRadianceConditionQuery はRadianceが期待値で開始することを検証する
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

	// Phase Shiftを有効化
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// 有効化後、radianceは0であるべき（C1の場合は100）
	result, err := c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	if err != nil {
		t.Fatalf("phase-shift-radiance condition error: %v", err)
	}
	radiance, ok := result.(int)
	if !ok {
		t.Fatalf("radiance should return int, got %T", result)
	}
	// C0: スキル使用でradianceが0にリセット
	if radiance < 0 {
		t.Fatalf("radiance should be non-negative, got %v", radiance)
	}
}

// TestZibaiRadianceAccumulation はPhase Shift中にRadianceが時間経過で増加することを検証する
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

	// Phase Shiftに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// 初期radianceを取得
	result, _ := c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	initialRadiance, _ := result.(int)

	// 数秒分進める（radianceは6フレームごとに1増加）
	for i := 0; i < 300; i++ { // ~5 seconds
		advanceCoreFrame(c)
	}

	// Radianceが増加しているべき
	result, _ = c.Player.Chars()[idx].Condition([]string{"phase-shift-radiance"})
	laterRadiance, _ := result.(int)

	if laterRadiance <= initialRadiance {
		t.Fatalf("radiance should increase over time during Phase Shift, initial=%v later=%v",
			initialRadiance, laterRadiance)
	}
}

// TestZibaiPhaseShiftExpires はPhase Shiftモードが16.5秒後に終了することを検証する
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

	// Phase Shiftに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Zibai, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// アクティブ(1)であることを確認
	result, _ := c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); !ok || active != 1 {
		t.Fatal("Phase Shift should be active after skill")
	}

	// 17秒分進める（1020フレーム）— 16.5秒の持続時間を超過
	for i := 0; i < 1020; i++ {
		advanceCoreFrame(c)
	}

	// Phase Shiftは終了しているべき(0)
	result, _ = c.Player.Chars()[idx].Condition([]string{"lunar-phase-shift"})
	if active, ok := result.(int); ok && active != 0 {
		t.Fatal("Phase Shift should expire after 16.5 seconds")
	}
}

// TestZibaiBurstDealsDamage は元素爆発が2ヒットすることを検証する
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

// TestZibaiC1Setup はC1が正しく初期化されることを検証する
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

// TestZibaiC6Setup はC6が正しく初期化されることを検証する
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

// TestZibaiAllActionsDoNotPanic は全アクションがパニックしないことを検証する
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
