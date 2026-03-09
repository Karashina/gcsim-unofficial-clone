package xiao

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Xiao, NewChar)
}

// 魉固有のキャラクター実装
type char struct {
	*tmpl.Character
	qStarted int
	a4stacks int
	a4buff   []float64
	c6Count  int
}

// キャラクターを初期化
// TODO: 4凸は未実装 - 防御力はそこまで重要ではない
func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 70
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum

	c.c6Count = 0

	c.SetNumCharges(action.ActionSkill, 2)
	if c.Base.Cons >= 1 {
		c.SetNumCharges(action.ActionSkill, 3)
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4buff = make([]float64, attributes.EndStatType)
	c.onExitField()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4()
	}
	return nil
}

// 魉固有の Snapshot 実装（元素爆発ボーナス用）。胡桃と同様。
// 元素爆発の風元素攻撃変換とダメージボーナスを実装。
// 固有天賦1も実装:
// 「妞降」の効果中、全ダメージ+5%。3秒ごとにさらに+5%。最大25%。
func (c *char) Snapshot(a *combat.AttackInfo) combat.Snapshot {
	ds := c.Character.Snapshot(a)

	if c.StatusIsActive("xiaoburst") {
		// 通常攻撃、重撃、落下攻撃への風元素変換とダメージボーナス適用
		// 元素爆発中の重撃ICD変更も処理（通常攻撃と共有）
		switch a.AttackTag {
		case attacks.AttackTagNormal:
			// 元素爆発中のN1-1は通常N1-1と異なるヒットラグ
			if a.Abil == "Normal 0" {
				// N1-2のHitlagHaltFramesも上書きされるが、同じ値なので問題ない
				a.HitlagHaltFrames = 0.01 * 60
			}
		case attacks.AttackTagExtra:
			// 元素爆発中の重撃は通常重撃と異なるヒットラグ
			a.ICDTag = attacks.ICDTagNormalAttack
			a.HitlagHaltFrames = 0.04 * 60
		case attacks.AttackTagPlunge:
		default:
			return ds
		}
		a.Element = attributes.Anemo
		bonus := burstBonus[c.TalentLvlBurst()]
		ds.Stats[attributes.DmgP] += bonus
		c.Core.Log.NewEvent("xiao burst damage bonus", glog.LogCharacterEvent, c.Index).
			Write("bonus", bonus).
			Write("final", ds.Stats[attributes.DmgP])
	}
	return ds
}
