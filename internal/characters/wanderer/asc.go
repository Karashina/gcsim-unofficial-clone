package wanderer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a4Key           = "wanderer-a4"
	a4Prevent       = "wanderer-a4-prevent"
	a4IcdKey        = "wanderer-a4-icd"
	a1ElectroKey    = "wanderer-a1-electro"
	a1ElectroIcdKey = "wanderer-a1-electro-icd"
	a1PyroKey       = "wanderer-a1-pyro"
	a1CryoKey       = "wanderer-a1-cryo"
)

// 放浪者が「風の恵み」状態で「空居・風崇弾」または「空居・風柄打」で敵に命中した際、
// 16%の確率で「随風」効果を獲得する。
// この効果が発生しなかった攻撃ごとに、次の攻撃での発動確率が12%増加する。
// 効果発動の判定は0.1秒ごとに1回行われる。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if !c.StatusIsActive(skillKey) || c.StatusIsActive(a4Key) || c.StatusIsActive(a4IcdKey) {
			return
		}

		c.AddStatus(a4IcdKey, 6, true)

		if c.Core.Rand.Float64() > c.a4Prob {
			c.a4Prob += 0.12
			return
		}

		c.Core.Log.NewEvent("wanderer-a4 available", glog.LogCharacterEvent, c.Index).
			Write("probability", c.a4Prob)

		c.a4Prob = 0.16

		c.AddStatus(a4Key, 20*60, true)

		if c.Core.Player.CurrentState() == action.DashState {
			c.a4()
			return
		}
	}
}

// 放浪者が「風の恵み」状態で次に空中加速する際、
// この効果が削除され、その加速は空居力を消費せず、
// 攻撃力の35%の風元素ダメージを与える風矢4本を発射する。
func (c *char) a4() bool {
	if !c.StatusIsActive(a4Key) {
		return false
	}
	if c.StatusIsActive(a4Prevent) {
		return false
	}
	c.DeleteStatus(a4Key)
	c.AddStatus(a4Prevent, 20, true) // 20フレーム内の固有天賦4の再発動を防止（1回のダッシュでの2重発動防止）

	c.Core.Log.NewEvent("wanderer-a4 proc'd", glog.LogCharacterEvent, c.Index)

	a4Mult := 0.35

	if c.StatusIsActive("wanderer-c1-atkspd") {
		a4Mult = 0.6
	}

	a4Info := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Gales of Reverie",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagWandererA4,
		ICDGroup:   attacks.ICDGroupWandererA4,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       a4Mult,
	}

	for i := 0; i < 4; i++ {
		c.Core.QueueAttack(a4Info, combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 1),
			a4Release[i], a4Release[i]+a4Hitmark)
	}

	return true
}

// 「羽唄・風喀」が発動時に水/炎/氷/雷元素と接触した場合、
// その「風の恵み」状態にバフが付与される。
func (c *char) absorbCheckA1() {
	a1AbsorbCheckLocation := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
	a1Proc := false // for C4
	// 吸収チェックから最大2つの固有天賦1元素を取得
	for i := 0; i < 2; i++ {
		absorbCheck := c.Core.Combat.AbsorbCheck(c.Index, a1AbsorbCheckLocation, c.a1ValidBuffs...)
		if absorbCheck == attributes.NoElement {
			continue
		}
		a1Proc = true
		c.addA1Buff(absorbCheck)
		c.deleteFromValidBuffs(absorbCheck)
		c.Core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index,
			"wanderer a1 absorbed ", absorbCheck.String(),
		)
	}
	// 少なくとも1つの固有天賦1元素が吸収された場合、第4命ノ星座で固有天賦1バフを追加
	if c.Base.Cons >= 4 && a1Proc {
		chosenElement := c.a1ValidBuffs[c.Core.Rand.Intn(len(c.a1ValidBuffs))]
		c.addA1Buff(chosenElement)
		c.deleteFromValidBuffs(chosenElement)
		c.Core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index,
			"wanderer c4 applied a1 ", chosenElement.String(),
		)
	}
}

// - 水: 空居力上限が20増加。
//
// - 炎: 攻撃力が30%増加。
//
// - 氷: 会心率が20%増加。
//
// - 雷: 通常攻撃と重撃が敵に命中すると、0.8エネルギーが回復する。0.2秒に1回発動可能。
//
// 同時に最大2種類のバフを持つことができる。
func (c *char) addA1Buff(absorbCheck attributes.Element) {
	// バフ（「風の恵み」状態終了時に手動で削除が必要）
	switch absorbCheck {
	case attributes.Hydro:
		c.maxSkydwellerPoints += 20
		c.skydwellerPoints += 20

	case attributes.Pyro:
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.3
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1PyroKey, 1200),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

	case attributes.Cryo:
		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = 0.2
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1CryoKey, 1200),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

	case attributes.Electro:
		c.AddStatus(a1ElectroKey, 1200, true)
	}
}

// 通常攻撃と重撃が敵に命中すると、0.8エネルギーが回復する。0.2秒に1回発動可能。
func (c *char) makeA1ElectroCB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if !c.StatusIsActive(a1ElectroKey) {
			return
		}
		if c.StatusIsActive(a1ElectroIcdKey) {
			return
		}
		c.AddStatus(a1ElectroIcdKey, 12, true)
		c.AddEnergy("wanderer-a1-electro-energy", 0.8)
	}
}

func (c *char) deleteFromValidBuffs(ele attributes.Element) {
	var results []attributes.Element
	for _, e := range c.a1ValidBuffs {
		if e != ele {
			results = append(results, e)
		}
	}
	c.a1ValidBuffs = results
}
