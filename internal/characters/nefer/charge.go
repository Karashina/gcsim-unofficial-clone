package nefer

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var (
	chargeFrames     []int
	phantasmFrames   []int
	phantasmHitmarks = []int{28, 81, 48, 56, 81} // 2 Nefer hits + 3 Shade hits
)

const (
	chargeHitmark = 45
)

func init() {
	chargeFrames = frames.InitAbilSlice(60)
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark

	phantasmFrames = frames.InitAbilSlice(92)
	phantasmFrames[action.ActionDash] = 28
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// 幻影公演を使用すべきか確認
	if c.StatusIsActive(skillKey) && c.Core.Player.Verdant.Count() >= 1 {
		return c.phantasmPerformance(p)
	}

	// 通常重撃（Slither）：前方移動とスタミナ消費、離脱時に草元素ダメージ。Shadow Danceは最大消費量を低下させる。

	if p["hold"] > 0 {
		// Slither長押し - スタミナドレインメカニクスを設定
		dur := p["hold"]
		chargeFrames[action.ActionSkill] = 1 // Neferがスリザー状態中に元素スキルを使用しても、その状態が解除されることはない。
		prevbonus := c.Core.Player.Verdant.GetGainBonus()
		frameremaining := c.Core.Player.Verdant.RemainingFrames()

		// A4相互作用：Shadow Dance中の翠露獲得を強化
		if c.StatusIsActive(skillKey) {
			emBonus := 1 + min(0.001*float64(max(0, c.Stat(attributes.EM)-500)), 0.5)
			c.Core.Player.Verdant.SetGainBonus(prevbonus + emBonus)
			if frameremaining < dur {
				c.Core.Player.Verdant.StartCharge(dur - frameremaining)
			}
			c.Core.Tasks.Add(func() {
				c.Core.Player.Verdant.SetGainBonus(prevbonus)
			}, dur)
		}
	}

	// 重撃ダメージを与える
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3.5),
		chargeHitmark,
		chargeHitmark,
	)

	// 重撃使用時に範囲内の欺きの種を吸収を試みる
	c.absorbSeeds(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) phantasmPerformance(_ map[string]int) (action.Info, error) {
	/*
		NeferがShadow Dance状態でパーティが翠露を1以上持っている時、Neferの重撃は特殊重撃「幻影公演」に置き換わり、スタミナを消費しない。
		Neferは「幻影公演 n-Hitダメージ（Nefer）」を2回、「幻影公演 n-Hitダメージ（影）」を3回与える。影のダメージはLunar-Bloomダメージとみなされる。
		幻影公演 1-Hitダメージ（影）使用後、翠露が1消費される。
	*/

	// 幻影公演：Neferの2ヒット（ATKスケーリング）と3影ヒット（Lunar-Bloom）。最初の影ヒット後に翠露1消費。

	// 幻影公演発動時に欺きの種を吸収
	c.absorbSeeds(6)
	c.QueueCharTask(func() {
		aiATK := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Phantasm Performance 1-Hit (Nefer / C)",
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagExtraAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    skillppn1atk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillppn1em[c.TalentLvlSkill()]*c.Stat(attributes.EM),
		}
		c.Core.QueueAttack(
			aiATK,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 4),
			0, 0,
			c.makePhantasmBonus(),
		)

	}, phantasmHitmarks[0])

	// Neferの第2ヒット（ATK）または6凸のLunar-Bloom変換
	c.QueueCharTask(func() {
		if c.Base.Cons >= 6 {
			// 6凸：第2ヒットを元素熟知スケーリングのLunar-Bloomダメージに変換
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Nefer C6 2nd Dummy (C)",
				FlatDmg:    0,
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
				0, 0,
			)
		} else {
			// 通常の第2ヒット
			aiATK := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Phantasm Performance 2-Hit (Nefer / C)",
				AttackTag:  attacks.AttackTagExtra,
				ICDTag:     attacks.ICDTagExtraAttack,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeDefault,
				Element:    attributes.Dendro,
				Durability: 25,
				FlatDmg:    skillppn2atk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillppn2em[c.TalentLvlSkill()]*c.Stat(attributes.EM),
			}
			c.Core.QueueAttack(
				aiATK,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 4),
				0, 0,
				c.makePhantasmBonus(),
			)
		}
	}, phantasmHitmarks[1])

	// 影ヒット（ダミー） -> フック経由でLunar-Bloomに解決、最初の影ヒット後に翠露を消費
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP1Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
		// 翠露1消費
		c.Core.Player.Verdant.Consume(1)
	}, phantasmHitmarks[2])

	// 影 2（ダミー）
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP2Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
	}, phantasmHitmarks[3])

	// 影 3（ダミー）
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Nefer PP3Shade Dummy (C)",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
			0, 0,
		)
	}, phantasmHitmarks[4])

	// 6凸：PP後の追加ダミーヒット
	if c.Base.Cons >= 6 {
		c.QueueCharTask(func() {
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Nefer C6 Extra Dummy (C)",
				FlatDmg:    0,
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99),
				0, 0,
			)
		}, phantasmHitmarks[4]+5)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(phantasmFrames),
		AnimationLength: phantasmFrames[action.InvalidAction],
		CanQueueAfter:   phantasmHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}

// 幻影公演に偽りのヴェールボーナスを適用
func (c *char) makePhantasmBonus() combat.AttackCBFunc {
	bonus := c.a1count * 0.08 // 各スタックがダメージを8%増加
	if bonus == 0 {
		return nil
	}
	return func(a combat.AttackCB) {
		// ボーナスを総ダメージの割合増加として適用
		a.AttackEvent.Info.FlatDmg *= (1 + bonus)
	}
}

// absorbSeedsはプレイヤー位置から半径内の欺きの種（IsSeed=trueのDendroCore）を探して吸収する
func (c *char) absorbSeeds(radius float64) {
	absorbed := 0
	player := c.Core.Combat.Player()
	for _, g := range c.Core.Combat.Gadgets() {
		if g == nil {
			continue
		}
		if g.GadgetTyp() != combat.GadgetTypDendroCore {
			continue
		}
		if dc, ok := g.(*reactable.DendroCore); ok {
			if !dc.IsSeed {
				continue
			}
			// 距離チェック
			if dc.Pos().Distance(player.Pos()) <= radius {
				absorbed++
				// ガジェットを削除
				dc.Gadget.Kill()
			}
		}
	}
	if absorbed == 0 {
		return
	}

	prev := c.a1count
	maxStacks := 3.0
	if c.Base.Cons >= 2 {
		maxStacks = 5.0
	}
	c.a1count = math.Min(maxStacks, c.a1count+float64(absorbed))

	// スタックごとの持続時間（9秒）を追加し、持続時間を更新
	c.AddStatus("veil-of-falsehood", 9*60, true)

	// 3スタック以上または3スタック目の持続時間更新時に元素熟知ボーナスを付与
	if c.a1count >= 3 || (prev >= 3 && absorbed > 0) {
		emBonus := 100.0
		if c.Base.Cons >= 2 && c.a1count >= 5 {
			emBonus = 200.0
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("veil-em-bonus", 8*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.EM] = emBonus
				return m, true
			},
		})
	}
}
