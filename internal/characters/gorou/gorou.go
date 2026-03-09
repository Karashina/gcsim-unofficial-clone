package gorou

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

const (
	defenseBuffKey           = "gorou-e-defbuff"
	generalWarBannerKey      = "gorou-e-warbanner"
	generalGloryKey          = "gorou-q-glory"
	generalWarBannerDuration = 600    // 10s
	generalGloryDuration     = 9 * 60 // 9s, dm says 9.1s but that would mean you get an extra Crystal Collapse tick so it's staying at 9s
	a1Key                    = "gorou-a1"
	c6key                    = "gorou-c6"
)

func init() {
	core.RegisterCharFunc(keys.Gorou, NewChar)
}

type char struct {
	*tmpl.Character
	eFieldArea     combat.AttackPattern
	eFieldSrc      int
	qFieldSrc      int
	gorouBuff      []float64
	geoCharCount   int
	c2Extension    int
	c6Buff         []float64
	a1Buff         []float64
	healFieldStats attributes.Stats
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 80
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	c.c6Buff = make([]float64, attributes.EndStatType)
	c.gorouBuff = make([]float64, attributes.EndStatType)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a1Buff = make([]float64, attributes.EndStatType)
	c.a1Buff[attributes.DEFP] = .25

	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element == attributes.Geo {
			c.geoCharCount++
		}
	}

	/**
	パーティ内の岩元素キャラクターの数に応じて、スキルのAoE内のアクティブキャラクターに最大3つのバフを付与（発動時の人数で決定）:
	• 岩元素キャラ1人: 「積石」追加 - 防御力ボーナス。
	• 岩元素キャラ2人: 「集岩」追加 - 中断耐性向上。
	• 岩元素キャラ3人: 「碎岩」追加 - 岩元素ダメージボーナス。
	**/
	c.gorouBuff[attributes.DEF] = skillDefBonus[c.TalentLvlSkill()]
	if c.geoCharCount > 2 {
		c.gorouBuff[attributes.GeoP] = 0.15 // 岩元素ダメージ15%
	}

	/**
	犬坂鐌繰の昭もしくは戦陣の誉を使用してから12秒間、
	使用時のスキルフィールドのバフレベルに応じて、
	付近の全パーティメンバーの岩元素ダメージの会心ダメージが増加:
	• 「積石」: +10%
	• 「集岩」: +20%
	• 「碎岩」: +40%
	この効果は重複せず、最後に発動したインスタンスを参照する。
	**/
	switch c.geoCharCount {
	case 1:
		c.c6Buff[attributes.CD] = 0.1
	case 2:
		c.c6Buff[attributes.CD] = 0.2
	default:
		// 1未満にはならないので3人以上
		c.c6Buff[attributes.CD] = 0.4
	}

	if c.Base.Cons > 0 {
		c.c1()
	}
	if c.Base.Cons >= 2 {
		c.c2()
	}

	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}
