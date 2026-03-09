package optimization

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

// 有限差分法を使用して、初期状態でのキャラクターごと・サブステータスごとの「勾配」を計算
func (stats *SubstatOptimizerDetails) optimizeERAndDMGSubstats() []string {
	var (
		opDebug   []string
		charDebug []string
	)

	stats.simcfg.Characters = stats.charProfilesCopy

	for idxChar := range stats.charProfilesCopy {
		charDebug = stats.optimizeERAndDMGSubstatsForChar(idxChar, stats.charProfilesCopy[idxChar])
		opDebug = append(opDebug, charDebug...)
	}
	return opDebug
}

// この関数は全てのサブステータスが割り当て済みであることを前提とする。ERのサブを得るごとにダメージのサブを1つ失う
// 1つのDMGサブのダメージ損失と1つのERサブのダメージ利得を比較する
// 全体的に利得がある場合、そのDMGサブを解除して1つのERサブを割り当てる
// ERサブを割り当てられなくなるか、DMG損失が利得を上回るまで繰り返す
// ERのダメージ利得はノイズの影響を受けやすいため、より多くのイテレーションが必要
//
// また、ERサブを失いDMGサブを得ることが全体的に利得かどうかも確認する。これは初期ERヒューリスティックが
// if .char.burst.ready { char burst; } 行が推奨通りに変更されていない場合に失敗するケースをカバーする。
func (stats *SubstatOptimizerDetails) optimizeERAndDMGSubstatsForChar(
	idxChar int,
	char info.CharacterProfile,
) []string {
	var opDebug []string
	opDebug = append(opDebug, fmt.Sprintf("%v", char.Base.Key))

	relevantSubstats := stats.charRelevantSubstats[idxChar]

	totalSubs := stats.getCharSubstatTotal(idxChar)
	if totalSubs != stats.charTotalLiquidSubstats[idxChar] {
		opDebug = append(opDebug, fmt.Sprint("Character has", totalSubs, "total liquid subs allocated but expected", stats.charTotalLiquidSubstats[idxChar]))
	}

	addedEr := false
	// ERサブの追加がダメージを増やすか確認
	for stats.charMaxExtraERSubs[idxChar] > 0.0 && stats.charSubstatFinal[idxChar][attributes.ER] < stats.charSubstatLimits[idxChar][attributes.ER] {
		origIter := stats.simcfg.Settings.Iterations
		stats.simcfg.Settings.IgnoreBurstEnergy = true
		stats.simcfg.Settings.Iterations = 25
		substatGradients := stats.calculateSubstatGradientsForChar(idxChar, relevantSubstats, -1)
		stats.simcfg.Settings.IgnoreBurstEnergy = false
		stats.simcfg.Settings.Iterations = 200
		erGainGradient := stats.calculateSubstatGradientsForChar(idxChar, []attributes.Stat{attributes.ER}, 1)
		stats.simcfg.Settings.Iterations = origIter
		lowestLoss := -999999999999.0
		lowestSub := attributes.NoStat
		for idxSubstat, gradient := range substatGradients {
			substat := relevantSubstats[idxSubstat]
			if stats.charSubstatFinal[idxChar][substat] > 0 && gradient > lowestLoss {
				lowestLoss = gradient
				lowestSub = substat
			}
		}

		// 全体的なダメージ利得0以下の場合、終了
		if erGainGradient[0]+lowestLoss <= 0 || lowestSub == attributes.NoStat {
			break
		}
		addedEr = true

		stats.charSubstatFinal[idxChar][lowestSub] -= 1
		stats.charProfilesCopy[idxChar].Stats[lowestSub] -= float64(1) * stats.substatValues[lowestSub] * stats.charSubstatRarityMod[idxChar]

		stats.charSubstatFinal[idxChar][attributes.ER] += 1
		stats.charProfilesCopy[idxChar].Stats[attributes.ER] += float64(1) * stats.substatValues[attributes.ER] * stats.charSubstatRarityMod[idxChar]
		stats.charMaxExtraERSubs[idxChar] -= 1
	}

	// ERサブを減らすことでダメージが増えるか確認
	// より良い標準偏差のためにERを多めに持つことを優先するため、
	// ここではイテレーション数を少なくし、より高い閾値を使用する。
	// これはユーザーが.char.burst.readyを意図したローテーション変更に適用しないケースをカバーするが、
	// 局所的最小値に引っかかる可能性がある。
	// TODO: 局所的最小値に対する調整方法は？
	for !addedEr && stats.charSubstatFinal[idxChar][attributes.ER] > 0 {
		origIter := stats.simcfg.Settings.Iterations
		stats.simcfg.Settings.IgnoreBurstEnergy = true
		stats.simcfg.Settings.Iterations = 25
		substatGradients := stats.calculateSubstatGradientsForChar(idxChar, relevantSubstats, 1)
		stats.simcfg.Settings.IgnoreBurstEnergy = false
		stats.simcfg.Settings.Iterations = 100
		erGainGradient := stats.calculateSubstatGradientsForChar(idxChar, []attributes.Stat{attributes.ER}, -1)
		stats.simcfg.Settings.Iterations = origIter
		largestGain := -999999999999.0
		largestSub := attributes.NoStat
		for idxSubstat, gradient := range substatGradients {
			substat := relevantSubstats[idxSubstat]
			if stats.charSubstatFinal[idxChar][substat] < stats.charSubstatLimits[idxChar][substat] && gradient > largestGain {
				largestGain = gradient
				largestSub = substat
			}
		}

		// 全体的なダメージ利得が100以下の場合、終了
		if erGainGradient[0]+largestGain <= 100 || largestSub == attributes.NoStat {
			break
		}

		stats.charSubstatFinal[idxChar][largestSub] += 1
		stats.charProfilesCopy[idxChar].Stats[largestSub] += float64(1) * stats.substatValues[largestSub] * stats.charSubstatRarityMod[idxChar]

		stats.charSubstatFinal[idxChar][attributes.ER] -= 1
		stats.charProfilesCopy[idxChar].Stats[attributes.ER] -= float64(1) * stats.substatValues[attributes.ER] * stats.charSubstatRarityMod[idxChar]
		stats.charMaxExtraERSubs[idxChar] += 1
	}

	opDebug = append(opDebug, "Final "+PrettyPrintStatsCounts(stats.charSubstatFinal[idxChar]))
	return opDebug
}
