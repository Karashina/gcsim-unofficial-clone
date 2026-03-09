package gestofthemightywolf

import (
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/testhelper"
)

func init() {
	core.RegisterCharFunc(keys.TestCharDoNotUse, testhelper.NewChar)
}

func TestGestStacksOnNormalHit(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// ヒット前: DmgPは0であるべき（スタックなし）
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	if atk.Snapshot.Stats[attributes.DmgP] != 0 {
		t.Fatalf("expected DmgP=0 before first hit, got %v", atk.Snapshot.Stats[attributes.DmgP])
	}

	// ICDを超えて進める
	c.F = stackICD + 2

	// 通常攻撃ヒット → 1スタック
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// DmgPが適用されているか確認
	atk2 := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk2, nil)
	if atk2.Snapshot.Stats[attributes.DmgP] <= 0 {
		t.Fatalf("expected DmgP > 0 after NA hit (1 stack), got %v", atk2.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestStacksOnSkillHit(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// ICDを超えて進める
	c.F = stackICD + 2

	// スキルヒット → 2スタック
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagElementalArt}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// DmgPを確認（R1で 2 × 7.5% = 15%のはず）
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	expectedDmg := 0.075 * 2 // R1で2スタック × 7.5%/スタック
	if atk.Snapshot.Stats[attributes.DmgP] < expectedDmg-0.001 || atk.Snapshot.Stats[attributes.DmgP] > expectedDmg+0.001 {
		t.Fatalf("expected DmgP ≈ %v after skill hit (2 stacks), got %v", expectedDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestMaxStacks(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// 6スタック分のヒットを適用（4でキャップされるべき）
	for i := 0; i < 6; i++ {
		c.F += stackICD + 2 // 毎回ICDを超えて進める
		hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
		c.Events.Emit(event.OnEnemyHit, nil, hit)
	}

	// DmgPが 4 × 7.5% = 30%でキャップされているか確認
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	maxDmg := 0.075 * 4 // 最大4スタック × 7.5%
	if atk.Snapshot.Stats[attributes.DmgP] > maxDmg+0.001 {
		t.Fatalf("DmgP should be capped at %v (4 stacks), got %v", maxDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestChargedAttackGives2Stacks(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// ICDを超えて進める
	c.F = stackICD + 2

	// 重撃ヒット → 2スタック
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagExtra}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// DmgPを確認（2 × 7.5% = 15%のはず）
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	expectedDmg := 0.075 * 2
	if atk.Snapshot.Stats[attributes.DmgP] < expectedDmg-0.001 || atk.Snapshot.Stats[attributes.DmgP] > expectedDmg+0.001 {
		t.Fatalf("expected DmgP ≈ %v after CA hit (2 stacks), got %v", expectedDmg, atk.Snapshot.Stats[attributes.DmgP])
	}
}

func TestGestBuffExpiresAfter4Seconds(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// スタックを追加
	c.F = stackICD + 2
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// バフがアクティブであることを確認
	if !char.StatusIsActive(buffKey) {
		t.Fatal("buff should be active after hit")
	}

	// 5秒進める（300フレーム）— 4秒のバフ持続時間を超過
	c.F += 300

	// バフが失効しているべき
	if char.StatusIsActive(buffKey) {
		t.Fatal("buff should expire after 4 seconds")
	}
}

func TestGestNoCritDMGWithoutHexerei(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// スタックを追加
	c.F = stackICD + 2
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// Hexereiパーティなし: CDボーナスは適用されないべき
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	char.ApplyAttackMods(atk, nil)
	if atk.Snapshot.Stats[attributes.CD] != 0 {
		t.Fatalf("expected no CRIT DMG bonus without Hexerei party, got %v", atk.Snapshot.Stats[attributes.CD])
	}
}

func TestGestAtkSpdPassive(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.GestOfTheMightyWolf, Refine: 1, Level: 90}
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	p.Stats = make([]float64, attributes.EndStatType)
	p.Sets = make(map[keys.Set]int)

	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatalf("AddChar: %v", err)
	}

	if err := c.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	char := c.Player.Chars()[idx]

	// 攻撃速度が永続的に+10%であるべき
	atkSpd := char.Stat(attributes.AtkSpd)
	if atkSpd < 0.09 || atkSpd > 0.11 {
		t.Fatalf("expected ATK SPD ≈ 0.10, got %v", atkSpd)
	}
}
