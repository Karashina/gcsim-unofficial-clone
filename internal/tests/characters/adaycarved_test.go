package characters

import (
	"testing"

	_ "github.com/Karashina/gcsim-unofficial-clone/internal/artifacts/adaycarvedfromrisingwinds"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/cyno"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/kirara"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TestADCFRW4pcNormalAttackTrigger は通常攻撃が4セットバフを発動することを検証する
func TestADCFRW4pcNormalAttackTrigger(t *testing.T) {
	c, _ := makeCore(1)
	prof := defProfile(keys.Cyno)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Sets[keys.ADayCarvedFromRisingWinds] = 4
	prof.SetParams[keys.ADayCarvedFromRisingWinds] = make(map[string]int)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("Cyno の追加に失敗: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := c.AddChar(defProfile(keys.TestCharDoNotUse)); err != nil {
			t.Fatalf("ダミーキャラの追加に失敗: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("コアの初期化に失敗: %v", err)
	}
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	triggered := false
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != idx {
			return false
		}
		if c.Player.Chars()[idx].StatusIsActive("adaycarved-4pc-buff") {
			triggered = true
		}
		return false
	}, "adcfrw-normal-check")

	p := make(map[string]int)
	c.Player.Exec(action.ActionAttack, keys.Cyno, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// バフが発動するまで追加フレームを進める
	for i := 0; i < 60; i++ {
		advanceCoreFrame(c)
	}

	if !triggered {
		t.Fatal("通常攻撃後に 4pc バフが発動していない")
	}
}

// TestADCFRW4pcHoldSkillTrigger は長押し元素スキル（AttackTagElementalArtHold）が4セットバフを発動することを検証する
// KiraraのskillHoldはroll tickにAttackTagElementalArtHoldを使用
func TestADCFRW4pcHoldSkillTrigger(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Kirara)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Sets[keys.ADayCarvedFromRisingWinds] = 4
	prof.SetParams[keys.ADayCarvedFromRisingWinds] = make(map[string]int)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("Kirara の追加に失敗: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := c.AddChar(defProfile(keys.TestCharDoNotUse)); err != nil {
			t.Fatalf("ダミーキャラの追加に失敗: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("コアの初期化に失敗: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	triggered := false
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != idx {
			return false
		}
		if c.Player.Chars()[idx].StatusIsActive("adaycarved-4pc-buff") {
			triggered = true
		}
		return false
	}, "adcfrw-holdskill-check")

	// p["hold"]=60 → Kirara の skillHold(60) を発動。roll ticks が AttackTagElementalArtHold を使用
	p := map[string]int{"hold": 60}
	c.Player.Exec(action.ActionSkill, keys.Kirara, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 120; i++ {
		advanceCoreFrame(c)
	}

	if !triggered {
		t.Fatal("Kirara Hold Skill 後に 4pc バフが発動していない (AttackTagElementalArtHold が未処理)")
	}
}

// TestADCFRW4pcBuffDuration はバフが6秒（360フレーム）後に失効することを検証する
func TestADCFRW4pcBuffDuration(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Cyno)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	prof.Sets[keys.ADayCarvedFromRisingWinds] = 4
	prof.SetParams[keys.ADayCarvedFromRisingWinds] = make(map[string]int)
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("Cyno の追加に失敗: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := c.AddChar(defProfile(keys.TestCharDoNotUse)); err != nil {
			t.Fatalf("ダミーキャラの追加に失敗: %v", err)
		}
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("コアの初期化に失敗: %v", err)
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.QueueParticle("system", 1000, attributes.NoElement, 0)
	advanceCoreFrame(c)

	// 通常攻撃でバフを発動
	p := make(map[string]int)
	c.Player.Exec(action.ActionAttack, keys.Cyno, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	for i := 0; i < 30; i++ {
		advanceCoreFrame(c)
	}

	// バフが発動していることを確認
	if !c.Player.Chars()[idx].StatusIsActive("adaycarved-4pc-buff") {
		t.Fatal("通常攻撃後にバフが発動していない")
	}

	// 6s = 360フレーム + 余裕分 進める
	for i := 0; i < 400; i++ {
		advanceCoreFrame(c)
	}

	// バフが期限切れになっているか確認
	if c.Player.Chars()[idx].StatusIsActive("adaycarved-4pc-buff") {
		t.Fatal("6s 後にバフがまだ有効になっている（期限切れになっていない）")
	}
}
