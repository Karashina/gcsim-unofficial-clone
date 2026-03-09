package geometry

import (
	"math"
	"math/rand"
)

type Shape interface {
	Pos() Point
	PointInShape(p Point) bool
	IntersectCircle(c Circle) bool
	IntersectRectangle(r Rectangle) bool
	String() string
}

func DefaultDirection() Point {
	return Point{X: 0, Y: 1}
}

// 指定された角度（度）を方向ベクトルに変換する。+は時計回り
func DegreesToDirection(angle float64) Point {
	radians := angle * math.Pi / 180
	return Point{
		X: math.Sin(radians),
		Y: math.Cos(radians),
	}
}

// dir はカスタム方向を渡す場合大きさが1である必要がある。
// pos + 回転済みオフセットの新しい Point を返す。
func CalcOffsetPoint(pos, offset, dir Point) Point {
	if dir == DefaultDirection() {
		return pos.Add(offset)
	}
	return pos.Add(offset.Rotate(dir))
}

// https://wumbo.net/formulas/angle-between-two-vectors-2d/
func CalcDirection(src, trg Point) Point {
	// Normalize()でのゼロ除算を回避
	if trg == src {
		return DefaultDirection()
	}
	return trg.Sub(src).Normalize()
}

// 中心から minRadius 以上 maxRadius 以下の距離にあるランダムな点を生成する
func CalcRandomPointFromCenter(center Point, minRadius, maxRadius float64, rand *rand.Rand) Point {
	// 棄却サンプリングで単位円内のランダム点を生成
	var result Point
	for {
		p := Point{
			X: -1 + rand.Float64()*2,
			Y: -1 + rand.Float64()*2,
		}
		if p.MagnitudeSquared() <= 1 {
			minRadiusSquared := minRadius * minRadius
			maxRadiusSquared := maxRadius * maxRadius
			// 指定範囲内のランダム半径を取得
			r := math.Sqrt(minRadiusSquared + rand.Float64()*(maxRadiusSquared-minRadiusSquared))
			// 生成した点をランダム半径上にスケーリングしてシフト
			if p.X == 0 && p.Y == 0 {
				p = Point{X: 0, Y: 1}
			}
			factor := r / p.Magnitude()
			result = p.Mul(Point{X: factor, Y: factor}).Add(center)
			break
		}
	}
	return result
}

func AABBTest(a, b []Point) bool {
	aMin := a[0]
	aMax := a[1]
	bMin := b[0]
	bMax := b[1]
	return aMin.X <= bMax.X && aMax.X >= bMin.X && aMin.Y <= bMax.Y && aMax.Y >= bMin.Y
}

// https://stackoverflow.com/questions/12234574/calculating-if-an-angle-is-between-two-angles
func fanAngleAreaCheck(attackCenter, trg, facingDirection Point, fanAngle float64) bool {
	// facingDirection と targetDirection は複数ターゲットの場合異なることがある
	targetDirection := CalcDirection(attackCenter, trg)
	dot := facingDirection.Dot(targetDirection)
	// 浮動小数点演算のため、ドット積を [-1, 1] にクランプする必要がある
	if dot > 1 {
		dot = 1
	}
	if dot < -1 {
		dot = -1
	}
	angleBetweenFacingAndTarget := math.Acos(dot) * 180 / math.Pi
	return angleBetweenFacingAndTarget >= -fanAngle/2 && angleBetweenFacingAndTarget <= fanAngle/2
}

// Circle と Rectangle で共有
// https://stackoverflow.com/questions/401847/circle-rectangle-collision-detection-intersection
// https://yal.cc/rot-rect-vs-circle-intersection/
func IntersectRectangle(r Rectangle, c Circle) bool {
	// TODO: fanAngleのハートボックス/ヒットボックスと矩形-円の衝突判定
	if c.segments != nil {
		panic("fanAngle hitbox and hurtbox aren't supported in rectangle-circle collision")
	}

	// AABB判定
	if !AABBTest(r.aabb, c.aabb) {
		return false
	}

	// 原点を矩形中心に設定（円の中心位置をシフト）
	relative := c.center.Sub(r.center)

	// 矩形の回転を除去するため、円の中心を逆方向に回転
	dir := r.dir.Mul(Point{X: -1, Y: 1})
	local := relative.Rotate(dir)

	// 円の中心を1象限に制約
	local.X = math.Abs(local.X)
	local.Y = math.Abs(local.Y)

	topRight := Point{
		X: r.w / 2,
		Y: r.h / 2,
	}

	// 円の中心が矩形の辺から遠すぎるケースを除外
	if local.X > c.r+topRight.X || local.Y > c.r+topRight.Y {
		return false
	}

	// この時点で円の中心は矩形の辺に十分近い
	// -> 円の中心が 0 <= x <= r.w/2 || 0 <= y <= r.h/2 の範囲内なら受理
	// -> その範囲内なら、確実に一辺と交差している
	if local.X <= topRight.X || local.Y <= topRight.Y {
		return true
	}

	// 円の中心が r.w/2 < x <= r.w/2+radius && r.h/2 < y <= r.h/2+radius の範囲にある
	// -> topRightの角に十分近い場合のみ交差
	return local.Sub(topRight).MagnitudeSquared() <= c.r*c.r
}
