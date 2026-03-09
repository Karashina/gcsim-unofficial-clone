package mizuki

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1SwirlKey                    = "mizuki-a1-swirl-%v"
	a1ICDKey                      = "mizuki-a1-icd"
	a1ICD                         = 0.3 * 60
	a4Duration                    = 4 * 60
	a4Key                         = "mizuki-a4"
	a4ICDKey                      = "mizuki-a4-icd"
	a4ICD                         = 0.3 * 60
	a4EMBuff                      = 100
	dreamDrifterExtensions        = 2
	dreamDrifterDurationExtension = 2.5 * 60
)

// 夢見月瑞希がDreamdrifter状態中に拡散を発動した時、Dreamdrifterの持続時間が2.5秒延長される。
// この効果は0.3秒に1回、各Dreamdrifter状態につき最大2回発動可能。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	swirlFunc := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}

		atk := args[1].(*combat.AttackEvent)

		// 瑞希が拡散を発動する必要がある
		if atk.Info.ActorIndex != c.Index {
			return false
		}

		// Dreamdrifterがアクティブな時のみ
		if !c.StatusIsActive(dreamDrifterStateKey) {
			return false
		}

		// スキル1回につき最大2回延長
		if c.dreamDrifterExtensionsRemaining <= 0 {
			return false
		}

		// ICD
		if c.StatusIsActive(a1ICDKey) {
			return false
		}

		c.AddStatus(a1ICDKey, a1ICD, true)

		c.ExtendStatus(dreamDrifterStateKey, dreamDrifterDurationExtension)

		c.dreamDrifterExtensionsRemaining--

		return false
	}

	c.Core.Events.Subscribe(event.OnSwirlPyro, swirlFunc, fmt.Sprintf(a1SwirlKey, attributes.Pyro))
	c.Core.Events.Subscribe(event.OnSwirlHydro, swirlFunc, fmt.Sprintf(a1SwirlKey, attributes.Hydro))
	c.Core.Events.Subscribe(event.OnSwirlElectro, swirlFunc, fmt.Sprintf(a1SwirlKey, attributes.Electro))
	c.Core.Events.Subscribe(event.OnSwirlCryo, swirlFunc, fmt.Sprintf(a1SwirlKey, attributes.Cryo))
}

// 夢見月瑞希がDreamdrifter状態中、周囲の他のパーティメンバーが炎・水・氷・雷攻撃で敵に命中した時、
// 彼女の元素熟知が4秒間100増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	c.a4Buff = make([]float64, attributes.EndStatType)
	c.a4Buff[attributes.EM] = a4EMBuff

	hitFunc := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}

		// 他キャラの攻撃時のみ
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == c.Index {
			return false
		}

		// Dreamdrifterがアクティブな時のみ
		if !c.StatusIsActive(dreamDrifterStateKey) {
			return false
		}

		// ICD
		if c.StatusIsActive(a4ICDKey) {
			return false
		}
		c.AddStatus(a4ICDKey, a4ICD, true)

		// 敵が炎・水・氷・雷でヒットされた時のみ
		switch atk.Info.Element {
		case attributes.Electro:
		case attributes.Hydro:
		case attributes.Pyro:
		case attributes.Cryo:
		default:
			return false
		}

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a4Key, a4Duration), // 4s
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return c.a4Buff, true
			},
		})

		return false
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, hitFunc, a4Key)
}
