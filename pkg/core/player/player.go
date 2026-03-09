// Package player はプレイヤーに関連するトラッキングと機能を含む:
// - チーム内キャラクターのトラッキング
// - アニメーション状態の管理
// - 通常攻撃状態の管理
// - キャラクターステータスと属性の管理
// - シールドの管理
package player

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/animation"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/infusion"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/verdant"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	MaxStam      = 240
	StamCDFrames = 90
	SwapCDFrames = 60
)

type Handler struct {
	Opt
	// ハンドラー
	*animation.AnimationHandler
	Shields *shield.Handler
	infusion.Handler
	Verdant *verdant.Handler

	// トラッキング
	chars   []*character.CharWrapper
	active  int
	charPos map[keys.Char]int

	// スタミナ
	Stam            float64
	LastStamUse     int
	stamPercentMods []stamPercentMod

	// 空中状態のソース
	airborne AirborneSource

	// 交代
	SwapCD  int
	SwapICD int

	// ダッシュ: ロックアウト中かつCD中の場合のみダッシュが失敗する
	DashCDExpirationFrame int
	DashLockout           bool

	// 最後のアクション
	LastAction struct {
		UsedAt int
		Type   action.Action
		Param  map[string]int
		Char   int
	}
}

type Opt struct {
	F            *int
	Log          glog.Logger
	Events       event.Eventter
	Tasks        task.Tasker
	Delays       info.Delays
	Debug        bool
	EnableHitlag bool
}

func New(opt Opt) *Handler {
	h := &Handler{
		chars:           make([]*character.CharWrapper, 0, 4),
		charPos:         make(map[keys.Char]int),
		stamPercentMods: make([]stamPercentMod, 0, 5),
		Opt:             opt,
		Stam:            MaxStam,
		SwapICD:         SwapCDFrames,
	}
	h.Shields = shield.New(opt.F, opt.Log, opt.Events)
	h.Handler = infusion.New(opt.F, opt.Log, opt.Debug)
	h.AnimationHandler = animation.New(opt.F, opt.Debug, opt.Log, opt.Events, opt.Tasks)
	h.Verdant = verdant.New(opt.F, opt.Events, opt.Tasks, opt.Log, &moonridgeImpl{h: h})
	return h
}

type moonridgeImpl struct {
	h *Handler
}

func (m *moonridgeImpl) IsUnlocked() bool {
	for _, c := range m.h.chars {
		if c.Base.Key == keys.Columbina && c.Base.Ascension >= 4 {
			return true
		}
	}
	return false
}

func (m *moonridgeImpl) HasLawOfNewMoon() bool {
	for _, c := range m.h.chars {
		if c.StatusIsActive("law-of-new-moon") {
			return true
		}
	}
	return false
}

func (m *moonridgeImpl) HasICD() bool {
	for _, c := range m.h.chars {
		if c.StatusIsActive("law-of-new-moon-icd") {
			return true
		}
	}
	return false
}

func (m *moonridgeImpl) AddICD(dur int) {
	// アクティブキャラクターに追加する
	if m.h.active >= 0 && m.h.active < len(m.h.chars) {
		m.h.chars[m.h.active].AddStatus("law-of-new-moon-icd", dur, false)
	}
}

func (h *Handler) SetSwapICD(delay int) {
	h.SwapICD = delay
}

func (h *Handler) swap(to keys.Char) func() {
	return func() {
		prev := h.active
		h.active = h.charPos[to]

		// ダッシュCDの残りフレームがある場合、再出場時のためキャラに保存する
		if h.DashCDExpirationFrame > *h.F {
			h.chars[prev].RemainingDashCD = h.DashCDExpirationFrame - *h.F
			h.chars[prev].DashLockout = h.DashLockout
		}

		// 新しい DashCDExpirationFrame を設定し、キャラの残りを0にリセットする
		h.DashCDExpirationFrame = *h.F + h.chars[h.active].RemainingDashCD
		h.DashLockout = h.chars[h.active].DashLockout
		h.chars[h.active].RemainingDashCD = 0

		h.SwapCD = h.SwapICD
		h.ResetAllNormalCounter()

		evt := h.Log.NewEvent("executed swap", glog.LogActionEvent, h.active).
			Write("action", "swap").
			Write("target", to.String())

		if h.chars[prev].RemainingDashCD > 0 {
			evt.Write("prev_dash_cd", h.chars[prev].RemainingDashCD).
				Write("prev_dash_lockout", h.chars[prev].DashLockout)
		}

		if h.DashCDExpirationFrame > *h.F {
			evt.Write("target_dash_cd", h.DashCDExpirationFrame-*h.F).
				Write("target_dash_expiry_frame", h.DashCDExpirationFrame).
				Write("target_dash_lockout", h.DashLockout)
		}

		h.Events.Emit(event.OnCharacterSwap, prev, h.active)
	}
}

func (h *Handler) AddChar(char *character.CharWrapper) int {
	h.chars = append(h.chars, char)
	index := len(h.chars) - 1
	char.SetIndex(index)
	h.charPos[char.Base.Key] = index

	return index
}

func (h *Handler) ByIndex(i int) *character.CharWrapper {
	return h.chars[i]
}

func (h *Handler) CombatByIndex(i int) combat.Character {
	return h.chars[i]
}

func (h *Handler) ByKey(k keys.Char) (*character.CharWrapper, bool) {
	i, ok := h.charPos[k]
	if !ok {
		return nil, false
	}
	return h.chars[i], true
}

func (h *Handler) Chars() []*character.CharWrapper {
	return h.chars
}

func (h *Handler) Active() int {
	return h.active
}

func (h *Handler) ActiveChar() *character.CharWrapper {
	return h.chars[h.active]
}

func (h *Handler) CharIsActive(k keys.Char) bool {
	return h.charPos[k] == h.active
}

func (h *Handler) SetActive(i int) {
	h.active = i
}

func (h *Handler) Adjust(src string, char int, amt float64) {
	h.chars[char].AddEnergy(src, amt)
}

func (h *Handler) ResetAllNormalCounter() {
	for _, char := range h.chars {
		char.ResetNormalCounter()
	}
}

func (h *Handler) DistributeParticle(p character.Particle) {
	for i, char := range h.chars {
		char.ReceiveParticle(p, h.active == i, len(h.chars))
	}
	h.Events.Emit(event.OnParticleReceived, p)
}

func (h *Handler) AbilStamCost(i int, a action.Action, p map[string]int) float64 {
	// スタミナ軽減modは負の値
	// スタミナ軽減は最大100%に制限する
	r := 1 + h.StamPercentMod(a)
	if r < 0 {
		r = 0
	}
	return r * h.chars[i].ActionStam(a, p)
}

func (h *Handler) UseStam(amount float64, a action.Action) {
	h.Stam -= amount
	// これは本来起きないはず？
	if h.Stam < 0 {
		h.Stam = 0
	}
	h.LastStamUse = *h.F
	h.Events.Emit(event.OnStamUse, a)
}

func (h *Handler) RestoreStam(v float64) {
	h.Stam += v
	if h.Stam > MaxStam {
		h.Stam = MaxStam
	}
}

func (h *Handler) ApplyHitlag(char int, factor, dur float64) {
	// このキャラがフィールド上にいる場合のみヒットラグを適用する
	if char != h.active {
		return
	}

	h.chars[char].ApplyHitlag(factor, dur)

	// 元素付与も延長する
	//TODO: ここで適用するのは非常に不自然
	h.ExtendInfusion(char, factor, dur)

	// ヒットラグ延長分だけダッシュCDを延長する
	if h.DashCDExpirationFrame > *h.F {
		ext := int(math.Ceil(dur * (1 - factor)))
		h.DashCDExpirationFrame += ext

		var evt glog.Event
		if h.DashLockout {
			evt = h.Log.NewEvent("dash cd hitlag extended", glog.LogHitlagEvent, char)
		} else {
			evt = h.Log.NewEvent("dash lockout evaluation hitlag extended", glog.LogHitlagEvent, char)
		}
		evt.Write("extension", ext).
			Write("expiry", h.DashCDExpirationFrame-*h.F).
			Write("expiry_frame", h.DashCDExpirationFrame).
			Write("lockout", h.DashLockout)
	}
}

// InitializeTeam は元素共鳴のイベントフックをセットアップし、
// 全キャラクターの基礎ステータスを計算する
func (h *Handler) InitializeTeam() error {
	var err error
	for _, c := range h.chars {
		err = c.UpdateBaseStats()
		if err != nil {
			return err
		}
	}
	// 再度ループして初期化する
	for i := range h.chars {
		err = h.chars[i].Init()
		if err != nil {
			return err
		}
		h.chars[i].Equip.Weapon.Init()
		for k := range h.chars[i].Equip.Sets {
			h.chars[i].Equip.Sets[k].Init()
		}
		// 各キャラの初期HP割合を設定する
		switch {
		case h.chars[i].StartHP > 0 && h.chars[i].StartHPRatio > 0:
			h.chars[i].SetHPByRatio(float64(h.chars[i].StartHPRatio) / 100.0)
			h.chars[i].ModifyHPByAmount(float64(h.chars[i].StartHP))
		case h.chars[i].StartHP > 0:
			h.chars[i].SetHPByAmount(float64(h.chars[i].StartHP))
		case h.chars[i].StartHPRatio > 0:
			h.chars[i].SetHPByRatio(float64(h.chars[i].StartHPRatio) / 100.0)
		default:
			h.chars[i].SetHPByRatio(1)
		}
		h.Log.NewEvent("starting hp set", glog.LogCharacterEvent, i).
			Write("starting_hp_ratio", h.chars[i].CurrentHPRatio()).
			Write("starting_hp", h.chars[i].CurrentHP())
	}

	// パーティ全体のMoonsign状態を一度だけ決定する（nascent/ascendant）
	count := 0
	for _, ch := range h.chars {
		if ch.StatusIsActive("moonsignKey") {
			count++
		}
	}
	for _, ch := range h.chars {
		ch.MoonsignNascent = false
		ch.MoonsignAscendant = false
		switch count {
		case 1:
			ch.MoonsignNascent = true
		case 2, 3, 4:
			ch.MoonsignAscendant = true
		}
	}
	h.Log.NewEvent("moonsign party init", glog.LogDebugEvent, -1).
		Write("count", count)

	// 非Moonsignキャラクターのバフを初期化する（Ascendant Gleamのみ）
	h.initNonMoonsignBuffs()

	return nil
}

func (h *Handler) Tick() {
	//	- プレイヤー（スタミナ、交代、アニメーション等）
	//		- キャラクター
	//		- シールド
	//		- アニメーション
	//		- スタミナ
	//		- 交代
	// スタミナを回復
	if h.Stam < MaxStam && *h.F-h.LastStamUse > StamCDFrames {
		h.Stam += 25.0 / 60
		if h.Stam > MaxStam {
			h.Stam = MaxStam
		}
	}
	if h.SwapCD > 0 {
		h.SwapCD--
	}
	h.Shields.Tick()
	h.AnimationHandler.Tick()
	for _, c := range h.chars {
		c.Tick()
	}
}

type AirborneSource int

const (
	Grounded AirborneSource = iota
	AirborneXiao
	AirborneVenti
	AirborneKazuha
	AirborneXianyun
	AirborneVaresa
	TerminateAirborne
)

func (h *Handler) SetAirborne(src AirborneSource) error {
	if src < Grounded || src >= TerminateAirborne {
		// 何もしない
		return fmt.Errorf("invalid airborne source: %v", src)
	}
	h.airborne = src
	return nil
}

func (h *Handler) Airborne() AirborneSource {
	return h.airborne
}

// initNonMoonsignBuffs は、Ascendant Gleam が有効な場合（満照状態）に、
// 非Moonsignキャラクター向けの月相反応ダメージバフを設定する。
func (h *Handler) initNonMoonsignBuffs() {
	const (
		buffDuration = 20 * 60 // 20秒
		buffKey      = "non-moonsign-lunar-buff"
	)

	for _, char := range h.chars {
		// Moonsignキャラクター（moonsignKeyを持つ）をスキップする
		if char.StatusIsActive("moonsignKey") {
			continue
		}

		// クロージャ用にループ変数をキャプチャする
		currentChar := char
		currentCharIndex := char.Index
		currentCharEle := char.Base.Element

		// OnSkill と OnBurst イベントを購読する
		subscribeKey := fmt.Sprintf("non-moonsign-buff-%d", currentCharIndex)

		handler := func(args ...interface{}) bool {
			// トリガーしたキャラクターがこのキャラクターの場合のみ有効化する
			if h.active != currentCharIndex {
				return false
			}

			// Ascendant Gleam 状態（満照）の場合のみ有効化する
			if !currentChar.MoonsignAscendant {
				return false
			}

			// トリガーしたキャラクターの元素タイプとステータスに基づいてボーナスを計算する
			bonus := h.calculateNonMoonsignBonus(currentCharEle, currentChar)

			h.Log.NewEvent("non-moonsign lunar buff activated", glog.LogCharacterEvent, currentCharIndex).
				Write("element", currentCharEle.String()).
				Write("bonus", bonus).
				Write("trigger_char", currentChar.Base.Key)

			// バフを全パーティメンバーに適用する
			h.applyNonMoonsignBuff(buffKey, buffDuration, bonus)

			return false
		}

		h.Events.Subscribe(event.OnSkill, handler, subscribeKey+"-skill")
		h.Events.Subscribe(event.OnBurst, handler, subscribeKey+"-burst")
	}
}

// calculateNonMoonsignBonus はキャラクターの元素タイプと現在のステータスに基づいて、
// 月相反応ダメージボーナスを計算する。
func (h *Handler) calculateNonMoonsignBonus(ele attributes.Element, char *character.CharWrapper) float64 {
	maxBonus := 0.36 // 最大ボーナス 36%

	switch ele {
	case attributes.Pyro, attributes.Electro, attributes.Cryo:
		// ATK * 0.009、最大 4000 ATK
		atk := char.Stat(attributes.ATK)
		bonus := (atk / 100) * 0.009
		if bonus > maxBonus {
			bonus = maxBonus
		}
		return bonus

	case attributes.Hydro:
		// MaxHP * 0.0006、最大 60000 HP
		maxHP := char.MaxHP()
		bonus := (maxHP / 1000) * 0.006
		if bonus > maxBonus {
			bonus = maxBonus
		}
		return bonus

	case attributes.Geo:
		// DEF * 0.01、最大 3600 DEF
		def := char.Stat(attributes.DEF)
		bonus := (def / 100) * 0.01
		if bonus > maxBonus {
			bonus = maxBonus
		}
		return bonus

	case attributes.Anemo, attributes.Dendro:
		// EM * 0.0225、最大 1600 EM
		em := char.Stat(attributes.EM)
		bonus := (em / 100) * 0.0225
		if bonus > maxBonus {
			bonus = maxBonus
		}
		return bonus

	default:
		return 0
	}
}

// applyNonMoonsignBuff は月相反応ダメージバフを全パーティメンバーに適用する。
func (h *Handler) applyNonMoonsignBuff(buffKey string, duration int, bonus float64) {
	for _, char := range h.chars {
		// 既存のバフがあれば削除する
		char.DeleteStatus(buffKey)

		// 新しいバフステータスを追加する
		char.AddStatus(buffKey, duration, true)

		// 月相反応ダメージボーナス用の ElevationMod を追加する
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBaseWithHitlag(buffKey, duration),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				// 全ての月相反応タイプに適用する
				if ai.AttackTag == attacks.AttackTagLCDamage ||
					ai.AttackTag == attacks.AttackTagLBDamage ||
					ai.AttackTag == attacks.AttackTagLCrsDamage {
					return bonus, false
				}
				return 0, false
			},
		})
	}
}

const (
	XianyunAirborneBuff = "xianyun-airborne-buff"
)
