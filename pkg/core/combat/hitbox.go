package combat

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

type AttackPattern struct {
	Shape       geometry.Shape
	SkipTargets [targets.TargettableTypeCount]bool
	IgnoredKeys []targets.TargetKey
}

type positional interface {
	Pos() geometry.Point
}

func NewSingleTargetHit(ind targets.TargetKey) AttackPattern {
	a := AttackPattern{
		Shape: &geometry.SingleTarget{Target: ind},
	}
	a.SkipTargets[targets.TargettablePlayer] = true
	return a
}

func getCenterAndDirection(src, center, offset positional) (geometry.Point, geometry.Point) {
	c := center.Pos()
	dir := geometry.DefaultDirection()
	srcTrg, srcIsATarget := src.(Target)
	centerTrg, centerIsATarget := center.(Target)

	// オフセット追加に使用する方向を決定
	if srcIsATarget {
		dir = srcTrg.Direction()
		// 方向を再計算
		// - 提供された中心が単なる位置の場合（弓の重撃ターゲットに有用）
		// - 提供されたターゲットのキーが異なる場合
		if !centerIsATarget || srcTrg.Key() != centerTrg.Key() {
			dir = srcTrg.CalcTempDirection(c)
		}
	}

	// オフセットなしのショートカットとしてnilを許可
	if offset == nil {
		return c, dir
	}

	off := offset.Pos()
	// オフセットを追加
	if off.X == 0 && off.Y == 0 {
		return c, dir
	}
	newCenter := geometry.CalcOffsetPoint(c, off, dir)
	return newCenter, dir
}

func NewCircleHit(src, center, offset positional, r float64) AttackPattern {
	c, dir := getCenterAndDirection(src, center, offset)
	a := AttackPattern{
		Shape: geometry.NewCircle(c, r, dir, 360),
	}
	a.SkipTargets[targets.TargettablePlayer] = true
	return a
}

func NewCircleHitFanAngle(src, center, offset positional, r, fanAngle float64) AttackPattern {
	c, dir := getCenterAndDirection(src, center, offset)
	a := AttackPattern{
		Shape: geometry.NewCircle(c, r, dir, fanAngle),
	}
	a.SkipTargets[targets.TargettablePlayer] = true
	return a
}

func NewCircleHitOnTarget(trg, offset positional, r float64) AttackPattern {
	return NewCircleHit(trg, trg, offset, r)
}

func NewCircleHitOnTargetFanAngle(trg, offset positional, r, fanAngle float64) AttackPattern {
	return NewCircleHitFanAngle(trg, trg, offset, r, fanAngle)
}

func NewBoxHit(src, center, offset positional, w, h float64) AttackPattern {
	c, dir := getCenterAndDirection(src, center, offset)
	a := AttackPattern{
		Shape: geometry.NewRectangle(c, w, h, dir),
	}
	a.SkipTargets[targets.TargettablePlayer] = true
	return a
}

func NewBoxHitOnTarget(trg, offset positional, w, h float64) AttackPattern {
	return NewBoxHit(trg, trg, offset, w, h)
}
