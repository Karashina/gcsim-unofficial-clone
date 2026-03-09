package neuvillette

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/internal/template/sourcewaterdroplet"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var chargeFrames []int
var endLag []int

const initialLegalEvalDur = 209

var dropletLegalEvalReduction = []int{0, 57, 57 + 54, 57 + 54 + 98}

const shortChargeHitmark = 27

const chargeJudgementName = "Charged Attack: Equitable Judgment"

func init() {
	chargeFrames = frames.InitAbilSlice(87)
	chargeFrames[action.ActionCharge] = 69
	chargeFrames[action.ActionSkill] = 26
	chargeFrames[action.ActionBurst] = 27
	chargeFrames[action.ActionDash] = 25
	chargeFrames[action.ActionJump] = 26
	chargeFrames[action.ActionWalk] = 61
	chargeFrames[action.ActionSwap] = 58

	endLag = frames.InitAbilSlice(51)
	endLag[action.ActionWalk] = 36
	endLag[action.ActionCharge] = 30
	endLag[action.ActionSwap] = 27
	endLag[action.ActionBurst] = 0
	endLag[action.ActionSkill] = 0
	endLag[action.ActionDash] = 0
	endLag[action.ActionJump] = 0
}

func (c *char) legalEvalFindDroplets() int {
	droplets := c.getSourcewaterDroplets()

	// TODO: 雫のチェック前にタイムアウトした場合はカウントされない。
	indices := c.Core.Combat.Rand.Perm(len(droplets))
	orbs := 0
	for _, ind := range indices {
		g := droplets[ind]
		c.consumeDroplet(g)
		orbs += 1
		if orbs >= 3 {
			break
		}
	}
	c.Core.Combat.Log.NewEvent(fmt.Sprint("Picked up ", orbs, " droplets"), glog.LogCharacterEvent, c.Index)
	return orbs
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.chargeEarlyCancelled {
		return action.Info{}, fmt.Errorf("%v: Cannot early cancel Charged Attack: Equitable Judgement with Charged Attack", c.CharWrapper.Base.Key)
	}
	// ダッシュ/ジャンプ/歩行/スワップからのワインドアップがある。それ以外はQ/E/CA/NA -> CA フレームに含まれる
	windup := 0
	switch c.Core.Player.CurrentState() {
	case action.Idle, action.DashState, action.JumpState, action.WalkState, action.SwapState:
		windup = 14
	}

	if p["short"] != 0 {
		return c.chargeAttackShort(windup)
	}

	return c.chargeAttackJudgement(p, windup)
}

func (c *char) chargeAttackJudgement(p map[string]int, windup int) (action.Info, error) {
	c.chargeJudgeDur = 0
	c.tickAnimLength = getChargeJudgementHitmarkDelay(1)
	// 現在のフレームワークはアクションの短縮をサポートしていないため、法律評価は0に設定されるが、後で増加する可能性がある
	chargeLegalEvalLeft := 0

	c.QueueCharTask(func() {
		chargeLegalEvalLeft = initialLegalEvalDur
		orbs := c.legalEvalFindDroplets()
		chargeLegalEvalLeft -= dropletLegalEvalReduction[orbs]

		c.chargeAi = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       chargeJudgementName,
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagExtraAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Hydro,
			Durability: 25,
			FlatDmg:    chargeJudgement[c.TalentLvlAttack()] * c.MaxHP(),
		}

		c.chargeJudgeStartF = c.Core.F + chargeLegalEvalLeft
		c.chargeJudgeDur = 173

		if c.Base.Cons >= 6 {
			c.QueueCharTask(c.c6DropletCheck(c.chargeJudgeStartF), chargeLegalEvalLeft)
		}

		ticks, ok := p["ticks"]
		if !ok {
			ticks = -1
		} else {
			ticks = max(ticks, 1)
		}

		c.Core.Player.SwapCD = math.MaxInt16

		// ヒットラグ影響キューは使用不可。ロジックが正しく動作しないため
		// → ヒットラグによるアニメーション長更新の遅延を考慮できない（シムは次のアクションに移行するがティックは継続する）

		// ticks パラメータ指定時に正しいティック数のため1からカウント開始
		c.Core.Tasks.Add(c.chargeJudgementTick(c.chargeJudgeStartF, 1, ticks, false), chargeLegalEvalLeft+getChargeJudgementHitmarkDelay(1))

		// TODO: HP消費タイミングはpingの影響を受けるか？
		// 3秒間で5回消費、フレーム40, 70, 100, 130, 160
		c.QueueCharTask(c.consumeHp(c.chargeJudgeStartF), chargeLegalEvalLeft+40)
	}, windup+3)

	return action.Info{
		Frames: func(next action.Action) int {
			return windup + 3 + chargeLegalEvalLeft + c.tickAnimLength + endLag[next]
		},
		AnimationLength: 1200, // 重撃の持続時間に上限はない
		CanQueueAfter:   windup + 3 + endLag[action.ActionDash],
		State:           action.ChargeAttackState,
		OnRemoved: func(next action.AnimationState) {
			// 早期キャンセル時の正確なスワップCDを計算する必要がある
			switch next {
			case action.SkillState, action.BurstState, action.DashState, action.JumpState:
				c.Core.Player.SwapCD = max(player.SwapCDFrames-(c.Core.F-c.lastSwap), 0)
			}
		},
	}, nil
}

func (c *char) chargeAttackShort(windup int) (action.Info, error) {
	// リリースが速すぎると3つのオーブを吸収しても大重撃が発動しない
	c.QueueCharTask(func() {
		c.legalEvalFindDroplets()
		// 重撃に必要なスタミナが不足している場合、何も起こらず浮き下がる
		r := 1 + c.Core.Player.StamPercentMod(action.ActionCharge)
		if r < 0 {
			r = 0
		}
		if c.Core.Player.Stam > 50*r {
			c.Core.Player.UseStam(50*r, action.ActionCharge)
			ai := combat.AttackInfo{
				ActorIndex: c.Index,
				Abil:       "Charge Attack",
				AttackTag:  attacks.AttackTagExtra,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypePierce,
				Element:    attributes.Hydro,
				Durability: 25,
				Mult:       charge[c.TalentLvlAttack()],
			}
			ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 3, 8)
			// TODO: スナップショットのタイミングが不明
			c.Core.QueueAttack(
				ai,
				ap,
				shortChargeHitmark+windup,
				shortChargeHitmark+windup,
			)
		}
	}, windup+3)

	return action.Info{
		Frames:          func(next action.Action) int { return windup + chargeFrames[next] },
		AnimationLength: windup + chargeFrames[action.InvalidAction],
		CanQueueAfter:   windup + chargeFrames[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) judgementWave() {
	// 毎ヒット計算する。canqueueafterが最初のティック後のため、重撃中にプライマリターゲット/エンティティ位置が変更される可能性がある
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 3.5, 15)
	if c.Base.Ascension >= 1 {
		c.chargeAi.FlatDmg = chargeJudgement[c.TalentLvlAttack()] * c.MaxHP() * a1Multipliers[c.countA1()]
	}
	if c.Base.Cons >= 6 {
		c.Core.QueueAttack(c.chargeAi, ap, 0, 0, c.c6cb)
	} else {
		c.Core.QueueAttack(c.chargeAi, ap, 0, 0)
	}
}

func getChargeJudgementHitmarkDelay(tick int) int {
	// 最初のティックは開始6f後、2番目は最初から22f後、以降は25f後、最後のティックは審判の波が終了する時点。
	// TODO: 6凸の場合もこの通りか確認
	switch tick {
	case 1:
		return 6
	case 2:
		return 22
	default:
		return 25
	}
}

func (c *char) chargeJudgementTick(src, tick, maxTick int, last bool) func() {
	return func() {
		if c.chargeJudgeStartF != src {
			return
		}
		// 重撃アニメーション外 → ティックなし
		if c.Core.F > c.chargeJudgeStartF+c.chargeJudgeDur {
			return
		}

		// 最後のティック → 6凸の延長を確認
		if last {
			// 6凸が重撃を延長しなかった → 波を発動し、ティックキューを停止してendLag分のスワップCDを設定
			if c.Core.F == c.chargeJudgeStartF+c.chargeJudgeDur {
				c.judgementWave()
				// フル重撃は確実に60f以上かかるため、ここでlastFrameチェックは不要
				c.Core.Player.SwapCD = endLag[action.ActionSwap]
			} else {
				// このティックのキュー時と実行時の間に6凸が重撃を延長した
				// → 他の非最終キュータスクを実行できるようアニメーション長を延長
				c.tickAnimLength = c.tickAnimLengthC6Extend
			}
			return
		}

		// tick パラメータ指定済みで上限に到達 → 波を発動し、次のアクションチェック用の早期キャンセルフラグを有効化してティックキューを停止
		if tick == maxTick {
			c.judgementWave()
			c.chargeEarlyCancelled = true
			return
		}

		c.judgementWave()

		// 次のTick処理
		if maxTick == -1 || tick < maxTick {
			tickDelay := getChargeJudgementHitmarkDelay(tick + 1)
			// 次のティック発生までの新しいアニメーション長を計算
			nextTickAnimLength := c.Core.F - c.chargeJudgeStartF + tickDelay

			// 6凸が発動した場合に実行される非最終ティックを常にキューに追加
			c.Core.Tasks.Add(c.chargeJudgementTick(src, tick+1, maxTick, false), tickDelay)

			// 次のティックがCA持続時間終了後に発生する場合、最終ティックをキューに追加
			if nextTickAnimLength > c.chargeJudgeDur {
				// CA持続時間の終了時に最終ティックをキューに追加
				c.Core.Tasks.Add(c.chargeJudgementTick(src, tick+1, maxTick, true), c.chargeJudgeDur-c.tickAnimLength)
				// tickAnimLengthを最終的にCA持続時間全体と等しくする
				c.tickAnimLength = c.chargeJudgeDur
				// 6凸が発動した場合tickAnimLengthが不正確になるため、元の最終ティック以降もティックが通常通り継続した場合の実際のtickAnimLengthをこの変数で保持
				c.tickAnimLengthC6Extend = nextTickAnimLength
			} else {
				// 次のティックがCA持続時間内に発生 → 通常通りtickAnimLengthを更新
				c.tickAnimLength = nextTickAnimLength
			}
		}
	}
}

func (c *char) consumeHp(src int) func() {
	return func() {
		if c.chargeJudgeStartF != src {
			return
		}
		if c.Core.F > c.chargeJudgeStartF+c.chargeJudgeDur {
			return
		}
		if c.CurrentHPRatio() > 0.5 {
			hpDrain := 0.08 * c.MaxHP()

			c.Core.Player.Drain(info.DrainInfo{
				ActorIndex: c.Index,
				Abil:       "Charged Attack: Equitable Judgment",
				Amount:     hpDrain,
			})
		}
		c.QueueCharTask(c.consumeHp(src), 30)
	}
}

func (c *char) consumeDroplet(g *sourcewaterdroplet.Gadget) {
	g.Kill()
	// TODO: ping量に応じてヒーリング遅延を調整
	// ヒーリングは5fの僅かな遅延がある
	c.QueueCharTask(func() {
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "Sourcewater Droplets Healing",
			Src:     c.MaxHP() * 0.16,
			Bonus:   c.Stat(attributes.Heal),
		})
	}, 5)
}
