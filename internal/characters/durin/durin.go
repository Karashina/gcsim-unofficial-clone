package durin

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Durin, NewChar)
}

// char はデュリンキャラクターの実装
// 変容状態と龍召喚能力を持つ炎元素キャラクター
type char struct {
	*tmpl.Character

	// 状態管理
	stateDenial bool // true = 暗黒の否定、false = 純化の肯定

	// 龍の追跡
	dragonWhiteFlame bool // 白焔の龍が召喚されているか
	dragonDarkDecay  bool // 暗蝕の龍が召喚されているか
	dragonExpiry     int  // アクティブな龍の有効期限フレーム
	dragonSrc        int  // 龍インスタンスを追跡するソースID（古い龍の攻撃を防止）

	// A4天賦: 夜に似た混沌 (Primordial Fusion)
	primordialFusionStacks int // 原初融合スタック数（最大10）
	primordialFusionExpiry int // スタック有効期限フレーム

	// 1凸（命ノ星座1）: アダマの救済 (Cycle of Enlightenment)
	cycleStacks map[int]int // キャラクターインデックス → 啓示のサイクルスタック数
	cycleExpiry map[int]int // キャラクターインデックス → 有効期限フレーム

	// Hexereiシステム（特殊キャラクタークラス）
	isHexerei bool // このキャラクターがHexerei属性を持つか

	// 元素スキル内部クールダウン（ICD）
	lastEnergyRestoreFrame int  // エネルギー回復がトリガーされた最後のフレーム
	particleIcd            bool // 粒子生成のICD状態
}

// NewChar はデュリンキャラクターの新しいインスタンスを生成
func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{
		cycleStacks:            make(map[int]int),
		cycleExpiry:            make(map[int]int),
		lastEnergyRestoreFrame: -9999, // 遠い過去のフレームに初期化
		isHexerei:              true,  // デフォルトでHexerei属性を持つ
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	// キャラクター基本パラメータ
	c.EnergyMax = 70   // 元素爆発に必要なエネルギー
	c.NormalHitNum = 4 // 通常攻撃のヒット数
	c.SkillCon = 5     // 元素スキルの天賦レベルを上げる命ノ星座
	c.BurstCon = 3     // 元素爆発の天賦レベルを上げる命ノ星座

	// nohex=1パラメータが指定された場合Hexerei属性を無効化
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c
	return nil
}

// Init はキャラクターを初期化し天賦と命ノ星座の効果を設定
func (c *char) Init() error {
	// 天賦効果を初期化
	c.a1() // A1: 神算の光顕

	// 命ノ星座効果を初期化
	if c.Base.Cons >= 1 {
		c.c1() // 1凸: アダマの救済
	}
	if c.Base.Cons >= 2 {
		c.c2() // 2凸: 不穏な幻視
	}
	if c.Base.Cons >= 4 {
		c.c4() // 4凸: エマナレの源泉
	}

	return nil
}

// ActionReady はアクションが使用可能か確認
// 二重スキルメカニクスを処理: Essential Transmutationは即座に再発動可能
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// Essential Transmutation状態かつ再発動CD非アクティブ時、スキルの即座再発動を許可
	if a == action.ActionSkill && c.StatusIsActive(essentialTransmutationKey) {
		if c.StatusIsActive(skillRecastCDKey) {
			// 既に再発動済み、メインスキルCDを待つ必要あり
			return false, action.SkillCD
		}
		// 通常スキルCDがアクティブでも再発動を許可
		return true, action.NoFailure
	}

	return c.Character.ActionReady(a, p)
}

// Condition はキャラクターの状態クエリに応答
// 変容状態、龍の状態、スタック数などの確認が可能
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "state": // 現在の変容状態を文字列で返す
		if c.stateDenial {
			return "denial", nil // 暗黒の否定
		}
		return "confirmation", nil // 純化の肯定
	case "denial": // 暗黒の否定状態かどうか
		return c.stateDenial, nil
	case "confirmation": // 純化の肯定状態かどうか
		return !c.stateDenial, nil
	case "dragon-white-flame": // 白焔の龍がアクティブかどうか
		return c.dragonWhiteFlame && c.Core.F < c.dragonExpiry, nil
	case "dragon-dark-decay": // 暗蝕の龍がアクティブかどうか
		return c.dragonDarkDecay && c.Core.F < c.dragonExpiry, nil
	case "primordial-fusion-stacks": // A4原初融合のスタック数
		if c.Core.F >= c.primordialFusionExpiry {
			return 0, nil
		}
		return c.primordialFusionStacks, nil
	case "hexerei": // Hexerei属性を持つかどうか
		return c.isHexerei, nil
	case "cycle-stacks": // 1凸啓示のサイクルのスタック数（現在のアクティブキャラクター）
		charIndex := c.Core.Player.ActiveChar().Index
		if expiry, ok := c.cycleExpiry[charIndex]; ok && c.Core.F < expiry {
			return c.cycleStacks[charIndex], nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// hasHexereiBonus はパーティにHexereiキャラクターが2人以上いるか確認
// Hexereiボーナス: Hexereiキャラクター2人以上でパーティ内の各種効果が75%増加（1.75倍）
// 対象キャラクター: デュリン、アルベド、クレー、ウェンティ、モナ、フィッシュル、レザー、スクロース
func (c *char) hasHexereiBonus() bool {
	// このキャラクターがHexerei属性を持たない場合ボーナスなし
	if !c.isHexerei {
		return false
	}

	// Hexerei属性を持つパーティメンバーを数える
	hexereiCount := 0
	for _, char := range c.Core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHexerei, ok := result.(bool); ok && isHexerei {
				hexereiCount++
			}
		}
	}

	// Hexereiキャラクター2人以上でボーナスが発動
	return hexereiCount >= 2
}
