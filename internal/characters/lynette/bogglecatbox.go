package lynette

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
)

type BogglecatBox struct {
	*gadget.Gadget
	char        *char
	pos         geometry.Point
	vividTravel int
}

func (c *char) newBogglecatBox(vividTravel int) *BogglecatBox {
	b := &BogglecatBox{}

	player := c.Core.Combat.Player()
	b.pos = geometry.CalcOffsetPoint(
		player.Pos(),
		geometry.Point{Y: 1.8},
		player.Direction(),
	)

	// TODO: ヒットボックスの見積もりを再確認
	b.Gadget = gadget.New(c.Core, b.pos, 1, combat.GadgetTypBogglecatBox)
	b.char = c
	b.vividTravel = vividTravel

	b.Duration = burstDuration
	b.char.AddStatus(burstKey, b.Duration, false)

	// 初撃
	initialAI := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Magic Trick: Astonishing Shift",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(initialAI, combat.NewCircleHitOnTarget(player, geometry.Point{Y: 1.5}, 4.5), burstHitmark-burstSpawn, burstHitmark-burstSpawn)

	// ボグルキャットのティック
	bogglecatAI := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Bogglecat Box",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       bogglecat[c.TalentLvlBurst()],
	}
	// Tickをキューに追加
	for t := burstFirstTick - burstSpawn; t <= b.Duration; t += burstTickInterval {
		c.Core.QueueAttack(bogglecatAI, combat.NewCircleHitOnTarget(b.pos, nil, 6), t, t)
	}

	// 被撃時の吸収に加え、0.3秒ごとに元素吸収をチェック
	b.OnThinkInterval = b.absorbCheck
	b.ThinkInterval = 0.3 * 60

	b.Core.Log.NewEvent("Lynette Bogglecat Box added", glog.LogCharacterEvent, c.Index).Write("src", b.Src())

	return b
}

func (b *BogglecatBox) HandleAttack(atk *combat.AttackEvent) float64 {
	b.Core.Events.Emit(event.OnGadgetHit, b, atk)

	b.Core.Log.NewEvent(fmt.Sprintf("lynette bogglecat box hit by %s", atk.Info.Abil), glog.LogCharacterEvent, b.char.Index)

	if atk.Info.Durability <= 0 {
		return 0
	}
	atk.Info.Durability *= reactions.Durability(b.WillApplyEle(atk.Info.ICDTag, atk.Info.ICDGroup, atk.Info.ActorIndex))
	if atk.Info.Durability <= 0 {
		return 0
	}

	// 氷/炎/水/雷元素のみ接触可能
	switch atk.Info.Element {
	case attributes.Cryo:
	case attributes.Pyro:
	case attributes.Hydro:
	case attributes.Electro:
	default:
		return 0
	}

	b.absorbRoutine(atk.Info.Element)

	return 0
}

func (b *BogglecatBox) absorbRoutine(absorbedElement attributes.Element) {
	b.Core.Log.NewEvent(fmt.Sprintf("lynette bogglecat box came into contact with %s", absorbedElement.String()), glog.LogCharacterEvent, b.char.Index)

	// ヴィヴィッドショット
	vividShotAI := combat.AttackInfo{
		ActorIndex: b.char.Index,
		Abil:       "Vivid Shot",
		AttackTag:  attacks.AttackTagElementalBurst,
		// ElementalBurstMixであるべきだが、キャラが使用する他のICDタグと異なればよいので追加ICDタグは不要
		ICDTag:     attacks.ICDTagElementalBurstAnemo,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    absorbedElement, // 吸収した元素を使用
		Durability: 25,
		Mult:       vivid[b.char.TalentLvlBurst()],
	}
	// ヴィヴィッドショットをキューに追加
	for t := burstVividInterval; t <= b.Duration; t += burstVividInterval {
		b.Core.Tasks.Add(func() {
			// 元素爆発位置から15m以内のランダムな敵をターゲット
			enemy := b.Core.Combat.RandomEnemyWithinArea(combat.NewCircleHitOnTarget(b.pos, nil, 15), nil)
			// 1または2（第2命ノ星座）のヴィヴィッドショットをキューに追加
			for i := 0; i < b.char.vividCount; i++ {
				// TODO: スナップショットはここで正しい？
				b.Core.QueueAttack(vividShotAI, combat.NewCircleHitOnTarget(enemy, nil, 1), 0, b.vividTravel)
			}
		}, t)
	}

	// 固有天賦2を適用
	b.char.a4(b.Duration)

	// 接触後は被弾対象でないためガジェットを削除
	b.Kill()
}

func (b *BogglecatBox) absorbCheck() {
	absorbedElement := b.Core.Combat.AbsorbCheck(b.char.Index, combat.NewCircleHitOnTarget(b.pos, nil, 0.48), attributes.Cryo, attributes.Pyro, attributes.Hydro, attributes.Electro)
	if absorbedElement == attributes.NoElement {
		return
	}
	b.absorbRoutine(absorbedElement)
}

func (b *BogglecatBox) SetDirection(trg geometry.Point) {}
func (b *BogglecatBox) SetDirectionToClosestEnemy()     {}
func (b *BogglecatBox) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}
