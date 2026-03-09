package compound

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.CompoundBow, NewWeapon)
}

/*
* Normal Attack and Charged Attack hits increase ATK by 4/5/6/7/8% and Normal ATK SPD by
* 1.2/1.5/1.8/2.1/2.4% for 6s. Max 4 stacks. Can only occur once every 0.3s.
 */
type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)

	incAtk := .03 + float64(r)*0.01
	incSpd := 0.009 + float64(r)*0.003

	stacks := 0
	maxStacks := 4
	const stackKey = "compoundbow-stacks"
	stackDuration := 360 // フレーム数 = 6秒 × 60fps
	const icdKey = "compoundbow-icd"

	cd := 18 // フレーム数 = 0.3秒 × 60fps

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		// 武器装備者からの攻撃かチェック
		if c.Player.Active() != char.Index {
			return false
		}

		// 通常攻撃または重撃でのみ適用
		if (atk.Info.AttackTag != attacks.AttackTagNormal) && (atk.Info.AttackTag != attacks.AttackTagExtra) {
			return false
		}

		// クールダウン中かチェック
		if char.StatusIsActive(icdKey) {
			return false
		}

		// スタックが期限切れの場合リセット
		if !char.StatusIsActive(stackKey) {
			stacks = 0
		}

		// チェック完了、武器パッシブを発動
		// スタック数を増加
		if stacks < maxStacks {
			stacks++
		}

		// クールダウンを発動
		char.AddStatus(icdKey, cd, true)
		char.AddStatus(stackKey, stackDuration, true)

		// バフ持続時間 6 × 60 = 360フレーム
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("compoundbow", stackDuration),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				m[attributes.ATKP] = incAtk * float64(stacks)
				m[attributes.AtkSpd] = incSpd * float64(stacks)
				return m, true
			},
		})

		return false
	}, fmt.Sprintf("compoundbow-%v", char.Base.Key.String()))

	return w, nil
}
