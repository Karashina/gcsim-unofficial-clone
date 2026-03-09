package neuvillette

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/internal/template/sourcewaterdroplet"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var burstFrames []int
var burstHitmarks = [3]int{95, 95 + 40, 95 + 40 + 19}

var dropletPosOffsets = [][][]float64{{{0, 7}, {-1, 7.5}, {0.8, 6.5}}, {{-3.5, 7.5}, {-2.5, 6}}, {{3.3, 6}}}
var dropletRandomRanges = [][]float64{{0.5, 2}, {0.5, 1.2}, {0.5, 1.2}}

var defaultBurstAtkPosOffsets = [][]float64{{-3, 7.5}, {4, 6}}
var burstTickTargetXOffsets = []float64{1.5, -1.5}

func init() {
	burstFrames = frames.InitAbilSlice(135)
	burstFrames[action.ActionCharge] = 133
	burstFrames[action.ActionSkill] = 127
	burstFrames[action.ActionDash] = 127
	burstFrames[action.ActionJump] = 128
	burstFrames[action.ActionWalk] = 134
	burstFrames[action.ActionSwap] = 120
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.chargeEarlyCancelled = false
	player := c.Core.Combat.Player()

	aiInitialHit := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "O Tides, I Have Returned: Skill DMG",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    burst[c.TalentLvlBurst()] * c.MaxHP(),
	}
	aiWaterfall := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "O Tides, I Have Returned: Waterfall DMG",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    burstWaterfall[c.TalentLvlBurst()] * c.MaxHP(),
	}
	for i := 0; i < 3; i++ {
		dropletCount := 3 - i
		ai := aiInitialHit
		if i > 0 {
			ai = aiWaterfall
		}

		c.QueueCharTask(func() {
			// プレイヤー位置からオフセット付きランダムポイントで現在のティックの水の雫を生成
			for j := 0; j < dropletCount; j++ {
				sourcewaterdroplet.New(
					c.Core,
					geometry.CalcRandomPointFromCenter(
						geometry.CalcOffsetPoint(
							player.Pos(),
							geometry.Point{X: dropletPosOffsets[i][j][0], Y: dropletPosOffsets[i][j][1]},
							player.Direction(),
						),
						dropletRandomRanges[i][0],
						dropletRandomRanges[i][1],
						c.Core.Rand,
					),
					combat.GadgetTypSourcewaterDropletNeuv,
				)
			}
			c.Core.Combat.Log.NewEvent(fmt.Sprint("Burst: Spawned ", dropletCount, " droplets"), glog.LogCharacterEvent, c.Index)

			// 攻撃パターンを決定
			// 初回ティック
			ap := combat.NewCircleHitOnTarget(player, geometry.Point{Y: 1}, 8)
			// 2回目と3回目のティック
			if i > 0 {
				// 攻撃パターンの位置を決定
				// デフォルト仮定: 射程内にターゲットなし → プレイヤーからの特定オフセットにティックを生成
				apPos := geometry.CalcOffsetPoint(
					player.Pos(),
					geometry.Point{
						X: defaultBurstAtkPosOffsets[i-1][0],
						Y: defaultBurstAtkPosOffsets[i-1][1],
					},
					player.Direction(),
				)

				// ターゲットが射程内か確認
				target := c.Core.Combat.PrimaryTarget()
				if target.IsWithinArea(combat.NewCircleHitOnTarget(player, nil, 10)) {
					// ターゲットが射程内 → 位置を調整
					// 位置はターゲット位置+オフセットからのランダム範囲内の点
					// TODO: 現在ターゲットは常にデフォルト方向を向いているためオフセットが不正確
					apPos = geometry.CalcRandomPointFromCenter(
						geometry.CalcOffsetPoint(
							target.Pos(),
							geometry.Point{X: burstTickTargetXOffsets[i-1]},
							target.Direction(),
						),
						0,
						1.5,
						c.Core.Rand,
					)
				}
				// 位置決定後にティック用の攻撃パターンを生成
				ap = combat.NewCircleHitOnTarget(apPos, nil, 5)
			}

			c.Core.QueueAttack(
				ai,
				ap,
				0,
				0,
			)
		}, burstHitmarks[i])
	}

	c.SetCD(action.ActionBurst, 18*60)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
