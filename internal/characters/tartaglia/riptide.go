package tartaglia

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

// 狙いを定めると、矢先に水元素の力が蓄積される。
// 激流でフルチャージされた矢は水元素ダメージを与え、断流を付与する。
func (c *char) aimedApplyRiptide(a combat.AttackCB) {
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	// 攻撃が水元素を持つか、元素スキルが発動中で物理の場合に付与
	// カバー範囲:
	// - フルチャージ狙い撃ちは元素スキル状態に関係なく付与
	// - 物理狙い撃ちは元素スキル状態中のみ付与
	if a.AttackEvent.Info.Element == attributes.Hydro || (c.StatusIsActive(meleeKey) && a.AttackEvent.Info.Element == attributes.Physical) {
		c.applyRiptide("aimed shot", t)
	}
}

// 水元素が込められた魔法の矢を素早く放ち、範囲水元素ダメージを与えて断流を付与する。
func (c *char) rangedBurstApplyRiptide(a combat.AttackCB) {
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	c.applyRiptide("ranged burst", t)
}

func (c *char) applyRiptide(src string, t *enemy.Enemy) {
	if c.Base.Cons >= 4 && !t.StatusIsActive(riptideKey) {
		c.c4Src = c.Core.F
		t.QueueEnemyTask(c.rtC4Tick(c.Core.F, t), 60*3.9)
	}

	t.AddStatus(riptideKey, c.riptideDuration, true)
	c.Core.Log.NewEvent(
		fmt.Sprintf("riptide applied (%v)", src),
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("target", t.Key()).
		Write("expiry", t.StatusExpiry(riptideKey))
}

// タルタリヤが近接スタンスの場合、断流状態の敵に対して4秒毎に断流・斉しをトリガー。それ以外の場合は断流・フラッシュをトリガー。
// この命ノ星座効果はICDの影響を受けない。
func (c *char) rtC4Tick(src int, t *enemy.Enemy) func() {
	return func() {
		if c.c4Src != src {
			c.Core.Log.NewEvent("tartaglia c4 src check ignored, src diff", glog.LogCharacterEvent, c.Index).
				Write("src", src).
				Write("new src", c.c4Src)
			return
		}
		if !t.StatusIsActive(riptideKey) {
			return
		}

		if c.StatusIsActive(meleeKey) {
			c.rtSlashTick(t)
		} else {
			c.rtFlashTick(t)
		}

		t.QueueEnemyTask(c.rtC4Tick(src, t), 60*3.9)
		c.Core.Log.NewEvent("tartaglia c4 applied", glog.LogCharacterEvent, c.Index).
			Write("src", src).
			Write("target", t.Key())
	}
}

// 断流・フラッシュ: 断流状態の敵にフルチャージ狙い撃ちが命中すると、
// 範囲ダメージを連続で与える。0.7秒毎に1回発動可能。
func (c *char) rtFlashCallback(a combat.AttackCB) {
	// 実際に敵であることを確認
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	// 攻撃が水元素を持たない場合は何もしない
	// - 物理重撃からのトリガーを防止
	if a.AttackEvent.Info.Element != attributes.Hydro {
		return
	}
	// ターゲットに断流がなければ何もしない
	if !t.StatusIsActive(riptideKey) {
		return
	}
	// フラッシュがまだICD中なら何もしない
	if t.StatusIsActive(riptideFlashICDKey) {
		return
	}
	// 0.7秒のICDを追加
	t.AddStatus(riptideFlashICDKey, 42, true)

	c.rtFlashTick(t)
}

func (c *char) rtFlashTick(t *enemy.Enemy) {
	// ダメージをキューに追加
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Riptide Flash",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagTartagliaRiptideFlash,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       rtFlash[c.TalentLvlAttack()],
	}

	// 3ヒットを発動
	for i := 1; i <= 3; i++ {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 3), 1, 1, c.particleCB)
	}

	c.Core.Log.NewEvent(
		"riptide flash triggered",
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("dur", c.StatusExpiry(meleeKey)-c.Core.F).
		Write("target", t.Key()).
		Write("riptide_flash_icd", t.StatusExpiry(riptideFlashICDKey)).
		Write("riptide_expiry", t.StatusExpiry(riptideKey))
}

// 断流状態の敵に近接攻撃が命中すると、範囲水元素ダメージの断流・斉しが発動する。
// このダメージは元素スキルダメージとみなされ、1.5秒毎に1回のみ発動可能。
func (c *char) rtSlashCallback(a combat.AttackCB) {
	// 実際に敵であることを確認
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	// 元素スキルが発動していない場合は何もしない
	// - 元素スキル状態中に敵に命中した狙い撃ちでもトリガー可能
	if !c.StatusIsActive(meleeKey) {
		return
	}
	// ターゲットに断流がなければ何もしない
	if !t.StatusIsActive(riptideKey) {
		return
	}
	// 斉しがまだICD中なら何もしない
	if t.StatusIsActive(riptideSlashICDKey) {
		return
	}
	// 1.5秒のICDを追加
	t.AddStatus(riptideSlashICDKey, 90, true)

	c.rtSlashTick(t)
}

func (c *char) rtSlashTick(t *enemy.Enemy) {
	// 攻撃をトリガー
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Riptide Slash",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       rtSlash[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 3), 1, 1, c.particleCB)

	c.Core.Log.NewEvent(
		"riptide slash ticked",
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("dur", c.StatusExpiry(meleeKey)-c.Core.F).
		Write("target", t.Key()).
		Write("riptide_slash_icd", t.StatusExpiry(riptideSlashICDKey)).
		Write("riptide_expiry", t.StatusExpiry(riptideKey))
}

// 殿滅の水が断流状態の敵に命中すると、断流を解除し、
// 範囲水元素ダメージの水爆発をトリガーする。このダメージは元素爆発ダメージとみなされる。
func (c *char) rtBlastCallback(a combat.AttackCB) {
	// 実際に敵であることを確認
	t, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	// ターゲットが断流状態の場合のみトリガー
	if !t.StatusIsActive(riptideKey) {
		return
	}
	// TODO: これは斉しとICDを共有している？？？
	if t.StatusIsActive(riptideSlashICDKey) {
		return
	}
	// ダメージをキューに追加
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Riptide Blast",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 50,
		Mult:       rtBlast[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 5), 1, 1)

	c.Core.Log.NewEvent(
		"riptide blast triggered",
		glog.LogCharacterEvent,
		c.Index,
	).
		Write("dur", c.StatusExpiry(meleeKey)-c.Core.F).
		Write("target", t.Key()).
		Write("rtExpiry", t.StatusExpiry(riptideKey))

	// 断流ステータスを解除
	t.DeleteStatus(riptideKey)
}

// 断流・破裂: 断流状態の敵を撃破すると、水爆発が発生し、
// 周囲の命中した敵に断流を付与する。
// タルタリヤの断流・破裂と2命の敵撃破時効果を処理
func (c *char) onDefeatTargets() {
	c.Core.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		// 敵でなければ何もしない
		if !ok {
			return false
		}
		// ターゲットに断流がなければ何もしない
		if !t.StatusIsActive(riptideKey) {
			return false
		}
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Riptide Burst",
				AttackTag:  attacks.AttackTagNormal,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeSlash,
				Element:    attributes.Hydro,
				Durability: 50,
				Mult:       rtBurst[c.TalentLvlAttack()],
			}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 5), 0, 0)
		}, 5)
		// TODO: 必要に応じて断流の有効期限フレーム配列を再インデックスする
		if c.Base.Cons >= 2 {
			c.AddEnergy("tartaglia-c2", 4)
		}
		return false
	}, "tartaglia-on-enemy-death")
}
