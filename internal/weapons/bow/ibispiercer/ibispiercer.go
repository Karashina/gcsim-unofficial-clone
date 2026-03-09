package ibispiercer

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
	core.RegisterWeaponFunc(keys.IbisPiercer, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 重撃が敵に命中した後6秒間、キャラクターの元素熟知が40/50/60/70/80増加する。
// 最大2スタック。この効果は0.5秒毎に1回発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	em := 30 + float64(r)*10

	m := make([]float64, attributes.EndStatType)

	stacks := 0
	maxStacks := 2
	const stackKey = "ibispiercer-stacks"
	stackDuration := 6 * 60
	const icdKey = "ibispiercer-icd"
	cd := int(0.5 * 60)

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		// 武器装備者からの攻撃かチェック
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		// 武器装備キャラクターがフィールド上にいるかチェック
		if c.Player.Active() != char.Index {
			return false
		}

		// 重撃でのみ適用
		if atk.Info.AttackTag != attacks.AttackTagExtra {
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

		// バフを追加
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(stackKey, stackDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				m[attributes.EM] = em * float64(stacks)
				return m, true
			},
		})

		return false
	}, fmt.Sprintf("ibispiercer-%v", char.Base.Key.String()))

	return w, nil
}
