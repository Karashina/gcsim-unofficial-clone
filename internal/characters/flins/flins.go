package flins

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Flins, NewChar)
}

type char struct {
	*tmpl.Character
	northlandCD int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 5
	c.BurstCon = 3

	c.EnergyMax = 80
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// チーム初期化のためこのキャラクターをムーンサイン保持候補としてマーク
	c.AddStatus("moonsignKey", -1, false)
	c.InitLCallback()
	c.a0()
	c.a1()
	c.a4()
	c.c1()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// スキル/元素爆発のカスタムステータスがアクティブな場合、特殊バリアントを許可する。
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		if c.StatusIsActive(northlandCdKey) {
			// 北地の槍嵐は独自のCDを持ち、まだ使用できない。
			return false, action.SkillCD
		}
		// 通常の元素スキルCDがアクティブでもスキルバリアントを許可。
		return true, action.NoFailure
	}

	if a == action.ActionBurst && c.StatusIsActive(northlandKey) {
		if !c.Core.Flags.IgnoreBurstEnergy && c.Energy < 30 {
			// 雷鳴の交響曲にIgnoreBurstEnergyが設定されていない限り30エネルギーが必要。
			return false, action.InsufficientEnergy
		}
		// 通常の元素爆発CDがアクティブでも元素爆発バリアントを許可。
		return true, action.NoFailure
	}

	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "northlandup":
		return c.StatusIsActive(northlandCdKey), nil
	default:
		return c.Character.Condition(fields)
	}
}
