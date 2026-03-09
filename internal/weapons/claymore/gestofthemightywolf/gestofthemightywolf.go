package gestofthemightywolf

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.GestOfTheMightyWolf, NewWeapon)
}

type Weapon struct {
	Index int

	stacks   int
	stackSrc int // スタック失効用のソースフレーム
	core     *core.Core
	char     *character.CharWrapper

	// パッシブ値
	dmgPerStack  float64
	cdmgPerStack float64
	hasHexBonus  bool
}

const (
	buffKey  = "gest-wolf-hymn"
	buffDur  = 4 * 60 // 4秒
	stackICD = 1      // 0.01秒 = 約1フレームICD
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	// 初期化時にHexereiボーナスをチェック（パーティ編成後）
	w.hasHexBonus = false
	hexereiCount := 0
	for _, char := range w.core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	w.hasHexBonus = hexereiCount >= 2
	return nil
}

// 攻撃速度が10%増加。
// 装備キャラクターの通常攻撃が敵に命中した時、元素スキルを発動した時、
// または重撃を開始した時、それぞれFour Winds' Hymnを1/2/2スタック獲得：
// 与えるダメージが4秒間7.5%/9.5%/11.5%/13.5%/15.5%増加。最大4スタック。
// この効果は0.01秒に1回発動可能。
// さらに、パーティに「Hexerei: Secret Rite」がある場合、
// Four Winds' Hymnの各スタックは装備キャラクターの会心ダメージも
// 7.5%/9.5%/11.5%/13.5%/15.5%増加させる。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	w.dmgPerStack = 0.055 + 0.02*float64(r) // R1=7.5%, R2=9.5%, R3=11.5%, R4=13.5%, R5=15.5%
	w.cdmgPerStack = w.dmgPerStack          // 会心ダメージも同じ値

	// 永続攻撃速度+10%（常時有効、説明文に基づき精錬でスケールしない）
	// 注: 攻撃速度は通常AtkSpdのStatModで適用
	atkSpd := make([]float64, attributes.EndStatType)
	atkSpd[attributes.AtkSpd] = 0.10
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("gest-wolf-atkspd", -1),
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			return atkSpd, true
		},
	})

	// 通常攻撃命中で1スタック獲得をサブスクライブ
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		w.addStacks(1)
		return false
	}, fmt.Sprintf("gest-wolf-na-%v", char.Base.Key.String()))

	// スキル発動で2スタック獲得をサブスクライブ（キャスト時ではなくダメージ時）
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		w.addStacks(2)
		return false
	}, fmt.Sprintf("gest-wolf-skill-%v", char.Base.Key.String()))

	// 重撃開始で2スタック獲得をサブスクライブ（ダメージ時）
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		w.addStacks(2)
		return false
	}, fmt.Sprintf("gest-wolf-ca-%v", char.Base.Key.String()))

	// AttackMod経由でダメージ%バフを適用（現在のスタック数に基づく動的計算）
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("gest-wolf-dmg", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = w.dmgPerStack * float64(w.stacks)
			if w.hasHexBonus {
				m[attributes.CD] = w.cdmgPerStack * float64(w.stacks)
			}
			return m, true
		},
	})

	return w, nil
}

// addStacks はFour Winds' Hymnバフにスタックを追加する
func (w *Weapon) addStacks(count int) {
	// ICDチェック：0.01秒
	if w.char.StatusIsActive("gest-wolf-stack-icd") {
		return
	}
	w.char.AddStatus("gest-wolf-stack-icd", stackICD, true)

	// バフが期限切れの場合、スタックをリセット
	if !w.char.StatusIsActive(buffKey) {
		w.stacks = 0
	}

	w.stacks += count
	if w.stacks > 4 {
		w.stacks = 4
	}

	// 持続時間を更新
	w.char.AddStatus(buffKey, buffDur, true)

	w.core.Log.NewEvent("gest-wolf stacks updated", glog.LogWeaponEvent, w.char.Index).
		Write("stacks", w.stacks).
		Write("count_added", count).
		Write("hex_bonus", w.hasHexBonus)
}
