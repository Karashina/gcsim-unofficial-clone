// パッケージ core はシミュレーションの中核機能を提供する:
//   - 戦闘
//   - タスク
//   - イベント処理
//   - ログ
//   - 設置物（本来は汎用オブジェクトにすべき？）
//   - ステータス
package core

import (
	"fmt"
	"math/rand"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/construct"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/status"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
)

type Core struct {
	F     int
	Flags Flags
	Seed  int64
	Rand  *rand.Rand
	// core の各機能
	Log        glog.Logger    // 1回を除く全実行でnilロガーを渡せるようインターフェースを使用
	Events     *event.Handler // イベント管理: 購読/解除/発行
	Status     *status.Handler
	Tasks      *task.Handler
	Combat     *combat.Handler
	Constructs *construct.Handler
	Player     *player.Handler
	// LCクラウドのグローバルシングルトン: 一度に1つのReactableだけがクラウドを保持可能
	ActiveLCCloud interface{}
}

type Flags struct {
	LogDebug          bool // ログレベルの判定に使用
	DamageMode        bool // HPモード用
	DefHalt           bool // ヒットラグ用
	EnableHitlag      bool // ヒットラグ有効化
	IgnoreBurstEnergy bool // 元素爆発使用時のエネルギー消費無視用
	Custom            map[string]float64
}

type Reactable interface {
	React(a *combat.AttackEvent)
	AuraContains(e ...attributes.Element) bool
	Tick()
}

// type Enemy interface {
// 	AddResistMod(key string, dur int, ele attributes.Element, val float64)
// 	DeleteResistMod(key string)
// 	ResistModIsActive(key string) bool
// 	AddDefMod(key string, dur int, val float64)
// 	DeleteDefMod(key string)
// 	DefModIsActive(key string) bool
// }

const MaxTeamSize = 4

type Opt struct {
	Seed              int64
	Debug             bool
	EnableHitlag      bool
	DefHalt           bool
	DamageMode        bool
	IgnoreBurstEnergy bool
	Delays            info.Delays
}

func New(opt Opt) (*Core, error) {
	c := &Core{}
	c.Seed = opt.Seed
	c.Rand = rand.New(rand.NewSource(opt.Seed))
	c.Flags.Custom = make(map[string]float64)
	if opt.Debug {
		c.Log = glog.New(&c.F, 500)
		c.Flags.LogDebug = true
	} else {
		c.Log = &glog.NilLogger{}
	}

	c.Flags.DamageMode = opt.DamageMode
	c.Flags.DefHalt = opt.DefHalt
	c.Flags.EnableHitlag = opt.EnableHitlag
	c.Flags.IgnoreBurstEnergy = opt.IgnoreBurstEnergy
	c.Events = event.New()
	c.Status = status.New(&c.F, c.Log)
	c.Tasks = task.New(&c.F)
	c.Constructs = construct.New(&c.F, c.Log, c.Events)
	c.Player = player.New(
		player.Opt{
			F:            &c.F,
			Delays:       opt.Delays,
			Log:          c.Log,
			Events:       c.Events,
			Tasks:        c.Tasks,
			Debug:        opt.Debug,
			EnableHitlag: opt.EnableHitlag,
		},
	)
	c.Combat = combat.New(combat.Opt{
		Events:       c.Events,
		Team:         c.Player,
		Rand:         c.Rand,
		Debug:        c.Flags.LogDebug,
		Log:          c.Log,
		DamageMode:   c.Flags.DamageMode,
		DefHalt:      c.Flags.DefHalt,
		EnableHitlag: c.Flags.EnableHitlag,
		Tasks:        c.Tasks,
	})

	return c, nil
}

func (c *Core) Init() error {
	var err error
	// セットアップ順序
	//	- 元素共鳴
	//	- 命中時エネルギー
	//	- 基礎ステータス
	//	- キャラクター初期化
	//	- 初期化コールバック
	c.SetupOnNormalHitEnergy()
	err = c.Player.InitializeTeam()
	if err != nil {
		return err
	}
	c.Events.Emit(event.OnInitialize)
	return nil
}

func (c *Core) Tick() error {
	// Tick処理対象:
	//	- ターゲット
	//	- 設置物
	//	- プレイヤー（スタミナ, 交代, アニメーション等…）
	//		- キャラクター
	//		- シールド
	//		- アニメーション
	//		- スタミナ
	//		- 交代
	//	- タスク
	//TODO: ここでエラーをチェックすべき？
	c.Combat.Tick()
	c.Constructs.Tick()
	c.Player.Tick()
	c.Tasks.Run()
	return nil
}

func (c *Core) AddChar(p info.CharacterProfile) (int, error) {
	var err error

	// キャラクターを初期化
	char, err := character.New(p, &c.F, c.Flags.LogDebug, c.Log, c.Events, c.Tasks)
	if err != nil {
		return -1, err
	}

	f, ok := NewCharFuncMap[p.Base.Key]
	if !ok {
		return -1, fmt.Errorf("invalid character: %v", p.Base.Key.String())
	}
	err = f(c, char, p)
	if err != nil {
		return -1, err
	}
	index := c.Player.AddChar(char)

	// 初期HPを取得
	char.StartHP = -1
	if hp, ok := p.Params["start_hp"]; ok {
		char.StartHP = hp
	}
	char.StartHPRatio = -1
	if hpRatio, ok := p.Params["start_hp%"]; ok {
		char.StartHPRatio = hpRatio
	}

	// エネルギーを設定
	char.Energy = char.EnergyMax
	if e, ok := p.Params["start_energy"]; ok {
		char.Energy = float64(e)
		// ユーザーが energy = 10000000 に設定した場合のサニティチェック
		if char.Energy > char.EnergyMax {
			char.Energy = char.EnergyMax
		}
	}

	// 武器を初期化
	wf, ok := weaponMap[p.Weapon.Key]
	if !ok {
		return -1, fmt.Errorf("unrecognized weapon %v for character %v", p.Weapon.Key, p.Base.Key.String())
	}
	weap, err := wf(c, char, p.Weapon)
	if err != nil {
		return -1, err
	}
	char.SetWeapon(weap)

	// セットボーナスを設定
	total := 0
	for key, count := range p.Sets {
		total += count
		af, ok := setMap[key]
		if ok {
			s, err := af(c, char, count, p.SetParams[key])
			if err != nil {
				return -1, err
			}
			char.SetArtifactSet(key, s)
		} else {
			return -1, fmt.Errorf("character %v has unrecognized artifact: %v", p.Base.Key.String(), key)
		}
	}
	//TODO: これはパーサーで処理すべき
	if total > 5 {
		return -1, fmt.Errorf("total set count cannot exceed 5, got %v", total)
	}

	return index, nil
}
