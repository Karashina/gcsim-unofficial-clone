package chasca

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		// TODO: ここでのナイトソウル消費方法と、スキルキャンセル/落下攻撃タイミングへの影響を調査
		return action.Info{}, errors.New("chasca jump in nightsoul blessing not implemented")
	}
	return c.Character.Jump(p)
}
