package nefer

import (
	"reflect"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

var (
	skillFrames []int
)

const (
	skillHitmark = 26
	skillKey     = "nefer-skill"
)

func init() {
	skillFrames = frames.InitAbilSlice(31)
}

// 元素スキル：範囲草元素 + Shadow Dance状態、チャージを付与。Shadow Dance中、翠露が幻影公演を可能にする。
func (c *char) Skill(p map[string]int) (action.Info, error) {

	// スキルダメージはATKとEMの両方のスケーリングあり
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Skill Initial DMG (E)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    skillatk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillem[c.TalentLvlSkill()]*c.Stat(attributes.EM),
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 5),
			0,
			0,
			c.particleCB,
		)

		if c.Base.Cons >= 2 {
			// Shadow Dance状態に入る
			c.AddStatus(skillKey, 15*60, true) // 15秒の持続時間
		} else {
			c.AddStatus(skillKey, 10*60, true) // 10秒の持続時間
		}

		// ムーンサインAscendantの場合：既存の草元素コアを欺きの種に変換し、15秒の変換ウィンドウを設定
		// デバッグ用にムーンサイン状態をログ
		c.Core.Log.NewEvent("nefer skill moonsign state", glog.LogDebugEvent, c.Index).
			Write("moonsign_nascent", c.MoonsignNascent).
			Write("moonsign_ascendant", c.MoonsignAscendant)

		if c.MoonsignAscendant {
			// ステータスウィンドウを設定
			c.AddStatus("nefer-seed-convert", 15*60, true)
			// 既存の草元素コアを変換
			gad := c.Core.Combat.Gadgets()
			c.Core.Log.NewEvent("nefer skill found gadgets", glog.LogDebugEvent, c.Index).
				Write("count", len(gad))
			for _, g := range gad {
				if g == nil {
					continue
				}
				// ガジェットの基本情報をログ
				c.Core.Log.NewEvent("nefer skill gadget info", glog.LogDebugEvent, c.Index).
					Write("gadget_src", g.Src()).
					Write("gadget_typ", g.GadgetTyp())
				if g.GadgetTyp() == combat.GadgetTypDendroCore {
					// reactable.DendroCoreに型アサーションして種としてマーク
					if dc, ok := g.(*reactable.DendroCore); ok {
						dc.IsSeed = true
						// 爆発と反応トリガーを無効化
						dc.Gadget.OnExpiry = nil
						dc.Gadget.OnKill = nil

						// デバッグ用に変換をログ
						c.Core.Log.NewEvent(
							"nefer converted dendro core to seed",
							glog.LogElementEvent,
							c.Index,
						).Write("gadget_src", g.Src()).
							Write("is_seed", dc.IsSeed)
					} else {
						// アサーションが失敗した理由をデバッグするために予期しない具象型をログ
						c.Core.Log.NewEvent(
							"nefer conversion type mismatch",
							glog.LogElementEvent,
							c.Index,
						).Write("gadget_src", g.Src()).
							Write("concrete_type", reflect.TypeOf(g).String())
					}
				}
			}
		}

		// 2凸：元素スキル使用時に偽りのヴェールを2スタック獲得
		if c.Base.Cons >= 2 && c.Base.Ascension >= 1 {
			// 最大2スタック追加、上限5
			c.a1count = min(5.0, c.a1count+2)
		}
	}, skillHitmark)

	c.SetCDWithDelay(action.ActionSkill, 9*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap],
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	count := 2.0
	if c.Core.Rand.Float64() < 0.667 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Dendro, c.ParticleDelay)
}
