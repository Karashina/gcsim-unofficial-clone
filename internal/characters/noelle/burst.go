package noelle

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

// TODO: これが正しいか不明
const (
	burstStart   = 80
	burstBuffKey = "noelle-burst"
)

func init() {
	burstFrames = frames.InitAbilSlice(121)
	burstFrames[action.ActionAttack] = 83
	burstFrames[action.ActionCharge] = 82
	burstFrames[action.ActionSkill] = 82
	burstFrames[action.ActionDash] = 81
	burstFrames[action.ActionJump] = 81
	burstFrames[action.ActionWalk] = 90
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// TODO: 変換が元素爆発の2ヒットにもバフを与えるため、発動時に即座にスナップショットされると仮定
	// デバッグで適用中のモッド一覧を表示するための「仮」スナップショットを生成
	aiSnapshot := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Sweeping Time (Stat Snapshot)",
	}
	c.Snapshot(&aiSnapshot)
	burstDefSnapshot := c.TotalDef(true)
	mult := defconv[c.TalentLvlBurst()]
	if c.Base.Cons >= 6 {
		mult += 0.5
	}
	// 防御力→攻撃力変換のモッドを追加
	c.burstBuff[attributes.ATK] = mult * burstDefSnapshot

	dur := 900 + burstStart // デフォルト持続時間
	if c.Base.Cons >= 6 {
		// https://library.keqingmains.com/evidence/characters/geo/noelle#noelle-c6-burst-extension
		// 延長を確認
		getExt := func() int {
			ext, ok := p["extend"]
			if !ok {
				return 10 // 以前のデフォルト動作（フル延長）を維持するため
			}
			if ext < 0 {
				ext = 0
			}
			if ext > 10 {
				ext = 10
			}
			return ext
		}

		ext := getExt()
		dur += ext * 60
		c.Core.Log.NewEvent("noelle c6 extension applied", glog.LogCharacterEvent, c.Index).
			Write("total_dur", dur).
			Write("ext", ext)
	}
	// TODO: バフの正確なタイミングを確認。現在は以前設定されたステータス持続時間（900+アニメーションフレーム）に合わせている
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("noelle-burst", dur),
		AffectedStat: attributes.ATK,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return c.burstBuff, true
		},
	})
	c.Core.Log.NewEvent("noelle burst", glog.LogSnapshotEvent, c.Index).
		Write("total def", burstDefSnapshot).
		Write("atk added", c.burstBuff[attributes.ATK]).
		Write("mult", mult)

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Sweeping Time (Burst)",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagElementalBurst,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150,
		Element:            attributes.Geo,
		Durability:         25,
		Mult:               burst[c.TalentLvlBurst()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.15 * 60,
		CanBeDefenseHalted: true,
	}

	// 元素爆発部分
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6.5),
			0,
			0,
			c.skillHealCB(),
		)
	}, 24)

	// 元素スキル部分
	// 元素爆発とスキル部分は同じヒットラグ値で、両方とも回復可能
	c.QueueCharTask(func() {
		ai.Abil = "Sweeping Time (Skill)"
		ai.Mult = burstskill[c.TalentLvlBurst()]
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4),
			0,
			0,
			c.skillHealCB(),
		)
	}, 65)

	c.SetCD(action.ActionBurst, 900)
	c.ConsumeEnergy(8)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
