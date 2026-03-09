package chiori

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

const (
	kinuDmgRatio       = 1.7
	kinuDuration       = int(3 * 60)
	kinuStartDelay     = int(0.6 * 60)
	kinuAttackInterval = int(0.5 * 60)
	kinuAttackDelay    = 5 // 0.08秒であるべき
)

// 「絹」は付近の敵を攻撃し、Tamotoの170%に相当する
// 岩元素範囲ダメージを与える。このダメージは元素スキルダメージ扱い。
//
// 「絹」は1回攻撃後または3秒経過後に退場する。
func (c *char) createKinu(src int, centerOffset, minRandom, maxRandom float64) func() {
	return func() {
		// 「絹」の位置を決定
		player := c.Core.Combat.Player()
		center := geometry.CalcOffsetPoint(
			player.Pos(),
			geometry.Point{Y: centerOffset},
			player.Direction(),
		)
		kinuPos := geometry.CalcRandomPointFromCenter(center, minRandom, maxRandom, c.Core.Rand)

		c.Core.Log.NewEvent("kinu spawned", glog.LogCharacterEvent, c.Index).Write("src", src)

		// 「絹」を生成
		kinu := newTicker(c.Core, kinuDuration, nil)
		kinu.cb = c.kinuAttack(src, kinu, kinuPos)
		kinu.interval = kinuAttackInterval
		c.Core.Tasks.Add(kinu.tick, kinuStartDelay)
		c.kinus = append(c.kinus, kinu)
	}
}

func (c *char) kinuAttack(src int, kinu *ticker, pos geometry.Point) func() {
	return func() {
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				Abil:       "Fluttering Hasode (Kinu)",
				ActorIndex: c.Index,
				AttackTag:  attacks.AttackTagElementalArt,
				ICDTag:     attacks.ICDTagChioriSkill,
				ICDGroup:   attacks.ICDGroupChioriSkill,
				StrikeType: attacks.StrikeTypeBlunt,
				PoiseDMG:   0,
				Element:    attributes.Geo,
				Durability: 25,
				Mult:       turretAtkScaling[c.TalentLvlSkill()] * kinuDmgRatio,
			}

			snap := c.Snapshot(&ai)
			ai.FlatDmg = snap.Stats.TotalDEF()
			ai.FlatDmg *= turretDefScaling[c.TalentLvlSkill()] * kinuDmgRatio

			// プレイヤーに攻撃ターゲットがある場合は常にこの敵を選択する
			// 検索AoE内にあることを確認するだけでよい
			t := c.Core.Combat.PrimaryTarget()
			if !t.IsWithinArea(combat.NewCircleHitOnTarget(pos, nil, c.skillSearchAoE)) {
				return
			}

			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(t, nil, skillDollAoE), 0)

			c.Core.Log.NewEvent("kinu killed on attack", glog.LogCharacterEvent, c.Index).Write("src", src)

			kinu.kill()
			c.cleanupKinu()
		}, kinuAttackDelay)
	}
}

func (c *char) cleanupKinu() {
	n := 0
	for _, t := range c.kinus {
		if t.alive {
			c.kinus[n] = t
			n++
		}
	}
	c.kinus = c.kinus[:n]
}
