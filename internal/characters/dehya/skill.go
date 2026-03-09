package dehya

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var (
	skillFrames       []int
	skillRecastFrames []int
)

const (
	skillHitmark           = 20
	skillRecastHitmark     = 40
	skillDoTAbil           = "Molten Inferno (DoT)"
	skillMitigationAbil    = "Fiery Sanctum Mitigation"
	skillSelfDoTAbil       = "Redmane's Blood"
	skillSelfDoTStatus     = "dehya-redmanes-blood"
	skillSelfDoTStart      = 0.1 * 60 // 最初のDoTティックは軽減から0.1秒後に発生する模様
	skillSelfDoTDuration   = 10 * 60  // 合訐10回のティック（0.1秒、1.1秒、…、9.1秒）
	skillSelfDoTRatio      = 0.1
	skillSelfDoTInterval   = 1 * 60
	skillICDKey            = "dehya-skill-icd"
	dehyaFieldKey          = "dehya-field-status"
	dehyaFieldDuration     = 12 * 60
	sanctumPickupExtension = 24 // 元素爆発/元素スキル再発動時に領域持続時間が0.4秒延長される
)

func init() {
	skillFrames = frames.InitAbilSlice(39) // E -> N1
	skillFrames[action.ActionSkill] = 30
	skillFrames[action.ActionBurst] = 29
	skillFrames[action.ActionDash] = 26
	skillFrames[action.ActionJump] = 28
	skillFrames[action.ActionSwap] = 25

	skillRecastFrames = frames.InitAbilSlice(74) // E -> N1
	skillRecastFrames[action.ActionSkill] = 45
	skillRecastFrames[action.ActionBurst] = 45
	skillRecastFrames[action.ActionDash] = 45
	skillRecastFrames[action.ActionJump] = 49
	skillRecastFrames[action.ActionSwap] = 44
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	burstAction := c.UseBurstAction()
	if burstAction != nil {
		return *burstAction, nil
	}
	needPickup := false
	if c.StatusIsActive(dehyaFieldKey) {
		// 再発動が使用済みの場合、再設置前に浄焔の領域を回収する必要がある
		if c.hasRecastSkill {
			needPickup = true
		} else {
			c.hasRecastSkill = true
			return c.skillRecast()
		}
	}

	c.hasRecastSkill = false
	c.hasC2DamageBuff = false

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Molten Inferno",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   50,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
		FlatDmg:    c.c1FlatDmgRatioE * c.MaxHP(),
	}
	// TODO: ダメージフレーム
	c.skillSnapshot = c.Snapshot(&ai)

	// 初撃を実行
	player := c.Core.Combat.Player()
	skillPos := geometry.CalcOffsetPoint(c.Core.Combat.Player().Pos(), geometry.Point{Y: 0.8}, player.Direction())
	c.skillArea = combat.NewCircleHitOnTarget(skillPos, nil, 10)
	c.Core.QueueAttackWithSnap(ai, c.skillSnapshot, combat.NewCircleHitOnTarget(skillPos, nil, 5), skillHitmark)

	// 領域を処理
	c.AddStatus(skillICDKey, skillHitmark+1, false) // 元素スキルICDを追加し、初撃で領域が発動しないようにする
	if needPickup {
		c.Core.Tasks.Add(func() { c.pickUpField() }, skillHitmark-1)
	}
	c.Core.Tasks.Add(func() { c.addField(dehyaFieldDuration) }, skillHitmark+1)

	c.SetCDWithDelay(action.ActionSkill, 20*60, 18)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap],
		State:           action.SkillState,
	}, nil
}

func (c *char) skillDmgHook() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		trg := args[0].(combat.Target)
		// atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		if !c.StatusIsActive(dehyaFieldKey) {
			return false
		}
		if c.StatusIsActive(skillICDKey) {
			return false
		}
		if dmg == 0 {
			return false
		}
		// 命中したターゲットがスキル範囲外なら発動しない
		if !trg.IsWithinArea(c.skillArea) {
			return false
		}

		// このICDはおそらく設置物に紐づいているため、ヒットラグで延長されない
		c.AddStatus(skillICDKey, 2.5*60, false)

		c.Core.QueueAttackWithSnap(
			c.skillAttackInfo,
			c.skillSnapshot,
			combat.NewCircleHitOnTarget(trg, nil, 4.5),
			2,
		)

		// DoTヒット直後に遅延3フレームでバフフラグをfalseに設定
		if c.hasC2DamageBuff {
			c.Core.Tasks.Add(func() { c.hasC2DamageBuff = false }, 3)
		}

		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Pyro, c.ParticleDelay)

		return false
	}, "dehya-skill")
}

func (c *char) skillRecast() (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Ranging Flame",
		AttackTag:        attacks.AttackTagElementalArt,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		PoiseDMG:         50,
		Element:          attributes.Pyro,
		Durability:       25,
		Mult:             skillReposition[c.TalentLvlSkill()],
		FlatDmg:          c.c1FlatDmgRatioE * c.MaxHP(),
		HitlagHaltFrames: 0.02 * 60,
		HitlagFactor:     0.01,
	}

	// 開始時に領域を回収
	c.pickUpField()

	// ICD延長を追加
	c.AddStatus(skillICDKey, skillRecastHitmark+c.sanctumICD, false)

	// 再配置

	// TODO: ダメージフレーム

	player := c.Core.Combat.Player()
	// 元素スキル単押しのヒットボックスオフセットを想定
	skillPos := geometry.CalcOffsetPoint(c.Core.Combat.Player().Pos(), geometry.Point{Y: 0.5}, player.Direction())
	c.skillArea = combat.NewCircleHitOnTarget(skillPos, nil, 10)
	c.Core.QueueAttackWithSnap(ai, c.skillSnapshot, combat.NewCircleHitOnTarget(skillPos, nil, 6), skillRecastHitmark)

	// 領域を再設置
	c.Core.Tasks.Add(func() { // 領域を設置
		c.c2IncreaseDur()
		c.addField(c.sanctumSavedDur)
	}, skillRecastHitmark+1)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillRecastFrames),
		AnimationLength: skillRecastFrames[action.InvalidAction],
		CanQueueAfter:   skillRecastFrames[action.ActionSwap], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

// 領域を回収し、現在のICDと持続時間を暗黙的な延長付きで保存
func (c *char) pickUpField() {
	c.a1Reduction()
	c.sanctumICD = c.StatusDuration(skillICDKey)
	c.sanctumSavedDur = c.StatusDuration(dehyaFieldKey) + sanctumPickupExtension // 領域再設置時にわずかに持続時間が延長される模様
	c.Core.Log.NewEvent("sanctum picked up", glog.LogCharacterEvent, c.Index).
		Write("Duration Remaining", c.sanctumSavedDur).
		Write("DoT tick CD", c.sanctumICD)
	c.Core.Tasks.Add(func() {
		c.DeleteStatus(dehyaFieldKey)
	}, 1)
}

func (c *char) addField(dur int) {
	// 領域を設置する
	c.AddStatus(dehyaFieldKey, dur, false)
	c.Core.Log.NewEvent("sanctum added", glog.LogCharacterEvent, c.Index).
		Write("Duration Remaining", dur).
		Write("New Expiry Frame", c.StatusExpiry(dehyaFieldKey)).
		Write("DoT tick CD", c.StatusDuration(skillICDKey))

	// Tick用のスナップショット
	c.skillAttackInfo = combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             skillDoTAbil,
		AttackTag:        attacks.AttackTagElementalArt,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Pyro,
		Durability:       25,
		Mult:             skillDotAtk[c.TalentLvlSkill()],
		FlatDmg:          (c.c1FlatDmgRatioE + skillDotHP[c.TalentLvlSkill()]) * c.MaxHP(),
		HitlagHaltFrames: 0.02 * 60,
		HitlagFactor:     0.01,
		IsDeployable:     true,
	}
	c.skillSnapshot = c.Snapshot(&c.skillAttackInfo)
}

// この領域内のアクティブキャラクターは中断耐性が上昇する（未実装）、
// そのようなキャラクターがダメージを受けると、そのダメージの一部が軽減され赤鬣の血に流れ込む。
// ディヘヤはこのダメージを10秒かけて受ける。赤鬣の血に蓄積された軽減ダメージが
// ディヘヤのHP上限の一定割合に達するか超えると、この方法によるダメージ軽減を停止する。
func (c *char) skillHurtHook() {
	// 真のダメージを軽減する
	// 侵食を軽減すべきではない（おそらくシムには追加されない…）
	c.Core.Events.Subscribe(event.OnPlayerPreHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		// 外部ダメージのみ軽減
		if !di.External {
			return false
		}
		// 0ダメージなら躽減不要
		if di.Amount <= 0 {
			return false
		}
		// 躽減には領域がアクティブである必要がある
		if !c.StatusIsActive(dehyaFieldKey) {
			return false
		}
		// 躽減にはプレイヤーが領域内にいる必要がある
		if !c.Core.Combat.Player().IsWithinArea(c.skillArea) {
			return false
		}
		// 自己DoTを無視
		if di.Abil == skillSelfDoTAbil {
			return false
		}
		// 閾値に達したらダメージ躽減を停止
		if c.skillRedmanesBlood >= 2*c.MaxHP() {
			return false
		}
		beforeAmount := di.Amount
		// 天賦レベルに基づいて躽減量を計算
		mitigation := di.Amount * skillMitigation[c.TalentLvlSkill()]
		// 赤鬣の血を調整
		c.skillRedmanesBlood += mitigation
		// HPドレインを変更
		di.Amount = max(di.Amount-mitigation, 0)
		// 躽減をログ出力
		c.Core.Log.NewEvent("dehya mitigating dmg", glog.LogCharacterEvent, c.Index).
			Write("hurt_before", beforeAmount).
			Write("mitigation", mitigation).
			Write("hurt", di.Amount)
		// 自己DoTステータスを追加
		c.AddStatus(skillSelfDoTStatus, skillSelfDoTDuration, true)
		// まだキューに入っていなければDoTをキューに追加
		// -> 再トリガーは間隔をリセットしないはず（未確認）
		// -> そうでないと、DoTティック間に躽減し続けた場合ディヘヤにダメージが入らなくなる
		if c.skillSelfDoTQueued {
			return false
		}
		c.skillSelfDoTQueued = true
		c.QueueCharTask(c.skillSelfDoT, skillSelfDoTStart)
		return false
	}, "dehya-field-dmgtaken")
}

func (c *char) skillSelfDoT() {
	if !c.StatusIsActive(skillSelfDoTStatus) {
		c.skillSelfDoTQueued = false
		return
	}

	// 次のティックをキューに追加
	c.QueueCharTask(c.skillSelfDoT, skillSelfDoTInterval)

	// 元素爆発の無敵フレーム中は自己DoTを実行しない
	if c.Core.Player.Active() == c.Index && c.Core.Player.CurrentState() == action.BurstState {
		return
	}

	// 毎ティックでダメージを再計算
	dmg := c.skillRedmanesBlood * skillSelfDoTRatio

	// 赤鬣の血を減少（シールド躽減/A1を考慮する前に！）
	c.skillRedmanesBlood = max(c.skillRedmanesBlood-dmg, 0)

	// A1がアクティブならダメージを変更（赤鬣の血はこのチェック前に全額減少済み）
	if c.StatusIsActive(a1ReductionKey) {
		dmgBefore := dmg
		dmg *= 1 - a1ReductionMult
		c.Core.Log.NewEvent("dehya a1 reducing redmane's blood dmg", glog.LogCharacterEvent, c.Index).
			Write("dmg_before", dmgBefore).
			Write("dmg", dmg)
	}

	// 自己DoTを実行
	// TODO: システムはオフフィールドのキャラクターに直接ヒットするよう設計されていないためのハック
	// これは真の物理ダメージなのでダメージ式/元素耐性は関係ない
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       skillSelfDoTAbil,
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
		Durability: 0,
		FlatDmg:    dmg,
	}
	ap := combat.NewSingleTargetHit(c.Core.Combat.Player().Key())
	snap := c.Snapshot(&ai)
	ae := &combat.AttackEvent{
		Info:        ai,
		Pattern:     ap,
		Snapshot:    snap,
		SourceFrame: c.Core.F,
	}

	c.Core.Combat.Events.Emit(event.OnPlayerHit, c.Index, ae)
	dmgLeft := c.Core.Player.Shields.OnDamage(c.Index, c.Core.Player.Active(), dmg, ae.Info.Element)
	if dmgLeft > 0 {
		c.Core.Player.Drain(info.DrainInfo{
			ActorIndex: c.Index,
			Abil:       ae.Info.Abil,
			Amount:     dmgLeft,
			External:   true,
		})
	}
}
