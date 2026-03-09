package character

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

type Character interface {
	Base
	HP

	Init() error // 各キャラクターに組み込まれた変数設定用の初期化関数

	Attack(p map[string]int) (action.Info, error)
	Aimed(p map[string]int) (action.Info, error)
	ChargeAttack(p map[string]int) (action.Info, error)
	HighPlungeAttack(p map[string]int) (action.Info, error)
	LowPlungeAttack(p map[string]int) (action.Info, error)
	Skill(p map[string]int) (action.Info, error)
	Burst(p map[string]int) (action.Info, error)
	Dash(p map[string]int) (action.Info, error)
	Walk(p map[string]int) (action.Info, error)
	Jump(p map[string]int) (action.Info, error)

	ActionStam(a action.Action, p map[string]int) float64

	ActionReady(a action.Action, p map[string]int) (bool, action.Failure)
	NextQueueItemIsValid(targetChar keys.Char, a action.Action, p map[string]int) error
	SetCD(a action.Action, dur int)
	Cooldown(a action.Action) int
	ResetActionCooldown(a action.Action)
	ReduceActionCooldown(a action.Action, v int)
	Charges(a action.Action) int

	Snapshot(a *combat.AttackInfo) combat.Snapshot

	AddEnergy(src string, amt float64)

	ApplyHitlag(factor, dur float64)
	AnimationStartDelay(model.AnimationDelayKey) int

	Condition([]string) (any, error)

	ResetNormalCounter()
	NextNormalCounter() int
}

// Base はキャラクターの基本情報を含む
type Base interface {
	Data() *model.AvatarData
}

// HP はキャラクターHPに関する情報とヘルパーを含む
type HP interface {
	CurrentHPRatio() float64
	CurrentHP() float64
	CurrentHPDebt() float64
	CurrentHPDebtRatio() float64

	SetHPByAmount(float64)
	SetHPByRatio(float64)
	ModifyHPByAmount(float64)
	ModifyHPByRatio(float64)

	ModifyHPDebtByAmount(float64)
	ModifyHPDebtByRatio(float64)

	Heal(*info.HealInfo) (float64, float64) // 実際に回復したHP量と解消されたHP負債量を返す
	Drain(*info.DrainInfo) float64

	ReceiveHeal(*info.HealInfo, float64) float64
}

type CharWrapper struct {
	Index int
	f     *int // 現在のフレーム
	debug bool // debug mode?
	Character
	events event.Eventter
	log    glog.Logger
	tasks  task.Tasker

	// 基本特性
	Base              info.CharacterBase
	Weapon            info.WeaponProfile
	Talents           info.TalentProfile
	CharZone          info.ZoneType
	CharBody          info.BodyType
	NormalCon         int
	SkillCon          int
	BurstCon          int
	HasArkhe          bool
	MoonsignNascent   bool
	MoonsignAscendant bool

	Equip struct {
		Weapon info.Weapon
		Sets   map[keys.Set]info.Set
	}

	// 現在のステータス
	ParticleDelay int // キャラクター固有の粒子遅延
	Energy        float64
	EnergyMax     float64
	// チーム初期化時に追加される HP mod の影響を初期 HP に与えないために必要
	StartHP      int
	StartHPRatio int

	// 通常攻撃カウンター
	NormalHitNum  int // 通常攻撃コンボのヒット数
	NormalCounter int

	// タグ
	Tags      map[string]int
	BaseStats [attributes.EndStatType]float64

	// 修飾子
	mods []modifier.Mod

	// ダッシュ CD: フィールド外キャラの残り CD フレームを追跡
	RemainingDashCD int
	DashLockout     bool

	// ヒットラグ関連
	TimePassed   int // シミュレーション開始からの経過フレーム数
	frozenFrames int // 凍結中の残りフレーム数
	queue        *task.Handler
}

func New(
	p info.CharacterProfile,
	f *int, // 現在のフレーム
	debug bool, // are we running in debug mode
	log glog.Logger, // logging, can be nil
	events event.Eventter, // event emitter
	tasker task.Tasker,
) (*CharWrapper, error) {
	c := &CharWrapper{
		Base:          p.Base,
		Weapon:        p.Weapon,
		Talents:       p.Talents,
		ParticleDelay: 100, // デフォルトの粒子遅延
		log:           log,
		events:        events,
		tasks:         tasker,
		Tags:          make(map[string]int),
		mods:          make([]modifier.Mod, 0, 20),
		f:             f,
		debug:         debug,
	}
	c.queue = task.New(&c.TimePassed)
	s := (*[attributes.EndStatType]float64)(p.Stats)
	c.BaseStats = *s
	c.Equip.Sets = make(map[keys.Set]info.Set)

	// デフォルトで -1 に設定し、各キャラが normal/skill/burst の命の座を指定する
	c.NormalCon = -1
	c.SkillCon = -1
	c.BurstCon = -1

	// 天賦レベルを検証
	if c.Talents.Attack < 1 || c.Talents.Attack > 10 {
		return nil, fmt.Errorf("invalid talent lvl: attack - %v", c.Talents.Attack)
	}
	if c.Talents.Skill < 1 || c.Talents.Skill > 10 {
		return nil, fmt.Errorf("invalid talent lvl: skill - %v", c.Talents.Skill)
	}
	if c.Talents.Burst < 1 || c.Talents.Burst > 10 {
		return nil, fmt.Errorf("invalid talent lvl: burst - %v", c.Talents.Burst)
	}

	return c, nil
}

// HasLCCloudOn は指定されたターゲットに現在アクティブな LC Cloud があるかを確認する。
// キャラクターコードから任意のターゲットの LC Cloud 状態を確認できる。
func (c *CharWrapper) HasLCCloudOn(target combat.Target) bool {
	// ターゲットから Reactable インターフェースの取得を試みる
	type hasLCCloud interface {
		HasLCCloud() bool
	}
	if r, ok := target.(hasLCCloud); ok {
		return r.HasLCCloud()
	}
	return false
}

func (c *CharWrapper) SetIndex(index int) {
	c.Index = index
}

func (c *CharWrapper) SetWeapon(w info.Weapon) {
	c.Equip.Weapon = w
}

func (c *CharWrapper) SetArtifactSet(key keys.Set, set info.Set) {
	c.Equip.Sets[key] = set
}

func (c *CharWrapper) Tag(key string) int {
	return c.Tags[key]
}

func (c *CharWrapper) SetTag(key string, val int) {
	c.Tags[key] = val
}

func (c *CharWrapper) RemoveTag(key string) {
	delete(c.Tags, key)
}

func (c *CharWrapper) consCheck() {
	consUnset := 0
	if c.NormalCon < 0 {
		consUnset++
	}
	if c.SkillCon < 0 {
		consUnset++
	}
	if c.BurstCon < 0 {
		consUnset++
	}
	if consUnset != 1 {
		panic(fmt.Sprintf("cons not set properly for %v, please set two out of three values:\nNormalCon: %v\nSkillCon: %v\nBurstCon: %v", c.Base.Key.String(), c.NormalCon, c.SkillCon, c.BurstCon))
	}
}

func (c *CharWrapper) TalentLvlAttack() int {
	c.consCheck()
	add := -1
	if c.Tags[keys.ChildePassive] > 0 {
		add++
	}
	if c.NormalCon > 0 && c.Base.Cons >= c.NormalCon {
		add += 3
	}
	if add >= 4 {
		add = 4
	}
	return c.Talents.Attack + add
}
func (c *CharWrapper) TalentLvlSkill() int {
	c.consCheck()
	add := -1
	if c.Tags[keys.SkirkPassive] > 0 {
		add++
	}
	if c.SkillCon > 0 && c.Base.Cons >= c.SkillCon {
		add += 3
	}
	if add >= 4 {
		add = 4
	}
	return c.Talents.Skill + add
}
func (c *CharWrapper) TalentLvlBurst() int {
	c.consCheck()
	add := -1
	if c.BurstCon > 0 && c.Base.Cons >= c.BurstCon {
		add += 3
	}
	if add >= 4 {
		add = 4
	}
	return c.Talents.Burst + add
}

type Particle struct {
	Source string
	Num    float64
	Ele    attributes.Element
}
