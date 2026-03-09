package sara

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var skillFrames []int

const (
	coverKey       = "sara-e-cover"
	particleICDKey = "sara-particle-icd"
	c2Hitmark      = 103
)

func init() {
	skillFrames = frames.InitAbilSlice(52) // E -> D
	skillFrames[action.ActionAttack] = 29  // E -> N1
	skillFrames[action.ActionAim] = 30     // E -> CA
	skillFrames[action.ActionBurst] = 32   // E -> Q
	skillFrames[action.ActionJump] = 51    // E -> J
	skillFrames[action.ActionSwap] = 50    // E -> Swap
}

// 元素スキルの処理。実際の処理の大半は他の場所で行われるため簡素な実装。
// クロウフェザー保護18秒間獲得。九條裟羅がフルチャージ狙い撃ちを放つと、
// クロウフェザー保護が消費され、着弾地点にクロウフェザーが残る。
// クロウフェザーは短時間後に天狗呉雷・待ち伏せを発動し、雷元素ダメージを与え、
// 九條裟羅の基礎攻撃力に基づく攻撃力バフを範囲内のアクティブキャラクターに付与する。
// 各種天狗呉雷の攻撃力バフは重複せず、最後に発動した天狗呉雷により決定される。
// 2凸も実装: 天狗召嗚発動時、元の位置に30%ダメージの弱いクロウフェザーを残す。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// クロウフェザーのスナップショットは発動時に取得される
	c.Core.Status.Add(coverKey, 18*60)

	// 2凸の処理
	if c.Base.Cons >= 2 {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Tengu Juurai: Ambush C2",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0.3 * skill[c.TalentLvlSkill()],
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)

		c.Core.QueueAttack(ai, ap, 50, c2Hitmark, c.makeA4CB())
		c.attackBuff(ap, c2Hitmark)
	}

	c.SetCDWithDelay(action.ActionSkill, 600, 7)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAttack], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.1*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Electro, c.ParticleDelay)
}

// 裟羅のスキルによる攻撃力バフの処理
// 遅延フレーム時のオンフィールドキャラクターをチェックし、そのキャラクターにバフを適用する
func (c *char) attackBuff(a combat.AttackPattern, delay int) {
	c.Core.Tasks.Add(func() {
		if collision, _ := c.Core.Combat.Player().AttackWillLand(a); !collision {
			return
		}

		active := c.Core.Player.ActiveChar()
		buff := atkBuff[c.TalentLvlSkill()] * c.Stat(attributes.BaseATK)

		c.Core.Log.NewEvent("sara attack buff applied", glog.LogCharacterEvent, c.Index).
			Write("char", active.Index).
			Write("buff", buff).
			Write("expiry", c.Core.F+360)

		m := make([]float64, attributes.EndStatType)
		m[attributes.ATK] = buff
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("sara-attack-buff", 360),
			AffectedStat: attributes.ATK,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// TODO: 後で単体攻撃に変更する
		c.Core.Player.Drain(info.DrainInfo{
			ActorIndex: active.Index,
			Abil:       "Tengu Juurai: Ambush",
			Amount:     0,
			External:   true,
		})

		if c.Base.Cons >= 1 {
			c.c1()
		}
		if c.Base.Cons >= 6 {
			c.c6(active)
		}
	}, delay)
}
