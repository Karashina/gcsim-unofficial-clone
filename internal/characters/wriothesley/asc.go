package wriothesley

import (
	"math"

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
	a1Status = "wriothesley-a1"
	a1ICD    = 5 * 60
	a1ICDKey = "wriothesley-a1-icd"
)

func (c *char) a1Ready() bool {
	return c.CurrentHPRatio() < 0.6 && !c.StatusIsActive(a1ICDKey)
}

// リオセスリのHPが60%未満の時、「重裁の嘃き」を獲得する。次の通常攻撃：極寒の拳の重撃が
// 「重裁の嘃き：飛蹴」に強化される。スタミナを消費せず、ダメージが50%増加し、
// 命中後にリオセスリのHPをHP上限の30%分回復する。
// この方法で「重裁の嘃き」を5秒に1回獲得可能。
func (c *char) a1(ai *combat.AttackInfo, snap *combat.Snapshot) combat.AttackCBFunc {
	if !c.a1Ready() {
		return nil
	}

	// 消費時に削除されるステータスを追加
	c.AddStatus(a1Status, -1, false)

	// AIを調整
	ai.Abil = "Rebuke: Vaulting Fist"
	ai.HitlagFactor = 0.03
	ai.HitlagHaltFrames = 0.12 * 60

	// ダメージ50%増加
	dmg := 0.5
	snap.Stats[attributes.DmgP] += dmg
	c.Core.Log.NewEvent("adding a1", glog.LogCharacterEvent, c.Index).Write("dmg%", dmg)

	// コールバックで回復、固有天賦1を削除、5秒CDを適用
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 固有天賦1がアクティブでない場合は発動しない
		if !c.StatusIsActive(a1Status) {
			return
		}
		// 固有天賦1を削除しCDを適用
		c.DeleteStatus(a1Status)
		c.AddStatus(a1ICDKey, a1ICD, true)

		// 回復
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "There Shall Be a Plea for Justice",
			Src:     c.caHeal * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}

// リオセスリの現在HPが増減した際、「氷牙の突進」による「氷牙の罰」状態の場合、
// 「弾劾の勅令」スタックを1つ獲得する。最大5スタック。各スタックでリオセスリの攻撃力が6%増加。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if di.ActorIndex != c.Index {
			return false
		}
		if di.Amount <= 0 {
			return false
		}

		if c.StatusIsActive(skillKey) && c.a4Stack < 5 {
			c.a4Stack++
			c.Core.Log.NewEvent("a4 gained stack", glog.LogCharacterEvent, c.Index).Write("stacks", c.a4Stack)
		}
		return false
	}, "wriothesley-a4-drain")

	c.Core.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		amount := args[2].(float64)
		overheal := args[3].(float64)
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if index != c.Index {
			return false
		}
		if amount <= 0 {
			return false
		}
		// 既にHP最大の場合は発動しない
		if math.Abs(amount-overheal) <= 1e-9 {
			return false
		}

		if c.StatusIsActive(skillKey) && c.a4Stack < 5 {
			c.a4Stack++
			c.Core.Log.NewEvent("a4 gained stack", glog.LogCharacterEvent, c.Index).Write("stacks", c.a4Stack)
		}
		return false
	}, "wriothesley-a4-heal")
}

func (c *char) applyA4(dur int) {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("wriothesley-a4", dur),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			m[attributes.ATKP] = float64(c.a4Stack) * 0.06
			return m, true
		},
	})
}

func (c *char) resetA4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4Stack = 0
}
