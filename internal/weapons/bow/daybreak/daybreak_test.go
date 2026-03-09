package daybreak

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

func TestDaybreakHitIncreasesBonus(t *testing.T) {
	core.RegisterCharFunc(keys.TestCharDoNotUse, testhelper.NewChar)

	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatalf("core.New: %v", err)
	}

	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Rarity = 5
	p.Base.Element = attributes.Geo
	p.Base.Level = 90
	p.Weapon = info.WeaponProfile{Key: keys.TheDaybreakChronicles, Refine: 1, Level: 1}
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

	// ICDを超過したフレームであることを確認し、最初のヒットが適用されるようにする
	c.F = hitICDFrames + 1

	// ヒットをシミュレートして通常攻撃ボーナスを増加させる
	hit := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, nil, hit)

	// 攻撃を構築してアタックModを適用；ダメージ%が適用されるべき
	atk := &combat.AttackEvent{Info: combat.AttackInfo{ActorIndex: idx, AttackTag: attacks.AttackTagNormal}}
	// スナップショットのステータスはデフォルトでゼロ；ApplyAttackModsがスナップショットを設定する
	char.ApplyAttackMods(atk, nil)
	if atk.Snapshot.Stats[attributes.DmgP] <= 0 {
		t.Fatalf("expected DmgP bonus to be applied after hit, got %v", atk.Snapshot.Stats[attributes.DmgP])
	}
}
