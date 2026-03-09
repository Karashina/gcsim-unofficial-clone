package yoimiya

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

var burstFrames []int

const burstHitmark = 75
const abDebuff = "aurous-blaze"
const abIcdKey = "aurous-blaze-icd"

func init() {
	burstFrames = frames.InitAbilSlice(113) // Q -> N1
	burstFrames[action.ActionSkill] = 112   // Q -> E
	burstFrames[action.ActionDash] = 111    // Q -> D
	burstFrames[action.ActionJump] = 112    // Q -> J
	burstFrames[action.ActionSwap] = 109    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// アニメーション終了時にスキルダメージを与えると仮定
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Aurous Blaze",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   220,
		Element:    attributes.Pyro,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}

	if c.Base.Ascension >= 4 {
		c.Core.Tasks.Add(c.a4, burstHitmark)
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6),
		0,
		burstHitmark,
		c.applyAB, // Aurous Blazeを適用するコールバック
		c.makeC2CB(),
	)

	// シミュレーションにクールダウンを追加
	c.SetCD(action.ActionBurst, 15*60)
	// エネルギーを消費
	c.ConsumeEnergy(5)

	c.abApplied = false

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) applyAB(a combat.AttackCB) {
	// 初撃命中後に敵にマーカーを付与
	// バウンス処理は無視（常にターゲット0と仮定）
	// ICD 2秒、擃破時に削除

	// 既に敵に過蒸の炎が適用済みなら何もしない
	if c.abApplied {
		return
	}

	trg, ok := a.Target.(*enemy.Enemy)
	// 敵でなければ何もしない
	if !ok {
		return
	}
	c.abApplied = true

	duration := 600
	if c.Base.Cons >= 1 {
		duration = 840
	}
	trg.AddStatus(abDebuff, duration, true) // Aurous Blazeを適用
}

func (c *char) burstHook() {
	// ターゲット0の攻撃着弾時にチェック
	// 過蒸の炎がアクティブならCDでない場合にダメージを発動
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		trg, ok := args[0].(*enemy.Enemy)
		// 敵でなければ無視
		if !ok {
			return false
		}
		// 敵にデバフがない場合は無視
		if !trg.StatusIsActive(abDebuff) {
			return false
		}
		// 自分自身の攻撃は無視
		if ae.Info.ActorIndex == c.Index {
			return false
		}
		// ICD中は無視
		if trg.StatusIsActive(abIcdKey) {
			return false
		}
		// 対象外の攻撃タグは無視
		switch ae.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		case attacks.AttackTagPlunge:
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalArtHold:
		case attacks.AttackTagElementalBurst:
		default:
			return false
		}
		// 爆発を実行し、ICDを設定
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Aurous Blaze (Explode)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   60,
			Element:    attributes.Pyro,
			Durability: 25,
			Mult:       burstExplode[c.TalentLvlBurst()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 1, c.makeC2CB())

		trg.AddStatus(abIcdKey, 120, true) // 過蒸の炎のICDを発動

		// 4凸
		if c.Base.Cons >= 4 {
			c.ReduceActionCooldown(action.ActionSkill, 72)
		}

		return false
	}, "yoimiya-burst-check")

	if c.Core.Flags.DamageMode {
		// 嬵宮が死亡した場合のチェックを追加
		c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(_ ...interface{}) bool {
			if c.CurrentHPRatio() <= 0 {
				// ターゲットから過蒸の炎を削除
				for _, x := range c.Core.Combat.Enemies() {
					trg := x.(*enemy.Enemy)
					if trg.StatusIsActive(abDebuff) {
						trg.DeleteStatus(abDebuff)
					}
				}
			}
			return false
		}, "yoimiya-died")
	}
}
