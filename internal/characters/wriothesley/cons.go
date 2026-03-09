package wriothesley

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Status         = "wriothesley-c1"
	c1ICD            = 2.5 * 60
	c1ICDKey         = "wriothesley-c1-icd"
	c1SkillExtension = 4 * 60
	c4Status         = "wriothesley-c4-spd"
)

func (c *char) c1Ready() bool {
	return c.c1N5Proc || (c.CurrentHPRatio() < 0.6 && !c.StatusIsActive(c1ICDKey))
}

// 固有天賦「司罪の傅執」の「重裁の嘃き」が以下に変更される:
// リオセスリのHPが60%未満、または「氷牙の突進」による「氷牙の罰」状態の際、
// 「極寒の拳」の5段目が命中すると「重裁の嘃き」を獲得する。
// 2.5秒に1回獲得可能。
// さらに「重裁の嘃き：飛蹴」に以下の強化が加わる:
// - ダメージボーナスが200%に増加。
// - 「氷牙の罰」状態で命中時、その状態の持続時間が4秒延長。「氷牙の罰」1回につき1回延長可能。
// 固有天賦「司罪の傅執」を先に解放する必要がある。
func (c *char) c1(ai *combat.AttackInfo, snap *combat.Snapshot) (combat.AttackCBFunc, bool) {
	if !c.c1Ready() {
		return nil, false
	}
	c.c1N5Proc = false

	// 消費時に削除されるステータスを追加
	c.AddStatus(c1Status, -1, false)

	// AIを調整
	ai.Abil = "Rebuke: Vaulting Fist"
	ai.HitlagFactor = 0.03
	ai.HitlagHaltFrames = 0.12 * 60

	// ダメージ200%増加
	dmg := 2.0
	snap.Stats[attributes.DmgP] += dmg
	c.Core.Log.NewEvent("adding c1", glog.LogCharacterEvent, c.Index).Write("dmg%", dmg)

	// 該当する場合第6命ノ星座を追加
	c.addC6Buff(snap)

	// コールバックで回復、元素スキル延長、第1命ノ星座削除、2.5秒CD適用
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 第1命ノ星座がアクティブでない場合は発動しない
		if !c.StatusIsActive(c1Status) {
			return
		}
		// 第1命ノ星座を削除しCDを適用
		c.DeleteStatus(c1Status)
		c.AddStatus(c1ICDKey, c1ICD, true)

		// 元素スキル延長
		if !c.c1SkillExtensionProc && c.StatusIsActive(skillKey) {
			c.ExtendStatus(skillKey, c1SkillExtension)
			c.c1SkillExtensionProc = true
			c.Core.Log.NewEvent("c1: skill duration is extended", glog.LogCharacterEvent, c.Index)
		}

		// 回復
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "There Shall Be a Plea for Justice",
			Src:     c.caHeal * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}, c.Base.Cons >= 6
}

func (c *char) makeC1N5CB() combat.AttackCBFunc {
	if c.Base.Cons < 1 || c.NormalCounter != 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 元素スキルがアクティブかチェック
		if !c.StatusIsActive(skillKey) {
			return
		}
		// CDチェック
		if c.StatusIsActive(c1ICDKey) {
			return
		}
		c.c1N5Proc = true
		c.Core.Log.NewEvent("gained Gracious Rebuke from C1 N5", glog.LogCharacterEvent, c.Index)
	}
}

func (c *char) resetC1SkillExtension() {
	if c.Base.Cons < 1 {
		return
	}
	c.c1SkillExtensionProc = false
}

// 「暗金狼嚀」使用時、固有天賦「罪に対する清算」の「弾劾の勅令」スタックごとに
// 該当元素爆発のダメージが40%増加する。
// 固有天賦「罪に対する清算」を先に解放する必要がある。
func (c *char) c2(snap *combat.Snapshot) {
	if c.Base.Ascension < 4 {
		return
	}
	if c.Base.Cons < 2 {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}

	dmg := 0.4 * float64(c.a4Stack)
	snap.Stats[attributes.DmgP] += dmg
	c.Core.Log.NewEvent("adding c2", glog.LogCharacterEvent, c.Index).Write("dmg%", dmg)
}

// 「重裁の嘃き：飛蹴」によるリオセスリのHP回復量がHP上限の50%に増加。
// 固有天賦「司罪の傅執」を先に解放する必要がある。
// さらにリオセスリが回復を受けた際、回復量が超過した場合、
// フィールド上: 攻撃速度+20% 4秒。フィールド外: チーム全員の攻撃速度+10% 6秒。
// この2つの攻撃速度増加ၯ重複しない。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		c.caHeal = 0.3
		return
	}
	c.caHeal = 0.5

	c.Core.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		amount := args[2].(float64)
		overheal := args[3].(float64)
		if index != c.Index {
			return false
		}
		if amount <= 0 {
			return false
		}
		if overheal <= 0 {
			return false
		}

		chars := c.Core.Player.Chars()
		m := make([]float64, attributes.EndStatType)

		// 古いバフを削除
		for _, char := range chars {
			char.DeleteStatus(c4Status)
		}

		if c.Core.Player.Active() == c.Index {
			m[attributes.AtkSpd] = 0.2
			c.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(c4Status, 4*60),
				AffectedStat: attributes.AtkSpd,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		} else {
			m[attributes.AtkSpd] = 0.1
			for _, char := range chars {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(c4Status, 6*60),
					AffectedStat: attributes.AtkSpd,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
		}

		return false
	}, "wriothesley-c4-heal")
}

// 「重裁の嘃き：飛蹴」の会心率が10%、会心ダメージが80%増加。
func (c *char) addC6Buff(snap *combat.Snapshot) {
	if c.Base.Cons < 6 {
		return
	}
	cr := 0.1
	cd := 0.8
	snap.Stats[attributes.CR] += cr
	snap.Stats[attributes.CD] += cd
	c.Core.Log.NewEvent("adding c6", glog.LogCharacterEvent, c.Index).Write("cr", cr).Write("cd", cd)
}
