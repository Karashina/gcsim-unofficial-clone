package combat

import (
	"sort"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

// 全ターゲット

func enemiesWithinAreaFiltered(a AttackPattern, filter func(t Enemy) bool, originalEnemies []Target) []Enemy {
	var enemies []Enemy
	hasFilter := filter != nil
	for _, v := range originalEnemies {
		e, ok := v.(Enemy)
		if !ok {
			panic("enemies should contain targets that implement the Enemy interface")
		}
		if hasFilter && !filter(e) {
			continue
		}
		if !v.IsAlive() {
			continue
		}
		if !e.IsWithinArea(a) {
			continue
		}
		enemies = append(enemies, e)
	}
	return enemies
}

func gadgetsWithinAreaFiltered(a AttackPattern, filter func(t Gadget) bool, originalGadgets []Gadget) []Gadget {
	var gadgets []Gadget
	hasFilter := filter != nil
	for _, v := range originalGadgets {
		if v == nil {
			continue
		}
		// ガジェットが敵陣営かチェック、味方ガジェットはアビリティの対象にならない
		if !(v.GadgetTyp() > StartGadgetTypEnemy && v.GadgetTyp() < EndGadgetTypEnemy) {
			continue
		}
		if hasFilter && !filter(v) {
			continue
		}
		if !v.IsAlive() {
			continue
		}
		if !v.IsWithinArea(a) {
			continue
		}
		gadgets = append(gadgets, v)
	}
	return gadgets
}

// 指定範囲内の敵を返す。ソートなし。フィルター不要の場合はnilを渡す
func (h *Handler) EnemiesWithinArea(a AttackPattern, filter func(t Enemy) bool) []Enemy {
	enemies := enemiesWithinAreaFiltered(a, filter, h.enemies)
	if len(enemies) == 0 {
		return nil
	}
	return enemies
}

// 指定範囲内のガジェットを返す。ソートなし。フィルター不要の場合はnilを渡す
func (h *Handler) GadgetsWithinArea(a AttackPattern, filter func(t Gadget) bool) []Gadget {
	gadgets := gadgetsWithinAreaFiltered(a, filter, h.gadgets)
	if len(gadgets) == 0 {
		return nil
	}
	return gadgets
}

// ランダムターゲット

// 指定範囲内のランダムな敵を返す。フィルター不要の場合はnilを渡す
func (h *Handler) RandomEnemyWithinArea(a AttackPattern, filter func(t Enemy) bool) Enemy {
	enemies := h.EnemiesWithinArea(a, filter)
	if enemies == nil {
		return nil
	}
	return enemies[h.Rand.Intn(len(enemies))]
}

// 指定範囲内のランダムなガジェットを返す。フィルター不要の場合はnilを渡す
func (h *Handler) RandomGadgetWithinArea(a AttackPattern, filter func(t Gadget) bool) Gadget {
	gadgets := h.GadgetsWithinArea(a, filter)
	if gadgets == nil {
		return nil
	}
	return gadgets[h.Rand.Intn(len(gadgets))]
}

// 指定範囲内のランダムな敵のリストを返す。フィルター不要の場合はnilを渡す
func (h *Handler) RandomEnemiesWithinArea(a AttackPattern, filter func(t Enemy) bool, maxCount int) []Enemy {
	enemies := h.EnemiesWithinArea(a, filter)
	if enemies == nil {
		return nil
	}
	enemyCount := len(enemies)

	// 重複なしのランダムインデックスを生成
	indexes := h.Rand.Perm(enemyCount)

	// 返却するスライスの長さを決定
	count := maxCount
	if enemyCount < maxCount {
		count = enemyCount
	}

	// インデックスに従って敵を結果に追加
	result := make([]Enemy, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, enemies[indexes[i]])
	}
	return result
}

// 指定範囲内のランダムなガジェットのリストを返す。フィルター不要の場合はnilを渡す
func (h *Handler) RandomGadgetsWithinArea(a AttackPattern, filter func(t Gadget) bool, maxCount int) []Gadget {
	gadgets := h.GadgetsWithinArea(a, filter)
	if gadgets == nil {
		return nil
	}
	gadgetCount := len(gadgets)

	// 重複なしのランダムインデックスを生成
	indexes := h.Rand.Perm(gadgetCount)

	// 返却するスライスの長さを決定
	count := maxCount
	if gadgetCount < maxCount {
		count = gadgetCount
	}

	// インデックスに従ってガジェットを結果に追加
	result := make([]Gadget, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, gadgets[indexes[i]])
	}
	return result
}

// 最近接ターゲット

type enemyTuple struct {
	enemy Enemy
	dist  float64
}

func enemiesWithinAreaSorted(a AttackPattern, filter func(t Enemy) bool, skipAttackPattern bool, originalEnemies []Target) []enemyTuple {
	var enemies []enemyTuple

	hasFilter := filter != nil
	for _, v := range originalEnemies {
		e, ok := v.(Enemy)
		if !ok {
			panic("c.enemies should contain targets that implement the Enemy interface")
		}
		if hasFilter && !filter(e) {
			continue
		}
		if !e.IsAlive() {
			continue
		}
		if !skipAttackPattern && !e.IsWithinArea(a) {
			continue
		}
		enemies = append(enemies, enemyTuple{enemy: e, dist: a.Shape.Pos().Sub(e.Pos()).MagnitudeSquared()})
	}

	if len(enemies) == 0 {
		return nil
	}

	sort.Slice(enemies, func(i, j int) bool {
		return enemies[i].dist < enemies[j].dist
	})

	return enemies
}

type gadgetTuple struct {
	gadget Gadget
	dist   float64
}

func gadgetsWithinAreaSorted(a AttackPattern, filter func(t Gadget) bool, skipAttackPattern bool, originalGadgets []Gadget) []gadgetTuple {
	var gadgets []gadgetTuple

	hasFilter := filter != nil
	for _, v := range originalGadgets {
		if v == nil {
			continue
		}
		// ガジェットが敵陣営かチェック、味方ガジェットはアビリティの対象にならない
		if !(v.GadgetTyp() > StartGadgetTypEnemy && v.GadgetTyp() < EndGadgetTypEnemy) {
			continue
		}
		if hasFilter && !filter(v) {
			continue
		}
		if !v.IsAlive() {
			continue
		}
		if !skipAttackPattern && !v.IsWithinArea(a) {
			continue
		}
		gadgets = append(gadgets, gadgetTuple{gadget: v, dist: a.Shape.Pos().Sub(v.Pos()).MagnitudeSquared()})
	}

	if len(gadgets) == 0 {
		return nil
	}

	sort.Slice(gadgets, func(i, j int) bool {
		return gadgets[i].dist < gadgets[j].dist
	})

	return gadgets
}

// 指定位置に最も近い敵を返す。距離制限なし。pkg外では使用しないこと
func (h *Handler) ClosestEnemy(pos geometry.Point) Enemy {
	enemies := enemiesWithinAreaSorted(NewCircleHitOnTarget(pos, nil, 1), nil, true, h.enemies)
	if enemies == nil {
		return nil
	}
	return enemies[0].enemy
}

// 指定位置に最も近いガジェットを返す。距離制限なし。pkg外では使用しないこと
func (h *Handler) ClosestGadget(pos geometry.Point) Gadget {
	gadgets := gadgetsWithinAreaSorted(NewCircleHitOnTarget(pos, nil, 1), nil, true, h.gadgets)
	if gadgets == nil {
		return nil
	}
	return gadgets[0].gadget
}

// 指定範囲内で最も近い敵を返す。フィルター不要の場合はnilを渡す
func (h *Handler) ClosestEnemyWithinArea(a AttackPattern, filter func(t Enemy) bool) Enemy {
	enemies := enemiesWithinAreaSorted(a, filter, false, h.enemies)
	if enemies == nil {
		return nil
	}
	return enemies[0].enemy
}

// 指定範囲内で最も近いガジェットを返す。フィルター不要の場合はnilを渡す
func (h *Handler) ClosestGadgetWithinArea(a AttackPattern, filter func(t Gadget) bool) Gadget {
	gadgets := gadgetsWithinAreaSorted(a, filter, false, h.gadgets)
	if gadgets == nil {
		return nil
	}
	return gadgets[0].gadget
}

// 指定範囲内の敵を近い順にソートして返す。フィルター不要の場合はnilを渡す
func (h *Handler) ClosestEnemiesWithinArea(a AttackPattern, filter func(t Enemy) bool) []Enemy {
	enemies := enemiesWithinAreaSorted(a, filter, false, h.enemies)
	if enemies == nil {
		return nil
	}

	result := make([]Enemy, 0, len(enemies))
	for _, v := range enemies {
		result = append(result, v.enemy)
	}
	return result
}

// 指定範囲内のガジェットを近い順にソートして返す。フィルター不要の場合はnilを渡す
func (h *Handler) ClosestGadgetsWithinArea(a AttackPattern, filter func(t Gadget) bool) []Gadget {
	gadgets := gadgetsWithinAreaSorted(a, filter, false, h.gadgets)
	if gadgets == nil {
		return nil
	}

	result := make([]Gadget, 0, len(gadgets))
	for _, v := range gadgets {
		result = append(result, v.gadget)
	}
	return result
}
