package venti

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1凸（オリジナル）：狙い撃ち1矢につき追加で2本の矢を発射し、それぞれオリジナルの矢の33%のダメージを与える。
// 1凸（Hexerei追加、魔女の試練必要）：
//
//	ストームウィンドアローも追尾型の分裂矢を2本発射し、それぞれオリジナルの矢の20%のダメージを与える。
//	この効果は0.25秒（15フレーム）に1回発動可能。
func (c *char) c1(ai combat.AttackInfo, hitmark, travel int) {
	ai.Abil += " (C1)"
	ai.Mult /= 3.0
	for i := 0; i < 2; i++ {
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				geometry.Point{Y: -0.5},
				0.1,
				1,
			),
			hitmark,
			hitmark+travel,
		)
	}
}

// makeC1StormwindSplitCBはストームウィンドアローのダメージの20%で
// 命中時に追加の追尾矢を2本発射するAttackCBFuncを返す。
// 必要条件：1凸、Hexereiモード、元素爆発の眼がアクティブ。ICD：0.25秒（15フレーム）。
func (c *char) makeC1StormwindSplitCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if c.Base.Cons < 1 {
			return
		}
		if !c.isHexerei || !c.hasHexBonus {
			return
		}
		if c.Core.F-c.lastStormwindSplit < 15 { // 0.25秒ICD
			return
		}
		c.lastStormwindSplit = c.Core.F
		// 実際のヒットのAttackEventから分裂倍率を導出
		splitMult := a.AttackEvent.Info.Mult * 0.20
		splitAI := a.AttackEvent.Info
		splitAI.Abil += " (C1 Split)"
		splitAI.Mult = splitMult
		for i := 0; i < 2; i++ {
			c.Core.QueueAttack(
				splitAI,
				combat.NewBoxHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					geometry.Point{Y: -0.5},
					0.1,
					1,
				),
				0,
				1,
			)
		}
	}
}

// 2凸（オリジナル）：「高天の歌」が敵の風元素耐性と物理耐性を12%低下させる（10秒間）。
// 2凸（Hexerei追加）：単押し「高天の歌」がオリジナルの300%のダメージを与える（Hexereiのみ）。
//
//	300%倍率はHexereiがアクティブな時にskill.goで適用される。
func (c *char) c2(a combat.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("venti-c2-anemo", 600),
		Ele:   attributes.Anemo,
		Value: -0.12,
	})
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("venti-c2-phys", 600),
		Ele:   attributes.Physical,
		Value: -0.12,
	})
}

// 4凸（オリジナル）：Ventiが元素オーブまたは粒子を取得すると、10秒間風元素ダメージ+25%を得る。
// 4凸（Hexerei追加）：Ventiが「高天の歌」または「風神の詩」を使用後、Ventiと他の
//
//	アクティブパーティメンバーが10秒間風元素ダメージ+25%を得る（Hexereiのみ、venti.go Initで初期化）。
func (c *char) c4Old() {
	c4bonus := make([]float64, attributes.EndStatType)
	c4bonus[attributes.AnemoP] = 0.25
	c.Core.Events.Subscribe(event.OnParticleReceived, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("venti-c4-old", 600),
			AffectedStat: attributes.AnemoP,
			Amount: func() ([]float64, bool) {
				return c4bonus, true
			},
		})
		return false
	}, "venti-c4-old")
}

func (c *char) c4New() {
	if !c.isHexerei {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.AnemoP] = 0.25
		ch.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("venti-c4-hex", 10*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// 6凸：（魔女の試練を完了すると解放）
// 「風神の詩」に命中した敵の風元素耐性が20%低下する。
// 元素変化が発生した場合、その元素の耐性も同様に20%低下する。
// さらに、これらの敵に対するVentiの会心ダメージが100%増加する。
func (c *char) c6(ele attributes.Element) func(a combat.AttackCB) {
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("venti-c6-"+ele.String(), 600),
			Ele:   ele,
			Value: -0.20,
		})
	}
}

// c6AttackModInitはVentiに永続AttackModを追加し、
// c6で風元素耐性デバフが適用された敵に対して+100%会心ダメージを与える（Hexereiのみ）。
func (c *char) c6AttackModInit() {
	if !c.isHexerei {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("venti-c6-cd", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			e, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			// ターゲットが元素爆発に命中したか確認（風元素耐性デバフがあるか）
			if !e.StatusIsActive("venti-c6-anemo") {
				return nil, false
			}
			for i := range m {
				m[i] = 0
			}
			m[attributes.CD] = 1.0
			return m, true
		},
	})
}

// hexAttackEnabledはHexerei通常攻撃パッシブがアクティブになるべき時にtrueを返す。
// 必要条件：Hexereiモード、2人以上のHexereiパーティメンバー、元素爆発の眼がアクティブ。
// この効果は魔女の試練（Hexereiフラグ）を完了すると解放され、命ノ星座は不要。
func (c *char) hexAttackEnabled() bool {
	return c.isHexerei && c.hasHexBonus && c.Core.F < c.burstEnd
}

// makeHexNormalCBは元素爆発の眼の持続時間を延長し、
// 通常攻撃命中時に元素爆発 CDを短縮するAttackCBFuncを返す（Hexereiパッシブ）。
func (c *char) makeHexNormalCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if c.normalHexCount >= 2 {
			return
		}
		if c.Core.F-c.lastHexTrigger < 6 { // 0.1秒ICD = 6フレーム
			return
		}
		c.lastHexTrigger = c.Core.F
		c.normalHexCount++
		// 元素爆発の眼の持続時間を1秒（60フレーム）延長
		c.burstEnd += 60
		// 元素爆発CDを0.5秒（30フレーム）短縮
		c.ReduceActionCooldown(action.ActionBurst, 30)
	}
}
