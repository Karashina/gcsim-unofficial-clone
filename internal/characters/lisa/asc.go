package lisa

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 重撃が命中すると、蓒花の雷光の導電状態を敵に付与する。
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		t, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		count := t.GetTag(conductiveTag)
		if count < 3 {
			t.SetTag(conductiveTag, count+1)
		}
	}
}

// 薔薇の雷光に命中された敵の防御力が10秒間15%低下する。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		t, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		t.AddDefMod(combat.DefMod{
			Base:  modifier.NewBaseWithHitlag("lisa-a4", 600),
			Value: -0.15,
		})
	}
}
