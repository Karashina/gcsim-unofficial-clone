package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillFrames     []int
	northlandFrames []int
)

const (
	northlandHitmark = 23
	skillKey         = "flins-skill"
	northlandKey     = "flins-northland"
	northlandCdKey   = "flins-northland-cd"
	particleICDKey   = "flins-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(25)     // E -> D/J
	northlandFrames = frames.InitAbilSlice(37) // E -> D/J
}

// E
/*
Flinsが「顯現の炎」形態に入り、skillKeyを取得する。この形態には以下の特性がある:
・Flinsの通常攻撃と重撃の元素が雷元素に変換され、落下攻撃が使用不可になる。
・元素スキル「Ancient Rite: Arcane Light」が特殊元素スキル「北地の槍嵐」に置き換わる。

北地の槍嵐
範囲雷元素ダメージを与え、次の6秒間northlandKeyを付与する。

---COPILOT向け追加情報---
「Flinsの通常攻撃と重撃の元素が雷元素に変換され、落下攻撃が使用不可になる」の実装は、
attack.goとcharge.goで処理すべき。追加の実装例はcynoを参照。
*/
func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		aiNorthland := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Northland Spearstorm DMG (E)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skill[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(
			aiNorthland,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 5),
			northlandHitmark, northlandHitmark, c.particleCB,
		)

		c.AddStatus(northlandKey, 6*60, true)
		c.AddStatus(northlandCdKey, c.northlandCD, true)

		// 2命ノ星座: 次の通常攻撃で追加ダメージのステータスを設定
		if c.Base.Cons >= 2 {
			c.AddStatus(c2NorthlandKey, 6*60, true)
		}

		return action.Info{
			Frames:          frames.NewAbilFunc(northlandFrames),
			AnimationLength: northlandFrames[action.InvalidAction],
			CanQueueAfter:   northlandFrames[action.ActionSwap], // 最速キャンセル
			State:           action.SkillState,
		}, nil
	} else {
		c.AddStatus(skillKey, 10*60, true) // 10s
		c.SetCD(action.ActionSkill, 16*60)

		return action.Info{
			Frames:          frames.NewAbilFunc(skillFrames),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセル
			State:           action.SkillState,
		}, nil
	}
}

// 粒子生成コールバック - 雷元素付与攻撃が命中するたびに呼び出される
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2.1*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
}
