package combat

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

type AttackEvent struct {
	Info    AttackInfo
	Pattern AttackPattern
	// Timing        AttackTiming
	Snapshot    Snapshot
	SourceFrame int            // ソースフレーム
	Callbacks   []AttackCBFunc `json:"-"`
	Reacted     bool           // trueの場合、元素反応が既に発生済み — 付着/補充用
}

type AttackCB struct {
	Target      Target
	AttackEvent *AttackEvent
	Damage      float64
	IsCrit      bool
}

type AttackCBFunc func(AttackCB)

type AttackInfo struct {
	ActorIndex       int               // この攻撃が属するキャラクター
	DamageSrc        targets.TargetKey // この攻撃のソース。ターゲットを識別するユニークキー
	Abil             string            // ダメージをトリガーするアビリティ名
	AttackTag        attacks.AttackTag
	AdditionalTags   []attacks.AdditionalTag
	PoiseDMG         float64 // 現時点では氷結消費前の粉砕用、銀撃攻撃のみ必要
	ICDTag           attacks.ICDTag
	ICDGroup         attacks.ICDGroup
	Element          attributes.Element   // アビリティの元素
	Durability       reactions.Durability // 元素付着量。付着無しの場合は0
	NoImpulse        bool
	HitWeakPoint     bool
	Mult             float64 // アビリティ倍率。初期モナダメージでは0に設定可能
	StrikeType       attacks.StrikeType
	UseDef           bool    // ステータススナップショットが正しく動作するようflatdmgの代わりに使用
	UseHP            bool    // ステータススナップショットが正しく動作するようflatdmgの代わりに使用
	FlatDmg          float64 // 固定ダメージ
	IgnoreDefPercent float64 // デフォルトは0。 1の場合は防御力無視。雷電将軍2凸は0.6（60%無視）に設定
	IgnoreInfusion   bool
	// 増幅情報
	Amped   bool                   // 新反応システム用フラグ
	AmpMult float64                // 増幅倍率
	AmpType reactions.ReactionType // 融解または蒸発
	// 激化情報
	Catalyzed     bool
	CatalyzedType reactions.ReactionType
	// シミュレーションが生成した攻撃用の特殊フラグ
	SourceIsSim bool
	DoNotLog    bool
	// ヒットラグ関連
	HitlagHaltFrames     float64 // 一時停止するフレーム数
	HitlagFactor         float64 // クロックを遅くする係数
	CanBeDefenseHalted   bool    // 遍機守衛などに対する攻撃用
	IsDeployable         bool    // trueの場合、ヒットラグは所有者に影響しない
	HitlagOnHeadshotOnly bool    // trueの場合、HitWeakpointがtrueの時のみ適用

	AuraExpiry int // LC貢献者追跡用
}

type Snapshot struct {
	Stats   attributes.Stats // 聖遺物・ボーナス等を含むキャラクターの合計ステータス
	CharLvl int

	SourceFrame int           // スナップショットが生成されたフレーム
	Logs        []interface{} // スナップショットのログ
}
