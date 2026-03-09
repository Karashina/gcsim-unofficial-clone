package reactable

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

func (r *Reactable) TryAddEC(a *combat.AttackEvent) bool {
	if r.core.Player.ByIndex(a.Info.ActorIndex).StatusIsActive("LC-Key") {
		return r.TryAddLC(a)
	}
	if a.Info.Durability < ZeroDur {
		return false
	}
	// 凍結がまだ残っている場合、感電反応を試みない
	// ゲームは凍結が存在する場合感電反応を積極的に拒否する
	if r.Durability[Frozen] > ZeroDur {
		return false
	}

	// 感電または水の追加は元素量を増加させるだけ
	switch a.Info.Element {
	case attributes.Hydro:
		// 既存の水または雷がなければ何もしない
		if r.Durability[Electro] < ZeroDur {
			return false
		}
		// 水元素量に追加（攻撃が既に反応していたら追加できない）
		//TODO: ここで発生すべきではない
		if !a.Reacted {
			r.attachOrRefillNormalEle(Hydro, a.Info.Durability)
		}
	case attributes.Electro:
		// 既存の水または雷がなければ何もしない
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		// 雷元素量に追加（攻撃が既に反応していたら追加できない）
		if !a.Reacted {
			r.attachOrRefillNormalEle(Electro, a.Info.Durability)
		}
	default:
		return false
	}

	a.Reacted = true
	r.core.Events.Emit(event.OnElectroCharged, r.self, a)

	// この時点で感電反応がリフレッシュされたので、反応を発動し
	// 所有権を変更する必要がある
	atk := combat.AttackInfo{
		ActorIndex:       a.Info.ActorIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.ElectroCharged),
		AttackTag:        attacks.AttackTagECDamage,
		ICDTag:           attacks.ICDTagECDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		IgnoreDefPercent: 1,
	}
	char := r.core.Player.ByIndex(a.Info.ActorIndex)
	em := char.Stat(attributes.EM)
	flatdmg, snap := calcReactionDmg(char, atk, em)
	atk.FlatDmg = 2.0 * flatdmg
	r.ecAtk = atk
	r.ecSnapshot = snap

	// 新規感電反応なら即座にティックを発動し、ティックをキューする
	// そうでなければ何もしない
	//TODO: 感電反応のリフレッシュが即座に新しいティックを発動するか確認が必要
	if r.ecTickSrc == -1 {
		r.ecTickSrc = r.core.F
		r.core.QueueAttackWithSnap(
			r.ecAtk,
			r.ecSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			10,
		)

		r.core.Tasks.Add(r.nextTick(r.core.F), 60+10)
		// 減衰ティックを購読
		r.core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			// ターゲットが最初、次にスナップショット
			n := args[0].(combat.Target)
			a := args[1].(*combat.AttackEvent)
			dmg := args[2].(float64)
			//TODO: ターゲットインデックスがない
			if n.Key() != r.self.Key() {
				return false
			}
			if a.Info.AttackTag != attacks.AttackTagECDamage {
				return false
			}
			// ICDによりこのダメージが消された場合は無視
			if dmg == 0 {
				return false
			}
			// 雷と水の両方がなくなった場合は無視
			if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
				return true
			}

			// 0.1秒後に減衰
			r.core.Tasks.Add(func() {
				r.waneEC()
			}, 6)
			return false
		}, fmt.Sprintf("ec-%v", r.self.Key()))
	}

	// ティックは前回のティックから60フレーム間隔
	// ティックダメージを受けると前回のティックがリセットされる
	return true
}

func (r *Reactable) waneEC() {
	r.Durability[Electro] -= 10
	r.Durability[Electro] = max(0, r.Durability[Electro])
	r.Durability[Hydro] -= 10
	r.Durability[Hydro] = max(0, r.Durability[Hydro])
	r.core.Log.NewEvent("ec wane",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "ec").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[Hydro]).
		Write("electro", r.Durability[Electro])

	// 感電反応が終了
	r.checkEC()
}

func (r *Reactable) checkEC() {
	if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
		r.ecTickSrc = -1
		r.core.Events.Unsubscribe(event.OnEnemyDamage, fmt.Sprintf("ec-%v", r.self.Key()))
		r.core.Log.NewEvent("ec expired",
			glog.LogElementEvent,
			-1,
		).
			Write("aura", "ec").
			Write("target", r.self.Key()).
			Write("hydro", r.Durability[Hydro]).
			Write("electro", r.Durability[Electro])
	}
}

func (r *Reactable) nextTick(src int) func() {
	return func() {
		if r.ecTickSrc != src {
			// ソースが変更されたので何もしない
			return
		}
		// 感電反応はまだアクティブであるはず。そうでなければクリーンアップが
		// 呼ばれてソースが-1に設定されているはず
		if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
			return
		}

		// 感電反応がアクティブなので、両オーラの値が0より大きいはず；ダメージを与える
		r.core.QueueAttackWithSnap(
			r.ecAtk,
			r.ecSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			0,
		)

		// 次のティックをキュー
		r.core.Tasks.Add(r.nextTick(src), 60)
	}
}
