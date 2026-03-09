package enemy

import (
	"log"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

var particleIDToElement = []attributes.Element{
	attributes.NoElement,
	attributes.Pyro,
	attributes.Hydro,
	attributes.Dendro,
	attributes.Electro,
	attributes.Anemo,
	attributes.Cryo,
	attributes.Geo,
}

func (e *Enemy) HandleAttack(atk *combat.AttackEvent) float64 {
	// この時点で攻撃が命中する
	e.Core.Combat.Events.Emit(event.OnEnemyHit, e, atk)

	var amp string
	var cata string
	var dmg float64
	var crit bool

	evt := e.Core.Combat.Log.NewEvent(atk.Info.Abil, glog.LogDamageEvent, atk.Info.ActorIndex).
		Write("target", e.Key()).
		Write("attack-tag", atk.Info.AttackTag).
		Write("ele", atk.Info.Element.String()).
		Write("damage", &dmg).
		Write("crit", &crit).
		Write("amp", &amp).
		Write("cata", &cata).
		Write("abil", atk.Info.Abil).
		Write("source_frame", atk.SourceFrame)
	evt.WriteBuildMsg(atk.Snapshot.Logs...)

	if !atk.Info.SourceIsSim {
		if atk.Info.ActorIndex < 0 {
			log.Println(atk)
		}
		preDmgModDebug := e.Core.Combat.Team.CombatByIndex(atk.Info.ActorIndex).ApplyAttackMods(atk, e)
		evt.Write("pre_damage_mods", preDmgModDebug)
	}

	dmg, crit = e.attack(atk, evt)

	// ダメージイベントをフレーム末まで遅延させる
	e.Core.Combat.Tasks.Add(func() {
		// ダメージを適用
		actualDmg := e.applyDamage(atk, dmg)
		e.Core.Combat.TotalDamage += actualDmg
		e.Core.Combat.Events.Emit(event.OnEnemyDamage, e, atk, actualDmg, crit)
		// コールバック
		cb := combat.AttackCB{
			Target:      e,
			AttackEvent: atk,
			Damage:      actualDmg,
			IsCrit:      crit,
		}
		for _, f := range atk.Callbacks {
			f(cb)
		}
	}, 0)

	// これがGolangで動作するのは、stringが内部的にスライスであり、&ampはスライス情報を
	// 指しているため。ampの内部文字列が変更（再割り当て）されてもポインタは
	// スライス「ヘッダ」を指しているので変わらない
	if atk.Info.Amped {
		amp = string(atk.Info.AmpType)
	}
	if atk.Info.Catalyzed {
		cata = string(atk.Info.CatalyzedType)
	}
	return dmg
}

func (e *Enemy) attack(atk *combat.AttackEvent, evt glog.Event) (float64, bool) {
	// 攻撃着弾前にターゲットが凍結されている場合、インパルスを0に設定
	// 解凍攻撃が実際のインパルスをトリガーさせる
	if e.Durability[reactable.Frozen] > reactable.ZeroDur {
		atk.Info.NoImpulse = true
	}

	// まず耐久値ダメージと粉砕をチェック
	e.PoiseDMGCheck(atk)
	e.ShatterCheck(atk)

	checkBurningICD := func() {
		// 燃焼ダメージのグローバルICDの特殊処理
		if atk.Info.ICDTag != attacks.ICDTagBurningDamage {
			return
		}
		// 他のキャラクターにも燃焼ダメージのICDをチェック
		for i := 0; i < len(e.Core.Player.Chars()); i++ {
			if i == atk.Info.ActorIndex {
				continue
			}
			// 他キャラがまだ燃焼ダメージICD中なら元素耐久をゼロにする
			atk.Info.Durability *= reactions.Durability(e.WillApplyEle(atk.Info.ICDTag, atk.Info.ICDGroup, i))
		}
	}
	// タグをチェック
	if atk.Info.Durability > 0 {
		// まずICDをチェック
		atk.Info.Durability *= reactions.Durability(e.WillApplyEle(atk.Info.ICDTag, atk.Info.ICDGroup, atk.Info.ActorIndex))
		checkBurningICD()
		if atk.Info.Durability > 0 && atk.Info.Element != attributes.Physical {
			existing := e.Reactable.ActiveAuraString()
			applied := atk.Info.Durability
			e.React(atk)
			if e.Core.Flags.LogDebug && atk.Reacted {
				e.Core.Log.NewEvent(
					"application",
					glog.LogElementEvent,
					atk.Info.ActorIndex,
				).
					Write("attack_tag", atk.Info.AttackTag).
					Write("applied_ele", atk.Info.Element.String()).
					Write("dur", applied).
					Write("abil", atk.Info.Abil).
					Write("target", e.Key()).
					Write("existing", existing).
					Write("after", e.Reactable.ActiveAuraString())
			}
		}
	}

	damage, isCrit := e.calc(atk, evt)

	// ヒットラグをチェック
	if e.Core.Combat.EnableHitlag {
		willapply := true
		if atk.Info.HitlagOnHeadshotOnly {
			willapply = atk.Info.HitWeakPoint
		}
		dur := atk.Info.HitlagHaltFrames
		if e.Core.Flags.DefHalt && atk.Info.CanBeDefenseHalted {
			dur += 3.6
		}
		dur = math.Ceil(dur)
		if willapply && dur > 0 {
			// 敵にヒットラグを適用
			e.ApplyHitlag(atk.Info.HitlagFactor, dur)
			// リアクタブルにもヒットラグを適用
			// e.Reactable.ApplyHitlag(atk.Info.HitlagFactor, dur)
		}
	}

	// 粒子ドロップをチェック
	count, element := e.tryHPDropParticle()
	if count > 0 {
		e.Core.Log.NewEvent("particle hp threshold triggered", glog.LogEnemyEvent, atk.Info.ActorIndex)
		e.Core.Tasks.Add(
			func() {
				e.Core.Player.DistributeParticle(character.Particle{
					Source: "hp_drop",
					Num:    count,
					Ele:    element,
				})
			},
			100, // TODO: グローバルディレイの影響を受けるべきか？
		)
	}

	return damage, isCrit
}

func (e *Enemy) tryHPDropParticle() (float64, attributes.Element) {
	// 敵プロファイルから粒子ドロップを使用
	if e.prof.ParticleDrops != nil {
		if e.particleDropIndex >= len(e.prof.ParticleDrops) {
			return 0, attributes.NoElement
		}
		info := e.prof.ParticleDrops[e.particleDropIndex]
		if (e.HP() / e.MaxHP()) > info.HpPercent {
			return 0, attributes.NoElement
		}
		e.particleDropIndex++
		// 22010017: 岩元素粒子1個
		diff := info.DropId - 22010000
		if diff < 0 || diff >= 100 {
			return 0, attributes.NoElement
		}
		count := (info.DropId / 10) % 10 // 2nd digit is particle count
		if count <= 0 {
			return 0, attributes.NoElement
		}
		element := particleIDToElement[info.DropId%10] // 1st digit is particle type
		return float64(count), element
	}

	// gcslから粒子ドロップを使用
	if e.prof.ParticleDropThreshold <= 0 {
		return 0, attributes.NoElement
	}
	next := int(e.damageTaken / e.prof.ParticleDropThreshold)
	if next <= e.lastParticleDrop {
		return 0, attributes.NoElement
	}
	// カウントもチェック
	count := next - e.lastParticleDrop
	e.lastParticleDrop = next
	return e.prof.ParticleDropCount * float64(count), e.prof.ParticleElement
}

func (e *Enemy) applyDamage(atk *combat.AttackEvent, damage float64) float64 {
	// ダメージを記録
	// HPが負にならないようにする（同一フレームで複数回呼ばれる可能性があるため）
	actualDmg := min(damage, e.hp) // ダメージが残りHPを超えないようにする
	e.hp -= actualDmg
	e.damageTaken += actualDmg //TODO: 実際にこれが必要か？

	// ターゲットが死亡したかチェック
	if e.Core.Flags.DamageMode && e.hp <= 0 {
		e.Kill()
		e.Core.Events.Emit(event.OnTargetDied, e, atk)
		return actualDmg
	}

	// オーラを適用
	if atk.Info.Durability > 0 && !atk.Reacted && atk.Info.Element != attributes.Physical {
		// まずICDをチェック
		existing := e.Reactable.ActiveAuraString()
		applied := atk.Info.Durability
		e.AttachOrRefill(atk)
		if e.Core.Flags.LogDebug {
			e.Core.Log.NewEvent(
				"application",
				glog.LogElementEvent,
				atk.Info.ActorIndex,
			).
				Write("attack_tag", atk.Info.AttackTag).
				Write("applied_ele", atk.Info.Element.String()).
				Write("dur", applied).
				Write("abil", atk.Info.Abil).
				Write("target", e.Key()).
				Write("existing", existing).
				Write("after", e.Reactable.ActiveAuraString())
		}
	}
	// 敵のHPを考慮せずダメージを返す:
	// - ダメージモードでターゲットが未死亡の場合（そうでなければ死亡のif文に入っている）
	// - 持続時間モード（トドメの概念なし）
	return damage
}
