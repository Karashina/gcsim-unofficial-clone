package ifa

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		// TODO: ここでのナイトソウル消費方法とスキルキャンセル/落下攻撃タイミングへの影響を調べる
		return action.Info{}, errors.New("ifa jump in nightsoul blessing not implemented")
	}
	return c.Character.Jump(p)
}
