package optimization

import (
	"context"
	"sort"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/optimization/optstats"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}
type SubstatOptimizerDetails struct {
	artifactSets4Star       []keys.Set
	substatValues           []float64
	mainstatValues          []float64
	mainstatTol             float64
	fourstarMod             float64
	charSubstatFinal        [][]int
	charSubstatLimits       [][]int
	charTotalLiquidSubstats []int
	charSubstatRarityMod    []float64
	charProfilesInitial     []info.CharacterProfile
	charWithFavonius        []bool
	charProfilesERBaseline  []info.CharacterProfile
	charRelevantSubstats    [][]attributes.Stat
	charProfilesCopy        []info.CharacterProfile
	charMaxExtraERSubs      []float64
	simcfg                  *info.ActionList
	gcsl                    ast.Node
	simopt                  simulator.Options
	cfg                     string
	fixedSubstatCount       int
	indivSubstatLiquidCap   int
	totalLiquidSubstats     int
	optimizer               *SubstatOptimizer
}

func (stats *SubstatOptimizerDetails) allocateSomeSubstatGradientsForChar(
	idxChar int,
	_ info.CharacterProfile,
	substatGradient []float64,
	relevantSubstats []attributes.Stat,
	amount int,
) []string {
	var opDebug []string
	sorted := newSlice(substatGradient...)
	sort.Sort(sort.Reverse(sorted))

	for _, idxSubstat := range sorted.idx {
		substat := relevantSubstats[idxSubstat]

		if amount > 0 {
			if stats.charSubstatFinal[idxChar][substat] < stats.charSubstatLimits[idxChar][substat] {
				stats.charSubstatFinal[idxChar][substat] += amount
				stats.charProfilesCopy[idxChar].Stats[substat] += float64(amount) * stats.substatValues[substat] * stats.charSubstatRarityMod[idxChar]
				return opDebug
			}
		}

		if stats.charSubstatFinal[idxChar][substat] > 0 {
			amount = clamp[int](-stats.charSubstatFinal[idxChar][substat], amount, amount)
			stats.charSubstatFinal[idxChar][substat] += amount
			stats.charProfilesCopy[idxChar].Stats[substat] += float64(amount) * stats.substatValues[substat] * stats.charSubstatRarityMod[idxChar]
			return opDebug
		}
	}

	// TODO: 関連するサブステータスの割り当て/解除ができない場合、ランダムな他のサブステータスを割り当て/解除すべき？
	opDebug = append(opDebug, "Couldn't alloc/dealloc anything?????")
	return opDebug
}

func (stats *SubstatOptimizerDetails) calculateSubstatGradientsForChar(
	idxChar int,
	relevantSubstats []attributes.Stat,
	amount int,
) []float64 {
	stats.simcfg.Characters = stats.charProfilesCopy

	seed := time.Now().UnixNano()
	init := optstats.NewDamageAggBuffer(stats.simcfg)
	_, err := optstats.RunWithConfigCustomStats(context.TODO(), stats.cfg, stats.simcfg, stats.gcsl, stats.simopt, seed, optstats.OptimizerDmgStat, init.Add)
	if err != nil {
		stats.optimizer.logger.Fatal(err.Error())
	}
	init.Flush()
	// TODO: 中央値と平均値のどちらがより良い結果を出すかテストする
	initialMean := mean(init.ExpectedDps)
	substatGradients := make([]float64, len(relevantSubstats))
	// サブステータスごとに「勾配」を構築
	for idxSubstat, substat := range relevantSubstats {
		stats.charProfilesCopy[idxChar].Stats[substat] += float64(amount) * stats.substatValues[substat] * stats.charSubstatRarityMod[idxChar]

		stats.simcfg.Characters = stats.charProfilesCopy

		a := optstats.NewDamageAggBuffer(stats.simcfg)
		_, err := optstats.RunWithConfigCustomStats(context.TODO(), stats.cfg, stats.simcfg, stats.gcsl, stats.simopt, seed, optstats.OptimizerDmgStat, a.Add)
		if err != nil {
			stats.optimizer.logger.Fatal(err.Error())
		}
		a.Flush()

		substatGradients[idxSubstat] = mean(a.ExpectedDps) - initialMean
		// 西風武器所持者が西風を安定的に発動するための十分な会心率を得られないケースを修正（代表例: 西風カズハ）
		// 「過剰な」会心率を与える可能性がある（=液体CRサブを最大化、または会心率100%超過）が、大きな問題ではない
		if stats.simcfg.Settings.IgnoreBurstEnergy && stats.charWithFavonius[idxChar] && substat == attributes.CR {
			substatGradients[idxSubstat] += 1000 * float64(amount)
		}
		stats.charProfilesCopy[idxChar].Stats[substat] -= float64(amount) * stats.substatValues[substat] * stats.charSubstatRarityMod[idxChar]
	}
	return substatGradients
}

func (stats *SubstatOptimizerDetails) setInitialSubstats(fixedSubstatCount int) {
	stats.cloneStatsWithFixedAllocations(fixedSubstatCount)
	stats.calculateERBaseline()
}

// 固定割り当て（各サブステータス2個）で初期キャラクター状態を保存するためのコピー
func (stats *SubstatOptimizerDetails) cloneStatsWithFixedAllocations(fixedSubstatCount int) {
	for i := range stats.simcfg.Characters {
		stats.charProfilesInitial[i] = stats.simcfg.Characters[i].Clone()
		for idxStat, stat := range stats.substatValues {
			if stat == 0 {
				continue
			}
			stats.charProfilesInitial[i].Stats[idxStat] += float64(fixedSubstatCount) * stat * stats.charSubstatRarityMod[i]
		}
	}
}
