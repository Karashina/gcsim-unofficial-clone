package combat

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

type Target interface {
	Key() targets.TargetKey        // ターゲットのユニークキー
	SetKey(k targets.TargetKey)    // キーを更新
	Type() targets.TargettableType // ターゲットの種類
	Shape() geometry.Shape         // ターゲットの形状
	Pos() geometry.Point           // ターゲットの中心
	SetPos(p geometry.Point)       // ターゲットを移動
	IsAlive() bool
	SetTag(key string, val int)
	GetTag(key string) int
	RemoveTag(key string)
	HandleAttack(*AttackEvent) float64
	AttackWillLand(a AttackPattern) (bool, string) // 被ダメボックスがAttackPatternと衝突するか
	IsWithinArea(a AttackPattern) bool             // 中心がAttackPattern内にあるか
	Tick()                                         // 毎ティック呼び出し
	Kill()
	// 衝突判定用
	CollidableWith(targets.TargettableType) bool
	CollidedWith(t Target)
	WillCollide(geometry.Shape) bool
	// 方向関連
	Direction() geometry.Point                           // 視線方向をgeometry.Pointとして返す
	SetDirection(trg geometry.Point)                     // デフォルト方向(0, 1)からの相対的な視線方向を計算
	SetDirectionToClosestEnemy()                         // 最も近い敵を向く
	CalcTempDirection(trg geometry.Point) geometry.Point // 弓重撃などに使用
}

type TargetWithAura interface {
	Target
	AuraContains(e ...attributes.Element) bool
}
