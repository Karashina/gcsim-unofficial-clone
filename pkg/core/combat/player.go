package combat

import "github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"

func (h *Handler) SetPlayer(t Target) {
	h.player = t
	t.SetKey(0)
}

func (h *Handler) Player() Target {
	return h.player
}

func (h *Handler) SetPlayerPos(p geometry.Point) {
	h.player.SetPos(p)
}
