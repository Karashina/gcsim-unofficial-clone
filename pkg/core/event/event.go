package event

type Event int

const (
	OnEnemyHit     Event = iota // ターゲット, AttackEvent
	OnPlayerHit                 // キャラクター, AttackEvent
	OnGadgetHit                 // ターゲット, AttackEvent
	OnEnemyDamage               // ターゲット, AttackEvent, ダメージ量, 会心
	OnGadgetDamage              // ターゲット, AttackEvent
	OnApplyAttack               // AttackEvent
	// 元素反応関連
	// OnReactionOccured // target, AttackEvent
	// OnTransReaction   // target, AttackEvent
	// OnAmpReaction     // target, AttackEvent

	OnAuraDurabilityAdded    // ターゲット, 元素, 元素量
	OnAuraDurabilityDepleted // ターゲット, 元素
	// OnReaction               // target, AttackEvent, ReactionType
	ReactionEventStartDelim
	OnOverload           // ターゲット, AttackEvent
	OnSuperconduct       // ターゲット, AttackEvent
	OnMelt               // ターゲット, AttackEvent
	OnVaporize           // ターゲット, AttackEvent
	OnFrozen             // ターゲット, AttackEvent
	OnElectroCharged     // ターゲット, AttackEvent
	OnSwirlHydro         // ターゲット, AttackEvent
	OnSwirlCryo          // ターゲット, AttackEvent
	OnSwirlElectro       // ターゲット, AttackEvent
	OnSwirlPyro          // ターゲット, AttackEvent
	OnCrystallizeHydro   // ターゲット, AttackEvent
	OnCrystallizeCryo    // ターゲット, AttackEvent
	OnCrystallizeElectro // ターゲット, AttackEvent
	OnCrystallizePyro    // ターゲット, AttackEvent
	OnAggravate          // ターゲット, AttackEvent
	OnSpread             // ターゲット, AttackEvent
	OnQuicken            // ターゲット, AttackEvent
	OnBloom              // ターゲット, AttackEvent
	OnHyperbloom         // ターゲット, AttackEvent
	OnBurgeon            // ターゲット, AttackEvent
	OnBurning            // ターゲット, AttackEvent
	OnLunarCharged       // ターゲット, AttackEvent
	OnLunarBloom         // ターゲット, AttackEvent
	OnLunarCrystallize   // ターゲット, AttackEvent
	OnShatter            // target, AttackEvent; 通常は元素反応とみなされないため、全反応イベントの購読を簡素化するため末尾に配置
	ReactionEventEndDelim
	OnDendroCore // Gadget
	// その他
	OnStamUse           // アビリティ
	OnShielded          // シールド
	OnShieldBreak       // シールド破壊
	OnConstructSpawned  // nil
	OnCharacterSwap     // 前, 次
	OnParticleReceived  // particle
	OnEnergyChange      // character_received, pre_energy, energy_change, src (後のenergyはcharacter_receivedで取得可), is_particle (boolean)
	OnEnergyBurst       // character_drained, pre_energy, burst_cost
	OnTargetDied        // ターゲット, AttackEvent
	OnTargetMoved       // ターゲット
	OnCharacterHit      // nil <- キャラクターが攻撃を受けるがシールドでダメージを防ぐ可能性がある場合
	OnCharacterHurt     // ダメージ量
	OnHPDebt            // 対象キャラクター, 量
	OnHeal              // 回復元キャラクター, 対象キャラクター, 量, 過剰回復, 負債前の量
	OnPlayerPreHPDrain  // 変更可能なDrainInfo
	OnPlayerHPDrain     // DrainInfo
	OnNightsoulBurst    // ターゲット, AttackEvent
	OnNightsoulGenerate // キャラクター, 量
	OnNightsoulConsume  // キャラクター, 量
	OnVerdantDewGain    // 付与量, 累計カウント (パーティレベルのVerdant Dew)
	// アビリティ使用
	OnActionFailed // ActiveCharIndex, action.Action, param, action.ActionFailure
	OnActionExec   // ActiveCharIndex, action.Action, param
	OnSkill        // nil
	OnBurst        // nil
	OnAttack       // nil
	OnChargeAttack // nil
	OnPlunge       // nil
	OnAimShoot     // nil
	OnDash
	// シミュレーション関連
	OnInitialize  // nil
	OnStateChange // 前, 次
	OnEnemyAdded  // t
	OnTick
	OnInfusion             // index, ele, duration
	OnSimEndedSuccessfully // nil
	EndEventTypes          // 終端
)

type Handler struct {
	events [][]ehook
}

type Hook func(args ...interface{}) bool

type Eventter interface {
	Subscribe(e Event, f Hook, key string)
	Unsubscribe(e Event, key string)
	Emit(e Event, args ...interface{})
}

type ehook struct {
	f   Hook
	key string
}

func New() *Handler {
	h := &Handler{
		events: make([][]ehook, EndEventTypes),
	}

	for i := range h.events {
		h.events[i] = make([]ehook, 0, 10)
	}

	return h
}

func (h *Handler) Subscribe(e Event, f Hook, key string) {
	a := h.events[e]

	evt := ehook{
		f:   f,
		key: key,
	}

	// 上書きがあるか先に確認
	ind := -1
	for i, v := range a {
		if v.key == key {
			ind = i
		}
	}
	if ind > -1 {
		a[ind] = evt
	} else {
		a = append(a, evt)
	}
	h.events[e] = a
}

func (h *Handler) Unsubscribe(e Event, key string) {
	n := 0
	for _, v := range h.events[e] {
		if v.key != key {
			h.events[e][n] = v
			n++
		}
	}
	h.events[e] = h.events[e][:n]
}

func (h *Handler) Emit(e Event, args ...interface{}) {
	n := 0
	for _, v := range h.events[e] {
		if !v.f(args...) {
			h.events[e][n] = v
			n++
		}
	}
	h.events[e] = h.events[e][:n]
}
