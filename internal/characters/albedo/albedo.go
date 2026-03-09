package albedo

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Albedo, NewChar)
}

type char struct {
	*tmpl.Character
	lastConstruct int
	bloomSnapshot combat.Snapshot
	// スキル情報を追跡
	skillActive     bool
	skillArea       combat.AttackPattern
	skillAttackInfo combat.AttackInfo
	skillSnapshot   combat.Snapshot
	// 2凸の追跡
	c2stacks int
	// Hexereiモード（nohex=1が指定されない限りデフォルトtrue）
	isHexerei bool
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 40
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	// nohex=1が指定されない限りデフォルトはHexereiキャラクター
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.skillHook()
	c.a1()

	// 2凸追加効果: フィールド外で4スタック時にFatal Blossomを自動発動
	if c.Base.Cons >= 2 {
		c.c2AutoBlossom()
	}

	// 4凸Hexerei効果: 落下攻撃時にSilver Isotomaを破壊
	if c.Base.Cons >= 4 && c.isHexerei {
		c.c4HexereiJumpBuff()
	}

	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "elevator":
		return c.skillActive, nil
	case "c2stacks":
		return c.c2stacks, nil
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}
