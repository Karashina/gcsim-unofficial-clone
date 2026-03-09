package fischl

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/internal/template/minazuki"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Fischl, NewChar)
}

type char struct {
	*tmpl.Character
	// オズのダメージ計算用のフィールド
	ozPos           geometry.Point
	ozSnapshot      combat.AttackEvent
	ozSource        int  // オズのソースを追跡（リセット用）
	ozActive        bool // GCSL条件判定専用
	ozTickSrc       int  // オズの再召喚攻撃用
	ozTravel        int
	burstOzSpawnSrc int // 元素爆発からのオズ二重召喚を防止
	c6Watcher       *minazuki.Watcher
	// Hexereiモード（nohex=1が指定されない限りデフォルトtrue）
	isHexerei bool
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.ozSource = -1
	c.ozActive = false
	c.ozTickSrc = -1

	c.ozTravel = 10
	travel, ok := p.Params["oz_travel"]
	if ok {
		c.ozTravel = travel
	}

	// nohex=1が指定されない限りデフォルトはHexereiキャラクター
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()
	c.hexPassive()
	c.hexC6Boost()

	if c.Base.Cons >= 6 {
		w, err := minazuki.New(
			minazuki.WithMandatory(keys.Fischl, "fischl c6", ozActiveKey, "", 60, c.c6Wave, c.Core),
			minazuki.WithTickOnActive(true),
			minazuki.WithAnimationDelayCheck(model.AnimationYelanN0StartDelay, func() bool {
				return c.Core.Player.ActiveChar().NormalCounter == 1
			}),
		)
		if err != nil {
			return err
		}
		c.c6Watcher = w
	}
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "oz":
		return c.ozActive, nil
	case "oz-source":
		return c.ozSource, nil
	case "oz-duration":
		return c.StatusDuration(ozActiveKey), nil
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// オズの再召喚が可能かチェック
	if a == action.ActionSkill && p["recast"] != 0 && c.ozActive {
		return !c.StatusIsActive(skillRecastCDKey), action.SkillCD
	}
	// スキルでオズがフィールド上にいる時に発動判定
	if a == action.ActionSkill && c.ozActive {
		return false, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}
