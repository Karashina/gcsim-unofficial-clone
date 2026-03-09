package verdict

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
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.Verdict, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	buffKey           = "verdict-skill-dmg"
	buffDuration      = 15 * 60
	dmgWindowKey      = "verdict-dmg-window"
	dmgWindowDuration = 0.2 * 60
)

// 攻撃力が20/25/30/35/40%増加。パーティメンバーが結晶化反応で元素欠片を獲得すると、
// 装備キャラクターはシール1つを獲得し、元素スキルダメージが18/22.5/27/31.5/36%増加。
// シールは15秒間持続し、装備者は同時に最大2つのシールを所持可能。
// 装備者の元素スキルがダメージを与えた0.2秒後に全シールが消滅する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.15 + float64(r)*0.05
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("verdict-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	// 結晶化シャード取得でシール獲得
	c.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		// シールドをチェック
		shd := args[0].(shield.Shield)
		if shd.Type() != shield.Crystallize {
			return false
		}

		if !char.StatModIsActive(buffKey) {
			w.stacks = 0
		}
		if w.stacks < 2 {
			w.stacks++
		}
		c.Log.NewEvent("verdict adding stack", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		char.AddStatus(buffKey, buffDuration, true)
		return false
	}, fmt.Sprintf("verdict-seal-%v", char.Base.Key.String()))

	// 紋章アクティブ中のスキルダメージ増加
	skillDmg := 0.135 + float64(r)*0.045
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		// バフ中でなければ何もしない
		if !char.StatusIsActive(buffKey) {
			return false
		}
		// 初回発動の場合
		// - ダメージウィンドウの持続時間を設定
		// - ウィンドウ終了時にバフをリセット
		if !char.StatusIsActive(dmgWindowKey) {
			char.AddStatus(dmgWindowKey, dmgWindowDuration, true)
			char.QueueCharTask(func() {
				char.DeleteStatus(buffKey)
				w.stacks = 0
			}, dmgWindowDuration)
		}
		skillDmgAdd := skillDmg * float64(w.stacks)
		atk.Snapshot.Stats[attributes.DmgP] += skillDmgAdd

		c.Log.NewEvent("verdict adding skill dmg", glog.LogPreDamageMod, char.Index).
			Write("skill_dmg_added", skillDmgAdd)
		return false
	}, fmt.Sprintf("verdict-onhit-%v", char.Base.Key.String()))

	return w, nil
}
