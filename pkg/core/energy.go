package core

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func (c *Core) QueueParticle(src string, num float64, ele attributes.Element, delay int) {
	p := character.Particle{
		Source: src,
		Num:    num,
		Ele:    ele,
	}
	if delay == 0 {
		c.Player.DistributeParticle(p)
		return
	}
	if delay < 0 {
		panic("queue particle called with delay < 0")
	}
	c.Tasks.Add(func() {
		c.Player.DistributeParticle(p)
	}, delay)
}

func (c *Core) SetupOnNormalHitEnergy() {
	var current [MaxTeamSize][info.EndWeaponClass]float64

	// https://genshin-impact.fandom.com/wiki/Energy#Energy_Generated_by_Normal_Attacks
	// 基本確率
	for i := range current {
		current[i][info.WeaponClassSword] = 0.10 // 片手剣
	}
	// 不発時の確率増加量
	inc := []float64{
		0.05, // 片手剣
		0.10, // 両手剣
		0.04, // 長柄武器
		0.05, // 弓
		0.10, // 法器
	}

	//TODO: 0.2秒のICDがあるか不明。安全のため暫定で追加
	icd := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		// ICD をチェック
		if icd > c.F {
			return false
		}
		// 確率をチェック
		char := c.Player.ByIndex(atk.Info.ActorIndex)

		if c.Rand.Float64() > current[atk.Info.ActorIndex][char.Weapon.Class] {
			// 確率を増加
			current[atk.Info.ActorIndex][char.Weapon.Class] += inc[char.Weapon.Class]
			return false
		}

		// エネルギーを追加
		char.AddEnergy("na-ca-on-hit", 1)
		// AddEnergy が既にログを生成するため、必要な場合のみこのログを出力
		c.Log.NewEvent("random energy on normal", glog.LogDebugEvent, char.Index).
			Write("char", atk.Info.ActorIndex).
			Write("chance", current[atk.Info.ActorIndex][char.Weapon.Class])
		// ICD をセット
		icd = c.F + 12
		current[atk.Info.ActorIndex][char.Weapon.Class] = 0
		if char.Weapon.Class == info.WeaponClassSword {
			current[atk.Info.ActorIndex][char.Weapon.Class] = 0.10
		}
		return false
	}, "random-energy-restore-on-hit")

	//TODO: 交代時に確率をリセットする想定
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		for i := range current {
			for j := range current[i] {
				current[i][j] = 0
			}
			current[i][info.WeaponClassSword] = 0.10
		}
		return false
	}, "random-energy-restore-on-hit-swap")
}
