package talkingstick

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/weapons/common"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

// 炎元素の影響を受けた後、攻撃力が15秒間16%増加。この効果は12秒に1回発動可能。
// 水元素、氷元素、雷元素、草元素の影響を受けた後、全元素ダメージバフが15秒間12%増加。
// この効果は12秒に1回発動可能。
// TODO: https://github.com/Karashina/gcsim-unofficial-clone/issues/850
func init() {
	core.RegisterWeaponFunc(keys.TalkingStick, NewWeapon)
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := common.NewNoEffect(base)
	return w.NewWeapon(c, char, p)
}
