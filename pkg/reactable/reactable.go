package reactable

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

type Modifier int

const (
	Invalid Modifier = iota
	Electro
	Pyro
	Cryo
	Hydro
	BurningFuel
	SpecialDecayDelim
	Dendro
	Quicken
	Frozen
	Anemo
	Geo
	Burning
	EndModifier
)

var modifierElement = []attributes.Element{
	attributes.UnknownElement,
	attributes.Electro,
	attributes.Pyro,
	attributes.Cryo,
	attributes.Hydro,
	attributes.Dendro,
	attributes.UnknownElement,
	attributes.Dendro,
	attributes.Quicken,
	attributes.Frozen,
	attributes.Anemo,
	attributes.Geo,
	attributes.Pyro,
	attributes.UnknownElement,
}

var ModifierString = []string{
	"",
	"electro",
	"pyro",
	"cryo",
	"hydro",
	"dendro-fuel",
	"",
	"dendro",
	"quicken",
	"frozen",
	"anemo",
	"geo",
	"burning",
	"",
}

var elementToModifier = map[attributes.Element]Modifier{
	attributes.Electro: Electro,
	attributes.Pyro:    Pyro,
	attributes.Cryo:    Cryo,
	attributes.Hydro:   Hydro,
	attributes.Dendro:  Dendro,
}

func (r Modifier) Element() attributes.Element { return modifierElement[r] }
func (r Modifier) String() string              { return ModifierString[r] }

func (r Modifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(ModifierString[r])
}

func (r *Modifier) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(s)
	for i, v := range ModifierString {
		if v == s {
			*r = Modifier(i)
			return nil
		}
	}
	return errors.New("unrecognized ReactableModifier")
}

type Reactable struct {
	Durability [EndModifier]reactions.Durability
	DecayRate  [EndModifier]reactions.Durability
	// Source     []int // オーラの発生フレーム
	self combat.Target
	core *core.Core
	// 感電反応固有
	ecAtk      combat.AttackInfo // 次の感電反応ティックの所有者のインデックス
	ecSnapshot combat.Snapshot
	ecTickSrc  int
	// 燃焼反応固有
	burningAtk      combat.AttackInfo
	burningSnapshot combat.Snapshot
	burningTickSrc  int
	// 凍結反応固有
	FreezeResist float64
	// GCD固有
	overloadGCD     int
	shatterGCD      int
	superconductGCD int
	swirlElectroGCD int
	swirlHydroGCD   int
	swirlCryoGCD    int
	swirlPyroGCD    int
	crystallizeGCD  int

	lcContributor        []int
	lcPrecalcDamages     []lcDamageRecord
	lcPrecalcDamagesCRIT []lcDamageRecord
	lcTickSrc            int
	lcActiveExpiry       int
	lastEleSource        map[attributes.Element]int
	expiryTaskMap        map[int]int
	// ルナチャージ雲の状態
	lcCloudActive bool
	lcCloudExpiry int
	// ルナ結晶化の状態
	lcrsContributor        []int
	lcrsPrecalcDamages     []lcDamageRecord
	lcrsPrecalcDamagesCRIT []lcDamageRecord
	lcrsTickSrc            int
	lcrsActiveExpiry       int
	lcrsExpiryTaskMap      map[int]int
	lcrsTriggerCount       int
}

type Enemy interface {
	QueueEnemyTask(f func(), delay int)
}

const frzDelta reactions.Durability = 2.5 / (60 * 60) // 2 * 1.25
const frzDecayCap reactions.Durability = 10.0 / 60.0

const ZeroDur reactions.Durability = 0.00000000001

func (r *Reactable) Init(self combat.Target, c *core.Core) *Reactable {
	r.self = self
	r.core = c
	r.DecayRate[Frozen] = frzDecayCap
	r.ecTickSrc = -1
	r.lcTickSrc = -1
	r.lcrsTickSrc = -1
	r.burningTickSrc = -1
	r.overloadGCD = -1
	r.shatterGCD = -1
	r.superconductGCD = -1
	r.swirlElectroGCD = -1
	r.swirlHydroGCD = -1
	r.swirlCryoGCD = -1
	r.swirlPyroGCD = -1
	r.crystallizeGCD = -1
	return r
}

func (r *Reactable) ActiveAuraString() []string {
	var result []string
	for i, v := range r.Durability {
		if v > ZeroDur {
			result = append(result, Modifier(i).String()+": "+strconv.FormatFloat(float64(v), 'f', 3, 64))
		}
	}
	return result
}

func (r *Reactable) AuraCount() int {
	count := 0
	for _, v := range r.Durability {
		if v > ZeroDur {
			count++
		}
	}
	return count
}

func (r *Reactable) React(a *combat.AttackEvent) {
	// TODO: 元素反応の順序を再確認
	switch a.Info.Element {
	case attributes.Electro:
		// 超開花
		r.TryAggravate(a)
		r.TryOverload(a)
		r.TryAddEC(a)
		r.TryFrozenSuperconduct(a)
		r.TrySuperconduct(a)
		r.TryQuicken(a)
	case attributes.Pyro:
		// 烈開花
		r.TryOverload(a)
		r.TryVaporize(a)
		r.TryMelt(a)
		r.TryBurning(a)
	case attributes.Cryo:
		r.TrySuperconduct(a)
		r.TryMelt(a)
		r.TryFreeze(a)
	case attributes.Hydro:
		r.TryVaporize(a)
		r.TryFreeze(a)
		r.TryBloom(a)
		r.TryAddEC(a)
	case attributes.Anemo:
		r.TrySwirlElectro(a)
		r.TrySwirlPyro(a)
		r.TrySwirlHydro(a)
		r.TrySwirlCryo(a)
		r.TrySwirlFrozen(a)
	case attributes.Geo:
		// 結晶化は二重に発生しないようだ
		// 凍結は先に水元素を発動できる
		//https://docs.google.com/spreadsheets/d/1lJSY2zRIkFDyLZxIor0DVMpYXx3E_jpDrSUZvQijesc/edit#gid=0
		r.TryCrystallizeElectro(a)
		r.TryCrystallizeHydro(a)
		r.TryCrystallizeCryo(a)
		r.TryCrystallizePyro(a)
		r.TryCrystallizeFrozen(a)
	case attributes.Dendro:
		r.TrySpread(a)
		r.TryQuicken(a)
		r.TryBurning(a)
		r.TryBloom(a)
	}
}

// AttachOrRefill はダメージイベント後に、攻撃が何とも反応しなかった場合に呼ばれる。
// 修飾子が存在しなければ新規作成し、存在すれば各修飾子のルールに従って更新する。
func (r *Reactable) AttachOrRefill(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if a.Reacted {
		return false
	}
	// 炎、雷、水、氷、草を処理する
	// 草の特殊付着（燃焼燃料）は tryBurning で処理される
	if mod, ok := elementToModifier[a.Info.Element]; ok {
		r.attachOrRefillNormalEle(mod, a.Info.Durability)
		return true
	}
	return false
}

// attachOrRefillNormalEle は特殊な付着ルールを持たない炎、雷、水、氷、草に使用される
func (r *Reactable) attachOrRefillNormalEle(mod Modifier, dur reactions.Durability) {
	amt := 0.8 * dur
	if mod == Pyro {
		r.attachOverlapRefreshDuration(Pyro, amt, 6*dur+420)
	} else {
		r.attachOverlap(mod, amt, 6*dur+420)
	}
}

func (r *Reactable) attachOverlap(mod Modifier, amt, length reactions.Durability) {
	if r.Durability[mod] > ZeroDur {
		add := max(amt-r.Durability[mod], 0)
		if add > 0 {
			r.addDurability(mod, add)
		}
	} else {
		r.Durability[mod] = amt
		if length > ZeroDur {
			r.DecayRate[mod] = amt / length
		}
	}
}

func (r *Reactable) attachOverlapRefreshDuration(mod Modifier, amt, length reactions.Durability) {
	if amt < r.Durability[mod] {
		return
	}
	r.Durability[mod] = amt
	r.DecayRate[mod] = amt / length
}

func (r *Reactable) attachBurning() {
	r.Durability[Burning] = 50
	r.DecayRate[Burning] = 0
}

func (r *Reactable) addDurability(mod Modifier, amt reactions.Durability) {
	r.Durability[mod] += amt
	r.core.Events.Emit(event.OnAuraDurabilityAdded, r.self, mod, amt)
}

// AuraContains は対象に元素eがいずれか付着している場合にtrueを返す
func (r *Reactable) AuraContains(e ...attributes.Element) bool {
	for _, v := range e {
		for i := Invalid; i < EndModifier; i++ {
			if i.Element() == v && r.Durability[i] > ZeroDur {
				return true
			}
		}
		//TODO: この方法が最善か不明。凍結元素を供給する方が良いかもしれない
		if v == attributes.Cryo && r.Durability[Frozen] > ZeroDur {
			return true
		}
	}
	return false
}

func (r *Reactable) IsBurning() bool {
	if r.Durability[BurningFuel] > ZeroDur && r.Durability[Burning] > ZeroDur {
		return true
	}
	return false
}

// 指定された元素を dur * factor 分減少させ、消費された元素量を返す
// 同じ元素の修飾子が複数ある場合、すべて減少する
// 減少量の最大値が消費量として使用される
func (r *Reactable) reduce(e attributes.Element, dur, factor reactions.Durability) reactions.Durability {
	m := dur * factor // maximum amount reduceable
	var reduced reactions.Durability

	for i := Invalid; i < EndModifier; i++ {
		if i.Element() != e {
			continue
		}
		if r.Durability[i] < ZeroDur {
			// 元素量が既に0の場合もスキップする
			// これにより元素が存在しない場合でも安全に reduce を呼べる
			continue
		}
		// 残量と m の小さい方で減少させる

		red := m

		if red > r.Durability[i] {
			red = r.Durability[i]
			// 減衰速度を0にリセット
		}

		r.Durability[i] -= red

		if red > reduced {
			reduced = red
		}
	}

	return reduced / factor
}

func (r *Reactable) deplete(m Modifier) {
	if r.Durability[m] <= ZeroDur {
		r.Durability[m] = 0
		r.DecayRate[m] = 0
		r.core.Events.Emit(event.OnAuraDurabilityDepleted, r.self, attributes.Element(m))
	}
}

func (r *Reactable) Tick() {
	// 元素量は decay * (1 + purge) で減少する
	// purge は凍結以外では0
	// 凍結の場合、purge = 0.25 * time
	// 解凍中は purge 速度が new = old - 0.5 * time で0に戻る
	// time は秒単位
	//
	// フレーム単位では decay * (1 + 0.25 * (x/60)) となる

	// 区切り以降は特殊減衰なので無視する
	for i := Invalid; i < SpecialDecayDelim; i++ {
		// 減衰速度0の修飾子（燃焼など減衰しないもの）をスキップ
		if r.DecayRate[i] == 0 {
			continue
		}
		if r.Durability[i] > ZeroDur {
			r.Durability[i] -= r.DecayRate[i]
			r.deplete(i)
		}
	}

	// 燃焼を先にチェックする（草/激化の減衰に影響するため）
	if r.burningTickSrc > -1 && r.Durability[BurningFuel] < ZeroDur {
		// 燃焼燃料がなくなったらソースをリセット
		r.burningTickSrc = -1
		// 燃焼を除去
		r.Durability[Burning] = 0
		// 既存の草元素と激化を除去
		r.Durability[Dendro] = 0
		r.DecayRate[Dendro] = 0
		r.Durability[Quicken] = 0
		r.DecayRate[Quicken] = 0
	}

	// 燃焼燃料が存在する場合、草と激化は燃焼燃料の減衰速度を使用する
	// そうでなければ自身の減衰速度を使用する
	for i := Dendro; i <= Quicken; i++ {
		if r.Durability[i] < ZeroDur {
			continue
		}
		rate := r.DecayRate[i]
		if r.Durability[BurningFuel] > ZeroDur {
			rate = r.DecayRate[BurningFuel]
			if i == Dendro {
				rate = max(rate, r.DecayRate[i]*2)
			}
		}
		r.Durability[i] -= rate
		r.deplete(i)
	}

	// 凍結の場合、元素量は以下で計算できる:
	// d_f(t) = -1.25 * (t/60)^2 - k * (t/60) + d_f(0)
	if r.Durability[Frozen] > ZeroDur {
		// まず減衰速度を上昇させる
		r.DecayRate[Frozen] += frzDelta
		r.Durability[Frozen] -= r.DecayRate[Frozen] / reactions.Durability(1.0-r.FreezeResist)

		r.checkFreeze()
	} else if r.DecayRate[Frozen] > frzDecayCap { // そうでなければ減衰速度を低下させる
		r.DecayRate[Frozen] -= frzDelta * 2

		// 減衰の上限
		if r.DecayRate[Frozen] < frzDecayCap {
			r.DecayRate[Frozen] = frzDecayCap
		}
	}

	// 感電反応がなくなった場合、ソースをリセットする必要がある
	if r.ecTickSrc > -1 {
		if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
			r.ecTickSrc = -1
		}
	}

	if r.lcTickSrc > -1 {
		if r.Durability[Electro] < ZeroDur {

		}
	}
}

func calcReactionDmg(char *character.CharWrapper, atk combat.AttackInfo, em float64) (float64, combat.Snapshot) {
	lvl := char.Base.Level - 1
	if lvl > 89 {
		lvl = 89
	}
	if lvl < 0 {
		lvl = 0
	}
	snap := combat.Snapshot{
		CharLvl: char.Base.Level,
	}
	snap.Stats[attributes.EM] = em
	return (1 + ((16 * em) / (2000 + em)) + char.ReactBonus(atk)) * reactionLvlBase[lvl], snap
}

func (r *Reactable) calcCatalyzeDmg(atk combat.AttackInfo, em float64) float64 {
	char := r.core.Player.ByIndex(atk.ActorIndex)
	lvl := char.Base.Level - 1
	if lvl > 89 {
		lvl = 89
	}
	if lvl < 0 {
		lvl = 0
	}
	return (1 + ((5 * em) / (1200 + em)) + r.core.Player.ByIndex(atk.ActorIndex).ReactBonus(atk)) * reactionLvlBase[lvl]
}

func calcLunarChargedDmg(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 89 {
		lvl = 89
	}
	if lvl < 0 {
		lvl = 0
	}
	return 1.8 * (reactionLvlBase[lvl] * (1 + char.LCBaseReactBonus(atk))) * (1 + ((6 * em) / (2000 + em)) + char.LCReactBonus(atk)) * (1 + char.ElevationBonus(atk))
}

func calcLunarChargedDmgCRIT(char *character.CharWrapper, atk combat.AttackInfo, em float64) float64 {
	lvl := char.Base.Level - 1
	if lvl > 89 {
		lvl = 89
	}
	if lvl < 0 {
		lvl = 0
	}
	rb := char.LCReactBonus(atk)
	brb := char.LCBaseReactBonus(atk)
	return 1.8 * (reactionLvlBase[lvl] * (1 + brb)) * (1 + ((6 * em) / (2000 + em)) + rb*(1+char.Stat(attributes.CD))) * (1 + char.ElevationBonus(atk))
}

var reactionLvlBase = []float64{
	17.1656055450439,
	18.5350475311279,
	19.9048538208007,
	21.27490234375,
	22.6453990936279,
	24.6496124267578,
	26.6406421661376,
	28.8685874938964,
	31.3676795959472,
	34.1433448791503,
	37.201000213623,
	40.6599998474121,
	44.4466667175292,
	48.5635185241699,
	53.7484817504882,
	59.0818977355957,
	64.4200439453125,
	69.7244567871093,
	75.1231384277343,
	80.5847778320312,
	86.1120300292968,
	91.703742980957,
	97.24462890625,
	102.812644958496,
	108.409561157226,
	113.201690673828,
	118.102905273437,
	122.979316711425,
	129.727325439453,
	136.292907714843,
	142.670852661132,
	149.029022216796,
	155.4169921875,
	161.825500488281,
	169.106307983398,
	176.518081665039,
	184.07273864746,
	191.709518432617,
	199.556915283203,
	207.382049560546,
	215.398895263671,
	224.165664672851,
	233.502166748046,
	243.35057067871,
	256.063079833984,
	268.543487548828,
	281.526062011718,
	295.013641357421,
	309.067199707031,
	323.601593017578,
	336.757537841796,
	350.530303955078,
	364.482696533203,
	378.619171142578,
	398.600402832031,
	416.398254394531,
	434.386993408203,
	452.951049804687,
	472.606231689453,
	492.884887695312,
	513.568542480468,
	539.103210449218,
	565.510559082031,
	592.538757324218,
	624.443420410156,
	651.470153808593,
	679.496826171875,
	707.794067382812,
	736.671447753906,
	765.640258789062,
	794.773376464843,
	824.677368164062,
	851.157775878906,
	877.742065429687,
	914.229125976562,
	946.746765136718,
	979.411376953125,
	1011.22302246093,
	1044.79174804687,
	1077.44372558593,
	1109.99755859375,
	1142.9765625,
	1176.36950683593,
	1210.18444824218,
	1253.83569335937,
	1288.95275878906,
	1325.48413085937,
	1363.45690917968,
	1405.09741210937,
	1446.853515625,
}
