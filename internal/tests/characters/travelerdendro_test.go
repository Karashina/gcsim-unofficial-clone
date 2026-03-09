package characters

import (
	"log"
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/common/dendro"
	_ "github.com/Karashina/gcsim-unofficial-clone/internal/characters/traveler/dendro/aether"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

func TestTravelerDendroBurstAttach(t *testing.T) {
	c, trg := makeCore(2)
	prof := defProfile(keys.AetherDendro)
	prof.Base.Cons = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Errorf("error adding char: %v", err)
		t.FailNow()
	}
	c.Player.SetActive(idx)
	err = c.Init()
	if err != nil {
		t.Errorf("error initializing core: %v", err)
		t.FailNow()
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.Events.Subscribe(event.OnGadgetHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		log.Printf("hit by %v attack, dur %v", atk.Info.Element, atk.Info.Durability)
		return false
	}, "hit-check")
	advanceCoreFrame(c)

	// 元素爆発で玉を生成
	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.AetherDendro, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// 草元素ガジェットが生成されるまで待機
	for c.Combat.GadgetCount() < 1 {
		advanceCoreFrame(c)
	}
	// 安全のため追加で1フレーム進める
	advanceCoreFrame(c)

	// ガジェットに草元素が付着していることを確認
	g := c.Combat.Gadget(0)
	gr, ok := g.(*dendro.LeaLotus)
	if !ok {
		t.Errorf("expecting gadget to be lea lotus. failed")
		t.FailNow()
	}
	log.Println("initial aura string: ", gr.ActiveAuraString())
	if gr.Durability[reactable.Dendro] != 10 {
		t.Errorf("expecting initial 10 dendro on traveler lea lotus, got %v", gr.Durability[reactable.Dendro])
	}

	// パターンはガジェットにのみ命中
	pattern := combat.NewCircleHitOnTarget(geometry.Point{}, nil, 100)
	pattern.SkipTargets[targets.TargettableEnemy] = true

	// 氷元素の付着をチェック
	c.QueueAttackEvent(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Cryo,
			Durability: 100,
		},
		Pattern: pattern,
	}, 0)
	advanceCoreFrame(c)

	log.Println("after applying 100 cyro: ", gr.ActiveAuraString())
	if gr.Durability[reactable.Cryo] != 80 {
		t.Errorf("expecting 80 dendro on traveler lea lotus, got %v", gr.Durability[reactable.Cryo])
	}
	if gr.Durability[reactable.Dendro] != 10 {
		t.Errorf("expecting 10 dendro on traveler lea lotus, got %v", gr.Durability[reactable.Dendro])
	}
}

func TestTravelerDendroBurstPyro(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.AetherDendro)
	prof.Base.Cons = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Errorf("error adding char: %v", err)
		t.FailNow()
	}
	c.Player.SetActive(idx)
	err = c.Init()
	if err != nil {
		t.Errorf("error initializing core: %v", err)
		t.FailNow()
	}
	c.Combat.DefaultTarget = trg[0].Key()
	c.Events.Subscribe(event.OnGadgetHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		log.Printf("gadget hit by %v attack, dur %v", atk.Info.Element, atk.Info.Durability)
		return false
	}, "hit-check")
	dmgCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Abil == "Lea Lotus Lamp Explosion" {
			dmgCount++
			log.Println("big boom at: ", c.F)
		}
		return false
	}, "hit-check")
	advanceCoreFrame(c)

	// 元素爆発で玉を生成
	p := make(map[string]int)
	c.Player.Exec(action.ActionBurst, keys.AetherDendro, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// 草元素ガジェットが生成されるまで待機
	for c.Combat.GadgetCount() < 1 {
		advanceCoreFrame(c)
	}
	// 安全のため追加で1フレーム進める
	advanceCoreFrame(c)

	// ガジェットに草元素が付着していることを確認
	g := c.Combat.Gadget(0)
	gr, ok := g.(*dendro.LeaLotus)
	if !ok {
		t.Errorf("expecting gadget to be lea lotus. failed")
		t.FailNow()
	}
	log.Println("initial aura string: ", gr.ActiveAuraString())
	if gr.Durability[reactable.Dendro] != 10 {
		t.Errorf("expecting initial 10 dendro on traveler lea lotus, got %v", gr.Durability[reactable.Dendro])
	}

	// パターンはガジェットにのみ命中
	pattern := combat.NewCircleHitOnTarget(geometry.Point{}, nil, 100)
	pattern.SkipTargets[targets.TargettableEnemy] = true

	// 氷元素の付着をチェック
	c.QueueAttackEvent(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Pyro,
			Durability: 100,
		},
		Pattern: pattern,
	}, 0)
	advanceCoreFrame(c)

	log.Printf("at f %v after applying 100 pyro: %v\n", c.F, gr.ActiveAuraString())
	if gr.Durability[reactable.Pyro] != 0 {
		t.Errorf("expecting 0 dendro on traveler lea lotus, got %v", gr.Durability[reactable.Pyro])
	}

	// 60フレーム後に爆発が発生するべき
	for i := 0; i < 100; i++ {
		advanceCoreFrame(c)
	}

	if dmgCount != 1 {
		t.Errorf("expected 1 dmg count, got %v", dmgCount)
	}
}

// lotusは出現後37フレーム目にtickすることが期待される（キャスト後54+37フレーム）
// その後は持続時間中90フレームごとにtick
// 持続時間はC0で12秒、C2で15秒
func TestTravelerDendroBurstTicks(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.AetherDendro)
	prof.Base.Cons = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Errorf("error adding char: %v", err)
		t.FailNow()
	}
	c.Player.SetActive(idx)
	err = c.Init()
	if err != nil {
		t.Errorf("error initializing core: %v", err)
		t.FailNow()
	}
	c.Combat.DefaultTarget = trg[0].Key()
	dmgCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Abil == "Lea Lotus Lamp" {
			dmgCount++
			log.Println("boom at (adjusted): ", c.F-54-1)
		}
		return false
	}, "hit-check")
	advanceCoreFrame(c)

	// 元素爆発で玉を生成
	p := make(map[string]int)
	log.Println("casting burst: ", c.F)
	c.Player.Exec(action.ActionBurst, keys.AetherDendro, p)

	// 出現まで合計54フレーム + 15秒の持続時間を期待
	totalDuration := 15 * 60
	expectedCount := 1 + (totalDuration-37)/90

	// 余分なtickのバグに備えて100フレーム追加
	for i := 0; i < 54+totalDuration+100; i++ {
		advanceCoreFrame(c)
	}

	if dmgCount != expectedCount {
		t.Errorf("expecting %v ticks, got %v", expectedCount, dmgCount)
	}
}

func TestTravelerDendroBurstElectroTicks(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.AetherDendro)
	prof.Base.Cons = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Errorf("error adding char: %v", err)
		t.FailNow()
	}
	c.Player.SetActive(idx)
	err = c.Init()
	if err != nil {
		t.Errorf("error initializing core: %v", err)
		t.FailNow()
	}
	c.Combat.DefaultTarget = trg[0].Key()
	dmgCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Abil == "Lea Lotus Lamp" {
			dmgCount++
			log.Println("boom at (adjusted): ", c.F-54-1)
		}
		return false
	}, "hit-check")
	advanceCoreFrame(c)

	// 元素爆発で玉を生成
	p := make(map[string]int)
	log.Println("casting burst: ", c.F)
	c.Player.Exec(action.ActionBurst, keys.AetherDendro, p)
	// 草元素ガジェットが生成されるまで待機
	for c.Combat.GadgetCount() < 1 {
		advanceCoreFrame(c)
	}

	// パターンはガジェットにのみ命中
	pattern := combat.NewCircleHitOnTarget(geometry.Point{}, nil, 100)
	pattern.SkipTargets[targets.TargettableEnemy] = true

	// 氷元素の付着をチェック
	c.QueueAttackEvent(&combat.AttackEvent{
		Info: combat.AttackInfo{
			Element:    attributes.Electro,
			Durability: 100,
		},
		Pattern: pattern,
	}, 0)

	// 最初のtickは15フレーム目、その後は54フレームごとにtick
	totalDuration := 15 * 60
	expectedCount := 1 + (totalDuration-15)/54

	// 余分なtickのバグに備えて100フレーム追加
	for i := 0; i < totalDuration+100; i++ {
		advanceCoreFrame(c)
	}

	if dmgCount != expectedCount {
		t.Errorf("expecting %v ticks, got %v", expectedCount, dmgCount)
	}
}
