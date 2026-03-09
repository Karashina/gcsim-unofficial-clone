package chiori

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillFrames            [][]int
	skillHitmarks          = []int{21, 37}
	skillCDStarts          = []int{19, 34} // 人形の出現と同じ
	skillA1WindowStarts    = []int{26, 42}
	skillA1WindowDurations = []int{78, 77}
)

const (
	skillDollDuration               = int(17.5 * 60)
	skillDollStartDelay             = int(0.6 * 60)
	skillDollAttackInterval         = int(3.6 * 60)
	skillDollConstructCheckInterval = int(0.5 * 60)
	skillDollAttackDelay            = 5 // 0.08秒であるべき
	skillDollXOffset                = 1.2
	skillDollYOffset                = -0.3
	skillDollAoE                    = 1.2

	skillRockDollStartDelay = int(1.2 * 60)

	skillCD = 16 * 60

	particleICDKey = "chiori-particle-icd"
)

func init() {
	skillFrames = make([][]int, 2)

	// 元素スキル（単押し）
	skillFrames[0] = frames.InitAbilSlice(51) // Tap E -> Walk
	skillFrames[0][action.ActionAttack] = 42
	skillFrames[0][action.ActionSkill] = 30
	skillFrames[0][action.ActionBurst] = 43
	skillFrames[0][action.ActionDash] = 42
	skillFrames[0][action.ActionJump] = 42
	skillFrames[0][action.ActionSwap] = 49

	// 元素スキル（長押し）
	skillFrames[1] = frames.InitAbilSlice(88) // Hold E -> N1/Q/D/J
	skillFrames[1][action.ActionLowPlunge] = 52
	skillFrames[1][action.ActionSkill] = 44
	skillFrames[1][action.ActionWalk] = 86
	skillFrames[1][action.ActionSwap] = 87
}

// 絹の歩みで素早く前方にダッシュ。ダッシュ終了時、千織は
// オートマトン人形「Tamoto」を召喚し、刃を振り上げて
// 攻撃力と防御力に基づく岩元素範囲ダメージを与える。
// 長押しで異なる動作をする。
//
// 長押しで照準モードに入り、ダッシュ方向を調整。
//
// Tamoto
// - 一定間隔で付近の敵を斜り、千織の攻撃力と防御力に基づく
// 岩元素範囲ダメージを与える。
// - アクティブ中、付近に岩元素設置物が生成されると、追加のTamotoが
// 千織の横に召喚される。この方法では1体のみ追加可能で、
// 持続時間は独立してカウントされる。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold < 0 {
		hold = 0
	}
	if hold > 1 {
		hold = 1
	}

	// 2回目の押下ならスワップして固有天賦1を発動
	if c.StatusIsActive(a1WindowKey) {
		return c.skillRecast()
	}

	// 現時点では分割は不要だが、将来の変更に対応可能
	// 長押しが特別な動作をする場合に備えて
	c.handleSkill(hold)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[hold]),
		AnimationLength: skillFrames[hold][action.InvalidAction],
		CanQueueAfter:   skillFrames[hold][action.ActionSkill],
		State:           action.SkillState,
	}, nil
}

func (c *char) skillRecast() (action.Info, error) {
	c.a1Tapestry()
	// 次のキャラを検索
	next := c.Index + 1
	if next >= len(c.Core.Player.Chars()) {
		next = 0
	}
	k := c.Core.Player.ByIndex(next).Base.Key
	c.Core.Tasks.Add(func() {
		c.Core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, "forcing swap to ", k.String())
		c.Core.Player.Exec(action.ActionSwap, k, nil)
	}, 1)
	// TODO: 強制スワップのためこの持続時間は実際には重要でない、スワップが実行されるまでの1fをカバーするだけでよい
	return action.Info{
		Frames:          func(action.Action) int { return c.Core.Player.Delays.Swap },
		AnimationLength: c.Core.Player.Delays.Swap,
		CanQueueAfter:   c.Core.Player.Delays.Swap,
		State:           action.SkillState,
	}, nil
}

func (c *char) handleSkill(hold int) {
	// 上昇攻撃を処理
	c.Core.Tasks.Add(func() {
		ai := combat.AttackInfo{
			Abil:       "Fluttering Hasode (Upward Sweep)",
			ActorIndex: c.Index,
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagChioriSkill,
			ICDGroup:   attacks.ICDGroupChioriSkill,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   50,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       thrustAtkScaling[c.TalentLvlSkill()],
		}

		snap := c.Snapshot(&ai)
		ai.FlatDmg = snap.Stats.TotalDEF()
		ai.FlatDmg *= thrustDefScaling[c.TalentLvlSkill()]

		c.Core.QueueAttackWithSnap(ai, snap, combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 5.5, 3.5), 0)
	}, skillHitmarks[hold])

	// CDをトリガーし、固有天賦1ウィンドウをアクティブ化
	c.SetCDWithDelay(action.ActionSkill, skillCD, skillCDStarts[hold])
	c.activateA1Window(skillA1WindowStarts[hold], skillA1WindowDurations[hold])

	// 人形の生成を処理
	c.Core.Tasks.Add(func() {
		// 1体目の人形を生成
		c.createDoll()

		// 1凸でない場合、rock dollを生成する設置物チェッカーを作成
		if !c.c1Active {
			c.createDollConstructChecker()
			return
		}

		// 1凸からrock dollを生成し、固有天賦4を発動
		c.Core.Log.NewEvent("c1 spawning rock doll", glog.LogCharacterEvent, c.Index)
		c.createRockDoll()
		c.applyA4Buff()
	}, skillCDStarts[hold])
}

func (c *char) createDoll() {
	// 既存の人形を破棄
	c.kill(c.skillDoll)

	// 人形の位置を決定
	player := c.Core.Combat.Player()
	dollPos := geometry.CalcOffsetPoint(
		player.Pos(),
		geometry.Point{X: skillDollXOffset, Y: skillDollYOffset},
		player.Direction(),
	)

	c.Core.Log.NewEvent("spawning doll", glog.LogCharacterEvent, c.Index)

	// 新しい人形を生成
	doll := newTicker(c.Core, skillDollDuration, nil)
	doll.cb = c.skillDollAttack(c.Core.F, "Fluttering Hasode (Tamato)", dollPos)
	doll.interval = skillDollAttackInterval
	c.Core.Tasks.Add(doll.tick, skillDollStartDelay)
	c.skillDoll = doll
}

func (c *char) createDollConstructChecker() {
	// 既存の設置物チェッカーを破棄
	c.kill(c.constructChecker)

	// rock dollを生成するための関連設置物チェッカーを生成
	cc := newTicker(c.Core, skillDollDuration, nil)
	cc.cb = c.skillDollConstructCheck
	cc.interval = skillDollConstructCheckInterval
	cc.tick() // t = 0sでティック開始
	c.constructChecker = cc
}

func (c *char) skillDollAttack(src int, abil string, pos geometry.Point) func() {
	return func() {
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				Abil:       abil,
				ActorIndex: c.Index,
				AttackTag:  attacks.AttackTagElementalArt,
				ICDTag:     attacks.ICDTagChioriSkill,
				ICDGroup:   attacks.ICDGroupChioriSkill,
				StrikeType: attacks.StrikeTypeBlunt,
				PoiseDMG:   0,
				Element:    attributes.Geo,
				Durability: 25,
				Mult:       turretAtkScaling[c.TalentLvlSkill()],
			}

			snap := c.Snapshot(&ai)
			ai.FlatDmg = snap.Stats.TotalDEF()
			ai.FlatDmg *= turretDefScaling[c.TalentLvlSkill()]

			// プレイヤーに攻撃ターゲットがある場合は常にこの敵を選択する
			// 検索AoE内にあることを確認するだけでよい
			t := c.Core.Combat.PrimaryTarget()
			if !t.IsWithinArea(combat.NewCircleHitOnTarget(pos, nil, c.skillSearchAoE)) {
				return
			}

			c.Core.Log.NewEvent("doll attacking", glog.LogCharacterEvent, c.Index).Write("src", src)

			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(t, nil, skillDollAoE), 0, c.particleCB)
		}, skillDollAttackDelay)
	}
}

func (c *char) skillDollConstructCheck() {
	// 設置物が存在する場合、rock dollを生成できない
	if c.rockDoll != nil && c.rockDoll.alive {
		return
	}
	// TODO: 技術的にはスキル人形から30m半径内の設置物をチェックすべき
	// 人形の位置は既に攻撃関数に渡されているので再利用可能
	// それほど重要ではなく、設置物ハンドラが直接公開していないため未実装
	if c.Core.Constructs.Count() == 0 {
		return
	}

	c.Core.Log.NewEvent("construct spawning rock doll", glog.LogCharacterEvent, c.Index)
	c.createRockDoll()

	// このチェックが再度発生しないようにする
	c.kill(c.constructChecker)
}

func (c *char) createRockDoll() {
	// 既存を破棄
	c.kill(c.rockDoll)

	// 人形の位置を決定
	player := c.Core.Combat.Player()
	dollPos := geometry.CalcOffsetPoint(
		player.Pos(),
		geometry.Point{X: skillDollXOffset, Y: skillDollYOffset},
		player.Direction(),
	)

	// 新しいrock dollを生成
	rd := newTicker(c.Core, skillDollDuration, nil)
	rd.cb = c.skillDollAttack(c.Core.F, "Fluttering Hasode (Tamato - Construct)", dollPos)
	rd.interval = skillDollAttackInterval
	c.Core.Tasks.Add(rd.tick, skillRockDollStartDelay)
	c.rockDoll = rd
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 3*60, true)

	count := 1.0
	if c.Core.Rand.Float64() < 0.2 {
		count = 2.0
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
}
