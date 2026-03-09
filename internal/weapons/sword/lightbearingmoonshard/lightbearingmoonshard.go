package lightbearingmoonshard

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.LightbearingMoonshard, NewWeapon)
}

// ライトベアリング・ムーンシャード
// 片手剣、★4
// 基礎攻撃力: 44
// サブステータス: 会心ダメージ 19.2% (Lv90時)
//
// パッシブ: 月宮の輝き
// 防御力 +20/25/30/35/40%（永続パッシブ）
// 元素スキル使用時:
// - Lunar-Crystallize (LCrs) 反応ダメージ +64/80/96/112/128%（5秒間）

const (
	defKey       = "lightbearingmoonshard-def"
	lcrsBonusKey = "lightbearingmoonshard-lcrs"
	lcrsBonusDur = 5 * 60 // 5s
)

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}

	r := p.Refine

	// 精錬ランク別の防御力%ボーナス（永続パッシブ）
	defBonus := []float64{0.20, 0.25, 0.30, 0.35, 0.40}
	// 精錬ランク別のLCrsダメージボーナス
	lcrsBonus := []float64{0.64, 0.80, 0.96, 1.12, 1.28}

	// W-2: 防御力%は永続パッシブであり、スキルトリガーではない
	mDef := make([]float64, attributes.EndStatType)
	mDef[attributes.DEFP] = defBonus[r-1]
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(defKey, -1),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return mDef, true
		},
	})

	// W-1/W-3: 元素スキル使用時、LCrsReactBonusMod経由でLCrsダメージボーナスを5秒間付与
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lcrsBonusKey, lcrsBonusDur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return lcrsBonus[r-1], false
			},
		})

		c.Log.NewEvent("Lightbearing Moonshard LCrs bonus activated", glog.LogWeaponEvent, char.Index).
			Write("def_bonus", defBonus[r-1]).
			Write("lcrs_bonus", lcrsBonus[r-1]).
			Write("lcrs_duration", lcrsBonusDur)

		return false
	}, "lightbearingmoonshard-skill-"+char.Base.Key.String())

	return w, nil
}
