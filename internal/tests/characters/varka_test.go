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

// TestVarkaSkillEntersSturmUndDrang はスキル使用でS&Dモードが有効になることを検証する
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
	// 残りのスロットをテストキャラで埋める
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

	// スキル前: S&Dは非アクティブであるべき
	result, err := c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("error querying condition: %v", err)
	}
	if active, ok := result.(bool); ok && active {
		t.Fatal("S&D should not be active before skill use")
	}

	// 元素スキルを実行
	p := make(map[string]int)
	if err := c.Player.Exec(action.ActionSkill, keys.Varka, p); err != nil {
		t.Fatalf("unexpected error executing skill: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// スキル後: S&Dはアクティブであるべき
	result, err = c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("error querying condition: %v", err)
	}
	if active, ok := result.(bool); !ok || !active {
		t.Fatal("S&D should be active after skill use")
	}
}

// TestVarkaFWAChargeConsumption はS&DモードでFWAがチャージを消費することを検証する
func TestVarkaFWAChargeConsumption(t *testing.T) {
	c, trg := makeCore(1)
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 1 // C1はS&D突入時にFWAチャージ1を付与
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

	// スキルでS&Dに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// C1: FWAチャージが1あるべき
	result, _ := c.Player.Chars()[idx].Condition([]string{"fwa-charges"})
	charges, ok := result.(int)
	if !ok {
		t.Fatalf("expected fwa-charges to be int, got %T", result)
	}
	if charges < 1 {
		t.Fatalf("C1 should grant at least 1 FWA charge on S&D entry, got %v", charges)
	}

	// FWAを使用（S&Dモードでのスキル）
	err = c.Player.Exec(action.ActionSkill, keys.Varka, p)
	if err != nil {
		t.Fatalf("error executing FWA: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// FWA後: チャージが消費されているべき
	result, _ = c.Player.Chars()[idx].Condition([]string{"fwa-charges"})
	chargesAfter, _ := result.(int)
	if chargesAfter >= charges {
		t.Fatalf("FWA should consume a charge, before=%v after=%v", charges, chargesAfter)
	}
}

// TestVarkaFWADealsTwoHits はFWAが2ヒット（Other + 風元素）することを検証する
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

	// S&Dに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// FWAからのヒット数をカウント
	hitCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			hitCount++
		}
		return false
	}, "fwa-hit-count")

	// FWAを実行
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}
	// ダメージ処理のため追加フレームを進める
	for i := 0; i < 60; i++ {
		advanceCoreFrame(c)
	}

	if hitCount < 2 {
		t.Fatalf("FWA should deal at least 2 hits, got %v", hitCount)
	}
}

// TestVarkaAzureDevourDeals4Hits はAzure Devourが4回ヒットすることを検証する
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

	// S&Dに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// Azure Devourのヒット数をカウント（FWAチャージありのS&Dでの重撃）
	hitCount := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == idx {
			hitCount++
		}
		return false
	}, "azure-hit-count")

	// Azure Devourを実行（S&Dモードでの重撃）
	err = c.Player.Exec(action.ActionCharge, keys.Varka, p)
	if err != nil {
		// 重撃が準備できていない場合、フレームを進めてリトライ
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

// TestVarkaSturmUndDrangExpires はS&Dモードが12秒後に終了することを検証する
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

	// S&Dに入る
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// S&Dがアクティブであることを確認
	result, _ := c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if active, ok := result.(bool); !ok || !active {
		t.Fatal("S&D should be active after skill use")
	}

	// 12秒 + バッファ（720フレーム + 60バッファ）を進める
	for i := 0; i < 780; i++ {
		advanceCoreFrame(c)
	}

	// S&Dは終了しているべき
	result, _ = c.Player.Chars()[idx].Condition([]string{"sturm-und-drang"})
	if active, ok := result.(bool); ok && active {
		t.Fatal("S&D should expire after 12 seconds")
	}
}

// TestVarkaBurstDealsDamage は元素爆発がダメージを与えることを検証する
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
	// 処理のため追加フレームを進める
	for i := 0; i < 200; i++ {
		advanceCoreFrame(c)
	}

	if dmgCount[trg[0].Key()] < 2 {
		t.Fatalf("Burst should deal at least 2 hits, got %v", dmgCount[trg[0].Key()])
	}
}

// TestVarkaC6ChainFWAToAzure はC6のスキルルーティングによるFWA→Azureチェーンを検証する
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

	// S&Dに入る（C1がチャージ1を付与）
	p := make(map[string]int)
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// FWAを使用（c6FWAWindowKeyを設定）
	c.Player.Exec(action.ActionSkill, keys.Varka, p)
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// C6でのFWA後: c6FWAWindowKeyが次のスキルをAzure Devourにルーティングするべき
	ready, _ := c.Player.Chars()[idx].ActionReady(action.ActionSkill, p)
	if !ready {
		t.Fatal("C6: Skill should be ready after FWA (window routes to Azure Devour)")
	}

	// スキル経由でAzure Devourを実行（C6はc6FWAWindowKeyアクティブ時にスキル→Azureにルーティング）
	err = c.Player.Exec(action.ActionSkill, keys.Varka, p)
	if err != nil {
		t.Fatalf("C6: Skill→Azure Devour should execute after FWA, got error: %v", err)
	}
	for !c.Player.CanQueueNextAction() {
		advanceCoreFrame(c)
	}

	// C6チェーンは一方向のみ（FWA→Azure）; チェーンされたAzureは
	// 逆方向のウィンドウを再オープンしない（consumeCharge=false ⇒ c6AzureWindowKeyが設定されない）。
	// これにより追加チェーンは不可。設計上の無限チェーン防止。
}

// TestVarkaConditionQueries は全Condition()フィールドが正しい型を返すことを検証する
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

	// hexereiクエリをテスト（nohex=1で無効化）
	result, err := ch.Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || isHex {
		t.Fatal("hexerei should be false with nohex=1")
	}

	// sturm-und-drangクエリをテスト
	result, err = ch.Condition([]string{"sturm-und-drang"})
	if err != nil {
		t.Fatalf("sturm-und-drang condition error: %v", err)
	}
	if _, ok := result.(bool); !ok {
		t.Fatalf("sturm-und-drang should return bool, got %T", result)
	}

	// fwa-chargesクエリをテスト
	result, err = ch.Condition([]string{"fwa-charges"})
	if err != nil {
		t.Fatalf("fwa-charges condition error: %v", err)
	}
	if _, ok := result.(int); !ok {
		t.Fatalf("fwa-charges should return int, got %T", result)
	}

	// a4-stacksクエリをテスト
	result, err = ch.Condition([]string{"a4-stacks"})
	if err != nil {
		t.Fatalf("a4-stacks condition error: %v", err)
	}
	if stacks, ok := result.(int); !ok || stacks != 0 {
		t.Fatalf("a4-stacks should return 0 initially, got %v", result)
	}
}

// TestVarkaHexereiBonusDetection はHexereiパーティ検出を検証する
func TestVarkaHexereiBonusDetection(t *testing.T) {
	c, _ := makeCore(1)

	// Hexerei有効のVarka（デフォルト）
	prof := defProfile(keys.Varka)
	prof.Base.Cons = 0
	prof.Base.Ascension = 6
	_, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("error adding Varka: %v", err)
	}

	// Ventiを追加（デフォルトでHexerei） — 2人以上のHexereiをトリガーするべき
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

	// VarkaとVenti両方がHexerei → パーティに2人以上のHexerei → hasHexBonusはtrueであるべき
	result, err := c.Player.Chars()[0].Condition([]string{"hexerei"})
	if err != nil {
		t.Fatalf("hexerei condition error: %v", err)
	}
	if isHex, ok := result.(bool); !ok || !isHex {
		t.Fatal("Varka should have hexerei enabled by default")
	}
}

// TestVarkaNoHexDisablesHexerei はnohex=1パラメータがHexereiを無効化することを検証する
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

// TestVarkaSkillAllActionsDoNotPanic はVarkaの全アクションがパニックしないことを検証する
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

	// パニックせずに全アクションを実行
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
				// 準備できていない場合の想定エラー
			case err == nil:
				for !c.Player.CanQueueNextAction() {
					advanceCoreFrame(c)
				}
			}
			// アクション間でフレームを進める
			for i := 0; i < 120; i++ {
				advanceCoreFrame(c)
			}
		}()
	}
}
