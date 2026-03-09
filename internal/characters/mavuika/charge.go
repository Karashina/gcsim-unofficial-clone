package mavuika

import (
	"errors"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var chargeFrames []int
var bikeChargeFrames []int
var bikeChargeFinalFrames []int
var bikeHittableEntityList []HittableEntity

// 重撃フィニッシュアニメーション前の最小重撃時間は50f
var bikeChargeAttackMinimumDuration = 50
var bikeChargeAttackStartupHitmark = 35

// 重撃フィニッシュアニメーション前の最大重撃時間は375f
var bikeChargeAttackMaximumDuration = 375
var bikeChargeFinalHitmark = 45

// TODO: 重撃の35-46フレームをより正確に再現する
// var bikeSpinInitialFrames = 11
// var bikeSpinInitialAngularVelocity = float64(-180 / 11)
// スピン速度は現在の角度に応じて変化
var bikeSpinQuadrantAngularVelocity = []float64{-90 / 9, -90 / 7, -90 / 15, -90 / 14} // 象限 4, 3, 2, 1
var bikeSpinQuadrantFrames = []int{9, 7, 15, 14} // 象限 4, 3, 2, 1

const chargeHitmark = 40
const bikeChargeAttackICD = 42         // 重撃ヒット間の最小時間
const bikeChargeAttackSpinFrames = 45  // 1回転約~45f
const bikeChargeAttackHitboxRadius = 3 // プレースホルダー
const bikeChargeAttackSpinOffset = 4.0 // マヴイカの原点からヒットボックス中心までの推定距離

func init() {
	chargeFrames = frames.InitAbilSlice(48)
	chargeFrames[action.ActionBurst] = 50
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = 50
	chargeFrames[action.ActionWalk] = 60

	// これらの静的カウントはほとんど使用されない。ゼロ値は動的ヒットマークでキャンセルされる。未記載のアクションは重撃フィニッシュにキューされる
	bikeChargeFrames = frames.InitAbilSlice(bikeChargeAttackMinimumDuration + bikeChargeFinalHitmark)
	bikeChargeFrames[action.ActionCharge] = 0
	bikeChargeFrames[action.ActionBurst] = 0
	bikeChargeFrames[action.ActionSkill] = 0
	bikeChargeFrames[action.ActionDash] = 0
	bikeChargeFrames[action.ActionJump] = 0
	bikeChargeFrames[action.ActionSwap] = 0

	bikeChargeFinalFrames = frames.InitAbilSlice(74) // CAF -> NA
	bikeChargeFinalFrames[action.ActionWalk] = 73
	bikeChargeFinalFrames[action.ActionBurst] = bikeChargeFinalHitmark
	bikeChargeFinalFrames[action.ActionDash] = bikeChargeFinalHitmark
	bikeChargeFinalFrames[action.ActionJump] = bikeChargeFinalHitmark
	bikeChargeFinalFrames[action.ActionSwap] = bikeChargeFinalHitmark
	bikeChargeFinalFrames[action.ActionSkill] = bikeChargeFinalHitmark
}

// 重撃状態の構造体
type ChargeState struct {
	StartFrame      int
	cAtkFrames      int
	skippedWindupF  int
	LastHit         map[targets.TargetKey]int
	FacingDirection float64
	srcFrame        int
}

type HittableEntity struct {
	Entity     combat.Target
	isOneTick  bool   // 1回のmaxHitCountでエンティティが破壊されるか？
	CollFrames [2]int // 衝突が発生する重撃スピンのフレーム
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.armamentState == bike && c.nightsoulState.HasBlessing() {
		return c.BikeCharge(p)
	}
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           120.0,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagHaltFrames:   0.15 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: 0.3},
			3.3,
		),
		chargeHitmark,
		chargeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

// 重撃を開始し、持続時間計算のためのループハンドラーに進む
func (c *char) BikeCharge(p map[string]int) (action.Info, error) {
	// 重撃調整用パラメータ
	durationCA := p["hold"]
	final := p["final"]
	bufferedFrames, ok := p["buffered"]
	if ok {
		bufferedFrames = min(bufferedFrames, 15) // 重撃入力がバッファされるフレーム数、最大15f
	} else {
		bufferedFrames = 15 // デフォルトで最大バッファフレームを仮定
	}

	bikeHittableEntities, hitboxError := c.BuildBikeChargeAttackHittableTargetList()

	if hitboxError != nil {
		return action.Info{}, hitboxError
	}

	// 継続重撃か新規かを確認
	skippedWindupFrames := 0
	if c.Core.Player.CurrentState() != action.ChargeAttackState || c.caState.StartFrame == 0 {
		c.caState = ChargeState{}
		c.caState.StartFrame = c.Core.F
		c.caState.LastHit = make(map[targets.TargetKey]int)
		for _, t := range bikeHittableEntities {
			targetIndex := t.Entity.Key()
			c.caState.LastHit[targetIndex] = 0
		}
		c.bikeChargeAttackHook()
		skippedWindupFrames = c.GetSkippedWindupFrames(bufferedFrames)
		c.caState.skippedWindupF = skippedWindupFrames // 重撃フックで重撃フレームを同期するために使用
	}

	c.caState.srcFrame = c.Core.F
	src := c.caState.srcFrame
	nightSoulDuration := c.GetRemainingNightSoulDuration()
	isForceFinalHit := false // 重撃持続時間超過時に重撃フィニッシュを強制

	if final == 1 {
		return c.BikeChargeAttackFinal(0, skippedWindupFrames)
	}

	// 部分的な重撃ホールドでの開始を許可しない
	if durationCA > 0 && c.caState.cAtkFrames > 0 {
		// 持続時間を1スピン、残りナイトソウル、最大重撃時間の最小値に制限
		durationCA = min(durationCA, bikeChargeAttackSpinFrames, nightSoulDuration, bikeChargeAttackMaximumDuration-c.caState.cAtkFrames)
		// 重撃ホールドロジック
		c.HoldBikeChargeAttack(durationCA, skippedWindupFrames, bikeHittableEntities)
	} else {
		hasValidTarget, ai, err := c.HasValidTargetCheck(bikeHittableEntities)
		if !hasValidTarget {
			return ai, err
		}
		durationCA = c.CountBikeChargeAttack(1, skippedWindupFrames, bikeHittableEntities, nightSoulDuration)
	}

	// 既存の重撃フレームを加算
	c.caState.cAtkFrames += durationCA
	durationCA -= skippedWindupFrames

	if durationCA >= nightSoulDuration || c.caState.cAtkFrames >= bikeChargeAttackMaximumDuration {
		isForceFinalHit = true
	}

	if isForceFinalHit {
		return c.BikeChargeAttackFinal(durationCA, skippedWindupFrames)
	}

	// 無効アクション用の重撃フィニッシュキューを開始
	// バイクの角度が重撃フィニッシュに遅延がある位置か確認、15fウィンドウ（重撃フィニッシュキュー用）
	currentBikeSpinFrame := c.caState.cAtkFrames % bikeChargeAttackSpinFrames
	newMinSpinDuration := GetCAFDelay(currentBikeSpinFrame)

	c.QueueCharTask(func() {
		if c.caState.srcFrame != src {
			return
		}
		c.BikeChargeAttackFinal(0, 0)
	}, durationCA+1)

	return action.Info{
		Frames: func(next action.Action) int {
			f := bikeChargeFrames[next]

			if f == 0 {
				f = durationCA
			} else {
				f = durationCA + newMinSpinDuration + bikeChargeFinalFrames[next]
			}
			return f
		},
		AnimationLength: durationCA + newMinSpinDuration + bikeChargeFinalFrames[action.InvalidAction],
		CanQueueAfter:   durationCA,
		State:           action.ChargeAttackState,
		OnRemoved: func(next action.AnimationState) {
			if next != action.ChargeAttackState {
				c.caState = ChargeState{}
				c.bikeChargeAttackUnhook()
			}
		},
	}, nil
}

// 指定された重撃長に対して、ヒット可能リスト内の各ターゲットへのヒットを計算
func (c *char) HoldBikeChargeAttack(cAtkFrames, skippedWindupFrames int, hittableEntities []HittableEntity) {
	for i := 0; i < len(hittableEntities); i++ {
		t := hittableEntities[i]
		enemyID := t.Entity.Key()
		lastHitFrame := c.caState.LastHit[enemyID]
		newLastHitFrame := -1

		if t.isOneTick && lastHitFrame > 0 {
			continue
		}

		// 重撃の最初の11fはやや不正確。本来はもっと左から開始し、より速くスイープすべき
		hitFrames := c.CalculateValidCollisionFrames(cAtkFrames, t.CollFrames, lastHitFrame)

		if len(hitFrames) > 0 {
			for _, f := range hitFrames {
				c.Core.Tasks.Add(func() {
					ai := c.GetBikeChargeAttackAttackInfo()
					c.Core.QueueAttack(ai, combat.NewSingleTargetHit(t.Entity.Key()), 0, 0)
				}, f-skippedWindupFrames)
				newLastHitFrame = f
			}
		}
		if newLastHitFrame >= 0 {
			c.caState.LastHit[enemyID] += newLastHitFrame + (c.caState.cAtkFrames - lastHitFrame)
		}
	}
}

// 指定されたmaxHitCountに対して、ターゲットへのヒットタイミングを計算し重撃持続時間を返す
func (c *char) CountBikeChargeAttack(maxHitCount, skippedWindupFrames int, hittableEntities []HittableEntity, nsDur int) int {
	// ナイトソウル残り時間（スキップされたワインドアップを考慮）と最大重撃時間の間の残り重撃時間を返す
	dur := min(nsDur+skippedWindupFrames, bikeChargeAttackMaximumDuration-c.caState.cAtkFrames)
	hitCounter := 0

	for i := 0; i < len(hittableEntities); i++ {
		t := hittableEntities[i]
		if t.Entity != c.Core.Combat.PrimaryTarget() {
			continue
		}

		enemyID := t.Entity.Key()
		lastHitFrame := c.caState.LastHit[enemyID]

		if t.isOneTick && lastHitFrame > 0 {
			continue
		}

		// 重撃の最初の11fはやや不正確。本来はもっと左から開始し、より速くスイープすべき
		hitFrames := c.CalculateValidCollisionFrames(dur, t.CollFrames, lastHitFrame)

		if len(hitFrames) > 0 {
			for _, f := range hitFrames {
				hitCounter++
				if hitCounter >= maxHitCount {
					dur = f
					break
				}
			}
		}
		if hitCounter >= maxHitCount {
			break
		}
	}

	for i := 0; i < len(hittableEntities); i++ {
		t := hittableEntities[i]
		enemyID := t.Entity.Key()
		lastHitFrame := c.caState.LastHit[enemyID]
		newLastHitFrame := -1

		hitFrames := c.CalculateValidCollisionFrames(dur, t.CollFrames, lastHitFrame)

		if len(hitFrames) > 0 {
			for _, f := range hitFrames {
				c.Core.Tasks.Add(func() {
					ai := c.GetBikeChargeAttackAttackInfo()
					c.Core.QueueAttack(ai, combat.NewSingleTargetHit(t.Entity.Key()), 0, 0)
				}, f-skippedWindupFrames)
				newLastHitFrame = f
			}
		}
		// 重撃がヒット間に開始された場合に使用（通常2番目以降のターゲット向け）
		if newLastHitFrame >= 0 {
			c.caState.LastHit[enemyID] += newLastHitFrame + (c.caState.cAtkFrames - lastHitFrame)
		}
	}
	return dur
}

// 重撃フィニッシュは最大重撃持続時間到達、ナイトソウル終了、または重撃解除時に発生
func (c *char) BikeChargeAttackFinal(caFrames, skippedWindupFrames int) (action.Info, error) {
	bikeChargeAttackElapsedTime := c.caState.cAtkFrames + caFrames
	var newMinSpinDuration int
	if bikeChargeAttackElapsedTime > 0 {
		// バイクの角度が重撃フィニッシュに遅延がある位置か確認、20fウィンドウ
		currentBikeSpinFrame := bikeChargeAttackElapsedTime % bikeChargeAttackSpinFrames
		newMinSpinDuration = GetCAFDelay(currentBikeSpinFrame)
	} else { // 新規重撃の場合、最早の重撃フィニッシュまでのフレームを含む
		newMinSpinDuration = bikeChargeAttackMinimumDuration
	}
	caFrames += newMinSpinDuration
	adjustedBikeChargeFinalHitmark := bikeChargeFinalHitmark + caFrames
	bikeHittableEntities, err := c.BuildBikeChargeAttackHittableTargetList()

	if err != nil {
		return action.Info{}, err
	}

	c.HoldBikeChargeAttack(newMinSpinDuration, skippedWindupFrames, bikeHittableEntities)

	src := c.caState.srcFrame
	c.QueueCharTask(func() {
		// キャラクターがフィールド上にいる必要がある
		if c.Core.Player.Active() != c.Index {
			return
		}

		// 重撃フィニッシュが元の重撃と一致するか確認
		if c.caState.srcFrame != src {
			return
		}

		// マヴイカがバイクに乗っている必要がある
		if c.armamentState != bike {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Flamestrider Charged Attack (Final)",
			AttackTag:        attacks.AttackTagExtra,
			AdditionalTags:   []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:           attacks.ICDTagMavuikaFlamestrider,
			ICDGroup:         attacks.ICDGroupDefault,
			StrikeType:       attacks.StrikeTypeBlunt,
			PoiseDMG:         120.0,
			Element:          attributes.Pyro,
			Durability:       25,
			Mult:             skillChargeFinal[c.TalentLvlSkill()],
			HitlagFactor:     0.01,
			HitlagHaltFrames: 0.04 * 60,
			IgnoreInfusion:   true,
			FlatDmg:          c.burstBuffCA() + c.c2BikeCA(),
		}

		radius := 4.0
		if c.StatusIsActive(burstKey) {
			radius = 4.5
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: 2},
				radius,
			),
			0,
			0,
		)

		// フィニッシャー着地時にc.caStateをリセット
		c.caState = ChargeState{}
	}, adjustedBikeChargeFinalHitmark)

	nightSoulDuration := c.GetRemainingNightSoulDuration()
	if nightSoulDuration <= adjustedBikeChargeFinalHitmark {
		// ダッシュキャンセルを考慮してヒットマークで退出
		c.QueueCharTask(func() {
			c.exitBike()
		}, adjustedBikeChargeFinalHitmark)

		c.QueueCharTask(func() {
			c.exitNightsoul()
		}, nightSoulDuration)
	}

	// 重撃フィニッシュモーション開始時に重撃フックを解除
	c.Core.Tasks.Add(func() {
		c.bikeChargeAttackUnhook()
	}, caFrames)

	return action.Info{
		Frames:          func(next action.Action) int { return bikeChargeFinalFrames[next] + caFrames },
		AnimationLength: bikeChargeFinalFrames[action.InvalidAction] + caFrames,
		CanQueueAfter:   bikeChargeFinalFrames[action.ActionDash] + caFrames,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) GetBikeChargeAttackAttackInfo() combat.AttackInfo {
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Flamestrider Charged Attack (Cyclic)",
		AttackTag:      attacks.AttackTagExtra,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagMavuikaFlamestrider,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       60.0,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skillCharge[c.TalentLvlSkill()],
		HitlagFactor:   0.01,
		// HitlagHaltFrames: 0.03 * 60,
		IsDeployable:   true,
		IgnoreInfusion: true,
		FlatDmg:        c.burstBuffCA() + c.c2BikeCA(),
	}
	return ai
}

func (c *char) GetSkippedWindupFrames(bufferedFrames int) int {
	x := c.Core.Player.CurrentState()
	var skippedWindupFrames int
	// TODO: 初期重撃フレームを固有速度用の別関数で処理する際にリファクタリング
	// 現在、角度/ヒットボックス追跡は生の重撃フレームを使用して位置を決定
	// 誤ったタイミングでこれを減算するとヒットが同期から外れる可能性がある
	switch {
	case x == action.DashState:
		skippedWindupFrames = 15
		// ゲーム内では稀に発動しないが、シムのフレームでは常に発生するはず
		c.Core.Events.Emit(event.OnStateChange, action.NormalAttackState, action.NormalAttackState)
		return skippedWindupFrames
	case x == action.NormalAttackState || x == action.ChargeAttackState && c.caState.StartFrame == c.Core.F:
		skippedWindupFrames = 15
	case x == action.BurstState:
		if bufferedFrames == 0 {
			skippedWindupFrames = 0
		} else {
			skippedWindupFrames = 15
		}
	// スキル再発動はスキルホールドと再発動から呼ばれる。再発動には強制的なn0フレームがある
	case x == action.SkillState && c.StatusIsActive(skillRecastCDKey):
		if c.StatusDuration(skillRecastCDKey) > 45 {
			skippedWindupFrames = 13
		} else {
			skippedWindupFrames = 15
		}
	case x == action.PlungeAttackState:
		skippedWindupFrames = 13
	}
	skippedWindupFrames = min(skippedWindupFrames, bufferedFrames)
	// ワインドアップが完全にスキップされない場合、マヴイカの重撃ワインドアップがYelan/XQ等のn0アビリティを発動する
	if skippedWindupFrames < 15 {
		c.Core.Events.Emit(event.OnStateChange, action.NormalAttackState, action.NormalAttackState)
	}
	return skippedWindupFrames
}

// 重撃のナイトソウル消費は11/s。skill.goの関数が6fごとに減少させる
func (c *char) GetRemainingNightSoulDuration() int {
	curPoints := c.nightsoulState.Points()
	framesSinceLastNSReduce := (c.Core.F - c.nightsoulSrc) % 6

	nsDur := int(math.Ceil(curPoints / 1.1))
	nsDur *= 6
	nsDur -= framesSinceLastNSReduce
	if c.StatusIsActive(burstKey) {
		nsDur += c.StatusDuration(burstKey)
	}

	return nsDur
}

// スピン中に重撃フィニッシュを開始できなや20fのウィンドウ
func GetCAFDelay(currentBikeSpinFrame int) int {
	newMinSpinDuration := 0

	if currentBikeSpinFrame < 10 {
		newMinSpinDuration = 10 - currentBikeSpinFrame
	} else if currentBikeSpinFrame >= 35 {
		newMinSpinDuration = 55 - currentBikeSpinFrame
	}
	return newMinSpinDuration
}

func (c *char) BuildBikeChargeAttackHittableTargetList() ([]HittableEntity, error) {
	targetList, hitboxError := c.buildValidTargetList()
	return append(targetList, c.buildValidGadgetList()...), hitboxError
}

func (c *char) buildValidTargetList() ([]HittableEntity, error) {
	enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8), nil)
	hittableEnemies := []HittableEntity{}
	for _, v := range enemies {
		if v == nil {
			continue
		}
		// 衝突の開始フレームと終了フレームを計算
		collisionFrames := [2]int{-1, -1}
		var facingDirection float64
		if c.caState.cAtkFrames == 0 {
			facingDirection = c.DirectionOffsetToPrimaryTarget()
			c.caState.FacingDirection = facingDirection
		} else {
			facingDirection = c.caState.FacingDirection
		}
		isIntersecting, err := c.BikeHitboxIntersectionAngles(v, collisionFrames[:], facingDirection)

		if err != nil {
			return hittableEnemies, err
		}

		if isIntersecting {
			hittableEnemies = append(hittableEnemies, HittableEntity{
				Entity:     combat.Target(v),
				isOneTick:  false,
				CollFrames: collisionFrames,
			})
		}
	}
	return hittableEnemies, nil
}

func (c *char) buildValidGadgetList() []HittableEntity {
	gadgets := c.Core.Combat.GadgetsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8), nil)
	var hittableGadgets []HittableEntity
	for _, g := range gadgets {
		if g == nil {
			continue
		}
		switch g.GadgetTyp() {
		case combat.GadgetTypDendroCore, combat.GadgetTypBogglecatBox:
			// 衝突の開始フレームと終了フレームを計算
			// これらのガジェットは円形ヒットボックスを持つため、ヒットボックス形状エラーは無視できる
			hittableGadget, isHittable, _ := c.IsGadgetHittable(g)
			if isHittable {
				hittableGadgets = append(hittableGadgets, hittableGadget)
			}
		case combat.GadgetTypLeaLotus:
			hittableGadget, isHittable, _ := c.IsGadgetHittable(g)
			if isHittable {
				hittableGadget.isOneTick = false
				hittableGadgets = append(hittableGadgets, hittableGadget)
			}
		}
	}
	return hittableGadgets
}

func (c *char) IsGadgetHittable(v combat.Gadget) (HittableEntity, bool, error) {
	collisionFrames := [2]int{-1, -1}
	var facingDirection float64
	if c.caState.cAtkFrames == 0 {
		facingDirection = c.DirectionOffsetToPrimaryTarget()
		c.caState.FacingDirection = facingDirection
	} else {
		facingDirection = c.caState.FacingDirection
	}
	isIntersecting, hitboxError := c.BikeHitboxIntersectionAngles(v, collisionFrames[:], facingDirection)
	newGadget := HittableEntity{}

	if isIntersecting {
		newGadget = HittableEntity{
			Entity:     combat.Target(v),
			isOneTick:  true,
			CollFrames: collisionFrames,
		}
	}
	return newGadget, isIntersecting, hitboxError
}

func (c *char) HasValidTargetCheck(bikeHittableEntities []HittableEntity) (bool, action.Info, error) {
	isTargetForCountsHittable := false
	if len(bikeHittableEntities) == 0 {
		c.SetHittableEntityList(bikeHittableEntities)
		return false, action.Info{}, errors.New("no valid targets within flamestrider area")
	}
	for _, t := range bikeHittableEntities {
		if t.Entity == c.Core.Combat.PrimaryTarget() {
			isTargetForCountsHittable = true
			break
		}
	}
	if !isTargetForCountsHittable {
		return false, action.Info{}, errors.New("primary target is not within flamestrider area")
	}
	return true, action.Info{}, nil
}

// 現在は草原核のスポーンに使用。その他の移動や追加は重撃アニメーション中に発生すべきではない
func (c *char) bikeChargeAttackHook() {
	c.Core.Events.Subscribe(event.OnDendroCore, func(args ...interface{}) bool {
		// バイク状態でない場合は無視
		if c.armamentState != bike && !c.nightsoulState.HasBlessing() {
			return false
		}
		// バイク状態の場合、ヒット可能ならガジェットをターゲットリストに追加
		g, ok := args[0].(combat.Gadget)
		if !ok {
			return false
		}
		if g.GadgetTyp() == combat.GadgetTypDendroCore {
			// リストへの追加は不要かもしれない？
			hittableGadget, isHittable, _ := c.IsGadgetHittable(g)
			if isHittable {
				remainingCADuration := c.caState.cAtkFrames - (c.Core.F - c.caState.StartFrame)
				hitFrames := c.CalculateValidCollisionFrames(remainingCADuration, hittableGadget.CollFrames, 0)
				if len(hitFrames) > 0 {
					for _, f := range hitFrames {
						c.Core.Tasks.Add(func() {
							ai := c.GetBikeChargeAttackAttackInfo()
							c.Core.QueueAttack(ai, combat.NewSingleTargetHit(hittableGadget.Entity.Key()), 0, 0)
						}, f)
					}
				}
				// フレームは0より大きければ正確な値でなくても問題ない
				c.caState.LastHit[g.Key()] += c.Core.F
			}
		}

		return false
	}, "mavuika-bike-gadget-check")
}

func (c *char) bikeChargeAttackUnhook() {
	c.Core.Events.Unsubscribe(event.OnDendroCore, "mavuika-bike-gadget-check")
}

func (*char) SetHittableEntityList(bikeHittableEntities []HittableEntity) {
	bikeHittableEntityList = bikeHittableEntities
}

func (*char) GetHittableEntityList() []HittableEntity {
	return bikeHittableEntityList
}

// 重撃フレームを反復し、ヒットマークから開始
func (c *char) CalculateValidCollisionFrames(durationCA int, collisionFrames [2]int, lastHitFrame int) []int {
	validFrames := []int{}
	currentFrame := bikeChargeAttackStartupHitmark // スピンヒットボックスは重撃アニメーションの35f目から開始（フルワインドアップ）	var timeSinceStart int

	// 重撃が前のアクションから継続している場合、現在のサイクルを調整
	timeSinceStart := c.Core.F - (c.caState.StartFrame - c.caState.skippedWindupF)
	timeSinceLastHit := timeSinceStart - lastHitFrame
	if timeSinceStart >= currentFrame {
		currentFrame = timeSinceStart
		if timeSinceLastHit < bikeChargeAttackICD {
			currentFrame += bikeChargeAttackICD - timeSinceLastHit
		}
	}
	totalFrames := currentFrame                // 経過した総フレーム数を追跡
	currentFrame %= bikeChargeAttackSpinFrames // スピンサイクル内で現在のフレームを開始

	collisionStart := collisionFrames[0]
	collisionEnd := collisionFrames[1]

	for totalFrames <= (durationCA + c.caState.cAtkFrames) {
		checkValidFrame := -1
		// フレームが衝突範囲外の場合、前方にシフト
		if collisionStart <= collisionEnd {
			if currentFrame > collisionEnd {
				currentFrame -= bikeChargeAttackSpinFrames
				totalFrames += collisionStart - currentFrame
				currentFrame = collisionStart
			} else if currentFrame < collisionStart {
				totalFrames += collisionStart - currentFrame
				currentFrame = collisionStart
			}
		} else {
			if currentFrame < collisionStart && currentFrame > collisionEnd {
				totalFrames += collisionStart - currentFrame
				currentFrame = collisionStart
			}
		}

		if collisionStart <= collisionEnd {
			if currentFrame >= collisionStart && currentFrame <= collisionEnd {
				checkValidFrame = totalFrames - timeSinceStart
			}
		} else {
			// collisionEndがcollisionStartより前のラップケースを処理
			if currentFrame >= collisionStart || currentFrame <= collisionEnd {
				checkValidFrame = totalFrames - timeSinceStart
			}
		}
		// 初回重撃ヒット計算では、スキップされたワインドアップフレームを考慮
		if c.Core.F == c.caState.StartFrame {
			checkValidFrame += c.caState.skippedWindupF
		}

		if checkValidFrame >= 0 && checkValidFrame <= durationCA {
			validFrames = append(validFrames, checkValidFrame)
		}

		// クールダウンフレーム分前進し、スピンアニメーション長内でラップ
		totalFrames += bikeChargeAttackICD
		currentFrame = (currentFrame + bikeChargeAttackICD) % bikeChargeAttackSpinFrames
	}

	return validFrames
}

// 各スピン中にターゲットがマヴイカのヒットボックス内にある開始フレームと終了フレームを計算
// ターゲットが円形でないか重なりがない場合はfalseを返す
func (c *char) BikeHitboxIntersectionAngles(v combat.Target, f []int, offsetAngle float64) (bool, error) {
	enemyShape := v.Shape()
	var enemyRadius float64
	switch v := enemyShape.(type) {
	case *geometry.Circle:
		enemyRadius = v.Radius() // Rt
	default:
		return false, errors.New("target has non-circular hitbox, Mavuika CA requires circle hitboxes for calculations")
	}

	bikeInnerRadius := bikeChargeAttackSpinOffset - bikeChargeAttackHitboxRadius // Ri
	bikeOuterRadius := bikeChargeAttackSpinOffset + bikeChargeAttackHitboxRadius // Ro

	posDifference := v.Pos().Sub(c.Core.Combat.Player().Pos())
	enemyDistance := posDifference.Magnitude() // Dt

	// 重なりがないか確認
	if enemyDistance+enemyRadius <= bikeInnerRadius || enemyDistance-enemyRadius >= bikeOuterRadius {
		return false, nil
	}

	// ターゲットが回転全体で常にヒットボックス範囲内にある
	if enemyRadius-enemyDistance > bikeInnerRadius {
		f[0] = 0
		f[1] = bikeChargeAttackSpinFrames
		return true, nil
	}

	sumRadii := bikeChargeAttackHitboxRadius + enemyRadius
	cosThetaM := (bikeChargeAttackSpinOffset*bikeChargeAttackSpinOffset + enemyDistance*enemyDistance - sumRadii*sumRadii) /
		(2 * bikeChargeAttackSpinOffset * enemyDistance)

	enemyAngle := math.Atan2(posDifference.Y, posDifference.X) * (180 / math.Pi)
	thetaM := math.Acos(cosThetaM) * (180 / math.Pi)

	enemyAngle = math.Mod(enemyAngle-offsetAngle+360, 360)

	intersectAngleStart := enemyAngle + thetaM
	intersectAngleEnd := enemyAngle - thetaM

	f[0] = c.ConvertAngleToFrame(intersectAngleStart)
	f[1] = c.ConvertAngleToFrame(intersectAngleEnd)

	return true, nil
}

func (c *char) DirectionOffsetToPrimaryTarget() float64 {
	var enemyDirection = geometry.CalcDirection(c.Core.Combat.Player().Pos(), c.Core.Combat.PrimaryTarget().Pos())
	if enemyDirection == geometry.DefaultDirection() {
		return 0
	}

	angleToTarget := math.Atan2(enemyDirection.X, enemyDirection.Y) * (180 / math.Pi)
	offsetAngle := 360 - angleToTarget

	return math.Mod(offsetAngle+360, 360)
}

func (c *char) ConvertAngleToFrame(theta float64) int {
	theta = math.Mod(theta+360, 360)

	var quadrant int
	var spinQuadrant int
	var accumulatedFrames int

	switch {
	case theta >= 270 || theta < 0:
		quadrant = 3
		spinQuadrant = 0
		accumulatedFrames = 0
	case theta >= 180:
		quadrant = 2
		spinQuadrant = 1
		accumulatedFrames = bikeSpinQuadrantFrames[0]
	case theta >= 90:
		quadrant = 1
		spinQuadrant = 2
		accumulatedFrames = bikeSpinQuadrantFrames[1] + bikeSpinQuadrantFrames[0]
	default:
		quadrant = 0
		spinQuadrant = 3
		accumulatedFrames = bikeSpinQuadrantFrames[2] + bikeSpinQuadrantFrames[1] + bikeSpinQuadrantFrames[0]
	}

	if accumulatedFrames > 0 {
		accumulatedFrames-- // スピンフレームカウントが0から始まることを考慮
	}

	// 象限内のフレームを計算
	quadrantStartAngle := float64(quadrant) * 90.0
	frameOffset := float64(bikeSpinQuadrantFrames[spinQuadrant]) + (theta-quadrantStartAngle)/bikeSpinQuadrantAngularVelocity[spinQuadrant]

	return accumulatedFrames + int(math.Round(frameOffset))
}
