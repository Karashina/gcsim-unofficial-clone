package venti

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Venti, NewChar)
}

type char struct {
	*tmpl.Character
	qPos                geometry.Point
	qAbsorb             attributes.Element
	absorbCheckLocation combat.AttackPattern
	aiAbsorb            combat.AttackInfo
	snapAbsorb          combat.Snapshot
	// Hexereiモード（nohex=1が指定されない限りデフォルトtrue）
	isHexerei   bool
	hasHexBonus bool // パーティに2人以上のHexereiキャラ

	// Hexerei元素爆発の眼の追跡
	burstEnd       int // 元素爆発の眼が失効する絶対フレーム
	normalHexCount int // 元素爆発ごとのHex通常攻撃トリガー数（最大2）
	lastHexTrigger int // Hex通常攻撃トリガーが最後に発火したフレーム

	// 1凸六角術: Stormwind Arrow分裂追跡（0.25秒ICD = 15フレーム）
	lastStormwindSplit int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	// nohex=1が指定されない限りデフォルトはHexereiキャラクター
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// Hexereiボーナスを確認（パーティに2人以上のHexereiキャラ）
	c.checkHexereiBonus()

	// A0：Hexerei秘術パッシブ（拡散 → ダメージバフ + 元素爆発強化）
	c.a0HexereiInit()

	// 4凸（オリジナル）：粒子取得時にVentiが10秒間風元素ダメージ+25%を得る
	if c.Base.Cons >= 4 {
		c.c4Old()
	}

	// 6凸：元素爆発で影響を受けた敵に対する会心ダメージボーナスの永続AttackMod（Hexereiのみ）
	if c.Base.Cons >= 6 {
		c.c6AttackModInit()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}

// checkHexereiBonusはパーティに2人以上のHexereiキャラがいるか判定する。
func (c *char) checkHexereiBonus() {
	if !c.isHexerei {
		c.hasHexBonus = false
		return
	}
	hexereiCount := 0
	for _, ch := range c.Core.Player.Chars() {
		if result, err := ch.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	c.hasHexBonus = hexereiCount >= 2
}
