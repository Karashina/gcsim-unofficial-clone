package ayaka

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Ayaka, NewChar)
}

type char struct {
	*tmpl.Character
	wp             *ReactableWeapon
	c6CDTimerAvail bool // 6凸CDタイマー（0.5秒）が開始可能かどうかを制御するフラグ
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 80
	c.BurstCon = 3
	c.SkillCon = 5
	c.NormalHitNum = normalHitNum

	c.c6CDTimerAvail = false

	// 6凸能力がアクティブな状態で開始
	if c.Base.Cons >= 6 {
		c.c6CDTimerAvail = true
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// 6凸能力がアクティブな状態で開始
	if c.Base.Cons >= 6 {
		c.c6AddBuff()
	}
	c.WeaponReactionHandler()
	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionDash {
		f, ok := p["f"]
		if !ok {
			return 10 // タップ = 36フレーム、つまり1秒未満
		}
		// 1秒経過ごとに追加で15消費
		extra := f / 60
		return float64(10 + 15*extra)
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 7
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) WeaponReactionHandler() {
	c.Core.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		c.wp = c.newReactableWeapons()
		return false
	}, "ayaka-weaponreactionhandler-init")

	c.Core.Events.Subscribe(event.OnInfusion, func(args ...interface{}) bool {
		index := args[0].(int)
		ele := args[1].(attributes.Element)
		dur := args[2].(int)
		if c.Core.Player.ActiveChar().Index != c.Index {
			return false
		}
		if index != c.Index {
			return false
		}
		infai := combat.AttackInfo{
			ActorIndex: index,
			Abil:       "Weapon Infusion",
			Element:    ele,
			Durability: 25,
		}
		infae := combat.AttackEvent{
			Info:        infai,
			Pattern:     combat.NewSingleTargetHit(0),
			SourceFrame: c.Core.F,
		}
		c.wp.weaponreact(&infae)
		c.QueueCharTask(c.wp.resetgauge, dur)
		return false
	}, "ayaka-infusion")
}
