// Package enemy は敵ターゲットを実装する
package enemy

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/target"
)

type Enemy struct {
	*target.Target
	*reactable.Reactable

	Level   int
	resists map[attributes.Element]float64
	prof    info.EnemyProfile
	hp      float64
	maxhp   float64

	damageTaken       float64
	lastParticleDrop  int
	particleDropIndex int // カスタムHPドロップ用

	// 修飾子
	mods []modifier.Mod

	// ヒットラグ関連
	timePassed   int
	frozenFrames int
	queue        *task.Handler
}

func New(core *core.Core, p info.EnemyProfile) *Enemy {
	e := &Enemy{}
	e.queue = task.New(&e.timePassed)
	e.Level = p.Level
	//TODO: 代わりにこのマップをクローンする必要があるか？
	e.resists = p.Resist
	//TODO: プロファイルとレベル/耐性の両方を保持するのは冗長
	e.prof = p
	e.Target = target.New(core, geometry.Point{X: p.Pos.X, Y: p.Pos.Y}, p.Pos.R)
	e.Reactable = &reactable.Reactable{}
	e.Reactable.Init(e, core)
	e.Reactable.FreezeResist = e.prof.FreezeResist
	e.mods = make([]modifier.Mod, 0, 10)
	if core.Combat.DamageMode {
		e.hp = p.HP
		e.maxhp = p.HP
	}
	return e
}

func (e *Enemy) Type() targets.TargettableType { return targets.TargettableEnemy }

func (e *Enemy) MaxHP() float64 { return e.maxhp }
func (e *Enemy) HP() float64    { return e.hp }
func (e *Enemy) Kill() {
	e.Alive = false
	if e.Key() == e.Core.Combat.DefaultTarget {
		player := e.Core.Combat.Player()
		// ターゲットが死亡した場合、プレイヤーに最も近い敵をデフォルトターゲットに設定
		enemy := e.Core.Combat.ClosestEnemy(player.Pos())
		if enemy == nil {
			// 全敵が死亡、現時点では何もしない
			return
		}
		e.Core.Combat.DefaultTarget = enemy.Key()
		e.Core.Combat.Log.NewEvent("default target changed on enemy death", glog.LogWarnings, -1)
		player.SetDirection(enemy.Pos())
	}
}

func (e *Enemy) SetDirection(trg geometry.Point) {}
func (e *Enemy) SetDirectionToClosestEnemy()     {}
func (e *Enemy) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}
