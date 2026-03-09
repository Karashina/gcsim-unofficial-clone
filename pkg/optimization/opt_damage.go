package optimization

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

// 有限差分法を使用して、初期状態でのキャラクターごと・サブステータスごとの「勾配」を計算
// ignore_burst_energyモードでエネルギーからのノイズを除去し、カスタムダメージコレクターで
// ランダム会心のノイズを除去する。これにより、勾配計算ごとに非常に少ない
// 25イテレーションで実行できる
//
// TODO: ユーザーがイテレーション数を増やせる設定を追加する（流浪の楽章やランダム遅延など
// 本質的なランダム性があるケース向け）
//
// TODO: 標準偏差が非常に高い場合に自動的にイテレーション数を増やす？
func (stats *SubstatOptimizerDetails) optimizeNonERSubstats() []string {
	var (
		opDebug   []string
		charDebug []string
	)
	origIter := stats.simcfg.Settings.Iterations
	stats.simcfg.Settings.IgnoreBurstEnergy = true
	stats.simcfg.Settings.Iterations = 25
	stats.simcfg.Characters = stats.charProfilesCopy

	for idxChar := range stats.charProfilesCopy {
		charDebug = stats.optimizeNonErSubstatsForChar(idxChar, stats.charProfilesCopy[idxChar])
		opDebug = append(opDebug, charDebug...)
	}
	stats.simcfg.Settings.IgnoreBurstEnergy = false
	stats.simcfg.Settings.Iterations = origIter
	return opDebug
}

// この計算は、全ての関連サブステータスを最大割り当てから開始する。
// これにより、攻撃力%/HP%/防御力%を積み上げる局所的最大値に引っかかる可能性を減らす。
// 勾配を使用して、1つ除去した時にダメージ損失が最小となるサブステータスを決定する。
// 合計が液体制限内に収まるまで続ける。
// 初期は計算を高速化するために5または2ずつ削除する。
//
// TODO: ユーザーが削除速度を指定できるようにする？
//
// TODO: 0割り当てからのマルチスタート勾配降下/上昇と比較？
func (stats *SubstatOptimizerDetails) optimizeNonErSubstatsForChar(
	idxChar int,
	char info.CharacterProfile,
) []string {
	var opDebug []string
	opDebug = append(opDebug, fmt.Sprintf("%v", char.Base.Key))

	// 西風武器持ちキャラの会心率をリセット
	if stats.charWithFavonius[idxChar] {
		stats.charProfilesCopy[idxChar].Stats[attributes.CR] -= FavCritRateBias * stats.substatValues[attributes.CR] * stats.charSubstatRarityMod[idxChar]
	}

	var relevantSubstats []attributes.Stat
	relevantSubstats = append(relevantSubstats, stats.charRelevantSubstats[idxChar]...)

	// 全ての関連サブステータスを最大液体から開始
	for _, substat := range relevantSubstats {
		stats.charProfilesCopy[idxChar].Stats[substat] +=
			float64(stats.charSubstatLimits[idxChar][substat]-stats.charSubstatFinal[idxChar][substat]) *
				stats.substatValues[substat] * stats.charSubstatRarityMod[idxChar]
		stats.charSubstatFinal[idxChar][substat] = stats.charSubstatLimits[idxChar][substat]
	}

	totalSubs := stats.getCharSubstatTotal(idxChar)
	stats.optimizer.logger.Debug(char.Base.Key.Pretty())
	stats.optimizer.logger.Debug(PrettyPrintStatsCounts(stats.charSubstatFinal[idxChar]))
	for totalSubs > stats.charTotalLiquidSubstats[idxChar] {
		amount := -1
		switch {
		case totalSubs-stats.charTotalLiquidSubstats[idxChar] >= 15:
			amount = -20 // サブステータス上限に応じて10/8にクランプされる
		case totalSubs-stats.charTotalLiquidSubstats[idxChar] >= 8:
			amount = -5
		case totalSubs-stats.charTotalLiquidSubstats[idxChar] >= 4:
			amount = -2
		}
		substatGradients := stats.calculateSubstatGradientsForChar(idxChar, relevantSubstats, amount)

		// totalSubs-stats.totalLiquidSubstats >= 25 の間、複数の勾配をループする
		// 最初の5～6個のサブステータスはDPSに0の影響しかないため、これが最も正確
		for ok := true; ok; ok = totalSubs-stats.charTotalLiquidSubstats[idxChar] >= 25 {
			allocDebug := stats.allocateSomeSubstatGradientsForChar(idxChar, char, substatGradients, relevantSubstats, amount)
			totalSubs = stats.getCharSubstatTotal(idxChar)
			opDebug = append(opDebug, allocDebug...)

			// 最小値のサブステータスをフィルター
			newRelevantSubstats := []attributes.Stat{}
			newSubstatGrad := []float64{}
			removedGrad := -100000000.0
			for idxSub, substat := range relevantSubstats {
				if stats.charSubstatFinal[idxChar][substat] > 0 {
					newRelevantSubstats = append(newRelevantSubstats, substat)
					newSubstatGrad = append(newSubstatGrad, substatGradients[idxSub])
				} else {
					removedGrad = max(removedGrad, substatGradients[idxSub])
				}
			}
			// 削除されたサブステータスの勾配が非常に小さい場合のみcharRelevantSubstatsを更新
			// これは後でopt_allstatsで使用される
			if stats.getCharSubstatTotal(idxChar)-stats.charTotalLiquidSubstats[idxChar] >= 15 ||
				removedGrad >= -100 {
				stats.charRelevantSubstats[idxChar] = nil
				stats.charRelevantSubstats[idxChar] = append(stats.charRelevantSubstats[idxChar], newRelevantSubstats...)
			}
			relevantSubstats = newRelevantSubstats
			substatGradients = newSubstatGrad
		}
		stats.optimizer.logger.Debug(PrettyPrintStatsCounts(stats.charSubstatFinal[idxChar]))
	}
	opDebug = append(opDebug, PrettyPrintStatsCounts(stats.charSubstatFinal[idxChar]))
	stats.optimizer.logger.Debug(char.Base.Key, " has relevant substats:", stats.charRelevantSubstats[idxChar])
	return opDebug
}
