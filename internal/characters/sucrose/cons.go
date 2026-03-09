package sucrose

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// C4処理: 通常攻撃と重撃の7回ごとに、風霊作成・六三〇八のCDを1～7秒短縮する
func (c *char) makeC4Callback() func(combat.AttackCB) {
	if c.Base.Cons < 4 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		c.c4Count++
		if c.c4Count < 7 {
			return
		}
		c.c4Count = 0

		// 変更は浮動小数点で可能。例としてTerrapinの動画を参照
		// https://youtu.be/jB3x5BTYWIA?t=20
		cdReduction := 60 * int(c.Core.Rand.Float64()*6+1)

		// アクションCDを単純に減少
		c.ReduceActionCooldown(action.ActionSkill, cdReduction)

		c.Core.Log.NewEvent("sucrose c4 reducing E CD", glog.LogCharacterEvent, c.Index).
			Write("cd_reduction", cdReduction)
	}
}

func (c *char) c6() {
	stat := attributes.EleToDmgP(c.qAbsorb)
	c.c6buff[stat] = .20

	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("sucrose-c6", 60*10),
			AffectedStat: stat,
			Amount: func() ([]float64, bool) {
				return c.c6buff, true
			},
		})
	}
}
