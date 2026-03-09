package lynette

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// TODO: 第1命ノ星座は未実装。渦巻/引き寄せメカニクスが実装されていないため

// 「魔術・アストニッシングシフト」で召喚された「ビックリボンボックス」が「ヴィヴィッドショット」を発射する際、追加でもう1発発射する。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.vividCount = 2
}

// 「エニグマティックフェイント」の使用回数を1増やす。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	c.SetNumCharges(action.ActionSkill, 2)
}

// リネットが「エニグマティックフェイント」の「エニグマスラスト」を使用すると、風元素付与と風元素ダメージボーナス+20%を6秒間獲得する。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	duration := int((6 + 0.4) * 60)

	// 風元素付与を追加
	c.Core.Player.AddWeaponInfuse(
		c.Index,
		"lynette-c6-infusion",
		attributes.Anemo,
		duration,
		true,
		attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
	)
	c.Core.Events.Emit(event.OnInfusion, c.Index, attributes.Anemo, duration)

	// 風元素ダメージボーナスバフを追加
	m := make([]float64, attributes.EndStatType)
	m[attributes.AnemoP] = 0.2
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("lynette-c6-buff", duration),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}
