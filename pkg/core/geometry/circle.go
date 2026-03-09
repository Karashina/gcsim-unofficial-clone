package geometry

import (
	"fmt"
	"math"
)

type Circle struct {
	center   Point
	r        float64
	dir      Point
	fanAngle float64
	segments []Point
	aabb     []Point
}

func NewCircle(center Point, r float64, dir Point, fanAngle float64) *Circle {
	var segments []Point
	if fanAngle > 0 && fanAngle < 360 {
		segments = calcSegments(center, r, dir, fanAngle)
	}
	return &Circle{
		center:   center,
		r:        r,
		dir:      dir,
		fanAngle: fanAngle,
		segments: segments,
		aabb:     calcCircleAABB(center, r),
	}
}

func (c *Circle) Pos() Point {
	return c.center
}

func (c *Circle) Radius() float64 {
	return c.r
}

func (c *Circle) SetPos(p Point) {
	if c.center == p {
		return
	}
	for i := 0; i < len(c.segments); i++ {
		c.segments[i] = c.segments[i].Add(p.Sub(c.center))
	}
	for i := 0; i < len(c.aabb); i++ {
		c.aabb[i] = c.aabb[i].Add(p.Sub(c.center))
	}
	c.center = p
}

func (c *Circle) String() string {
	return fmt.Sprintf(
		"r: %v x: %v y: %v dir: %v fanAngle: %v segments: %v",
		c.r, c.center.X, c.center.Y, c.dir, c.fanAngle, c.segments,
	)
}

func calcSegments(center Point, r float64, dir Point, fanAngle float64) []Point {
	// 最初は回転処理のために円の中心を原点と仮定する
	segmentStart := Point{X: 0, Y: r}.Rotate(dir)
	segmentLeft := segmentStart.Rotate(DegreesToDirection(-fanAngle / 2))
	segmentRight := segmentStart.Rotate(DegreesToDirection(fanAngle / 2))
	// セグメントの点を保存（円の中心とセグメント点で線分を構成）
	// セグメントを実際の円の中心位置に移動する必要がある
	return []Point{segmentLeft.Add(center), segmentRight.Add(center)}
}

// AABB は常に完全な円に対して計算する
func calcCircleAABB(center Point, r float64) []Point {
	return []Point{{X: center.X - r, Y: center.Y - r}, {X: center.X + r, Y: center.Y + r}}
}

// 衝突関連

func (c *Circle) PointInShape(p Point) bool {
	rangeCheck := c.center.Sub(p).MagnitudeSquared() <= c.r*c.r
	if c.segments == nil {
		return rangeCheck
	}
	return rangeCheck && fanAngleAreaCheck(c.center, p, c.dir, c.fanAngle)
}

func (c *Circle) IntersectCircle(c2 Circle) bool {
	// TODO: fanAngle 付き円形ハートボックスの円-円衝突は未実装
	if c.segments != nil {
		panic("target with fanAngle hurtbox isn't supported in circle-circle collision")
	}
	// https://stackoverflow.com/a/4226473
	// A: 完全な円が交差している必要がある
	// (R0 - R1)^2 <= (x0 - x1)^2 + (y0 - y1)^2 <= (R0 + R1)^2
	radiusSum := c.r + c2.r
	if c.center.Sub(c2.center).MagnitudeSquared() > radiusSum*radiusSum {
		return false
	}

	// c2 に fanAngle がない -> A を満たせば交差
	if c2.segments == nil {
		return true
	}

	// c2 に fanAngle がある -> A && (B || C) を満たせば交差
	// https://www.baeldung.com/cs/circle-line-segment-collision-detection
	// (注: maxDist チェックは不要。セグメント全体が円内に含まれる場合も交差として扱うため)
	// B: c1 が c2 のセグメントのいずれかと交差するか確認。交差すれば早期リターン。
	// (c1 の円中心が c2 の fanAngle 範囲外でも、c2 の fanAngle 領域と
	// 衝突する可能性があるため、このチェックが必要)
	o := c.center
	p := c2.center

	op := p.Sub(o)
	opDist := o.Distance(p)
	for _, segment := range c2.segments {
		q := segment

		qp := p.Sub(q)
		pq := q.Sub(p)

		oq := q.Sub(o)
		oqDist := o.Distance(q)

		minDist := min(opDist, oqDist)
		if op.Dot(qp) > 0 && oq.Dot(pq) > 0 {
			minDist = math.Abs(op.Cross(oq)) / c2.r
		}
		if minDist <= c.r {
			return true
		}
	}

	// C: c2 から c1 への方向ベクトルと y 軸のなす角が c2 の fanAngle 内にあるか確認
	return fanAngleAreaCheck(c2.center, c.center, c2.dir, c2.fanAngle)
}

func (c *Circle) IntersectRectangle(r Rectangle) bool {
	return IntersectRectangle(r, *c)
}
