package optimization

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/optimization/optstats"
)

const FavCritRateBias = 8

// 最小ERから開始
// ERサブステータス割り当てを防ぐための推奨コードが機能することを保証する:
//
//	if .char.burst.ready && .char.energy == .char.energymax {
//	  char burst;
//	}
func (stats *SubstatOptimizerDetails) calculateERBaseline() {
	for i := range stats.charProfilesInitial {
		stats.charProfilesERBaseline[i] = stats.charProfilesInitial[i].Clone()
		// 雷電将軍は爆発のメカニクス上、特別な例外処理が必要
		// TODO: 全雷電将軍のER状態を再帰的にチェックする高コストな解決策なしには、より良い方法はないと思われる
		// 実際には高ERサブの雷電将軍は常に非最適なので、初期スタックを低めに設定する
		erSubs := 0
		if stats.charProfilesInitial[i].Base.Key == keys.Raiden {
			erSubs = 4
		}
		stats.charSubstatFinal[i][attributes.ER] = erSubs

		stats.charProfilesERBaseline[i].Stats[attributes.ER] += float64(erSubs) * stats.substatValues[attributes.ER] * stats.charSubstatRarityMod[i]

		if strings.Contains(stats.charProfilesInitial[i].Weapon.Name, "favonius") {
			stats.calculateERBaselineHandleFav(i)
		}
	}
}

// 西風武器の現在の戦略は、そのキャラクターの会心値を最適ER計算のために少し余分にブーストすること
// その後のサブステータス最適化ステップで、高い会心が重要な場合に自然と大きなDPS向上が見られるはず
// TODO: 西風用のより良い特殊ケースが必要？
func (stats *SubstatOptimizerDetails) calculateERBaselineHandleFav(i int) {
	stats.charProfilesERBaseline[i].Stats[attributes.CR] += FavCritRateBias * stats.substatValues[attributes.CR] * stats.charSubstatRarityMod[i]
	stats.charWithFavonius[i] = true
}

// 各キャラクターの最適ERカットオフを検索
// ignore_burst_energyモードを使用して、各キャラクターが75%の確率で
// 複数ローテーションを成功させるために必要なER量を判定する。
// TODO: ユーザーが最適化に使用するパーセンタイルを設定できるオプションを追加
func (stats *SubstatOptimizerDetails) optimizeERSubstats() {
	stats.simcfg.Settings.Iterations = 350

	// 現時点では雷電将軍を無視する。通常、電池のためだけに最大ERサブを持つことはない。スケーリングがそれほど強くない
	// 最小サブ(0.1102 ER)から最大サブ(0.6612 ER)で、ローテーションごとに4のフラットエネルギーが増える程度
	// 初期4液体なので+/-2フラットエネルギー
	stats.findOptimalERforChars()

	// 以前に見つけたER値で固定し、他の全サブステータスを最適化
	stats.optimizer.logger.Info("Initial Calculated ER Liquid Substats by character:")
	output := ""
	for i := range stats.charProfilesInitial {
		output +=
			fmt.Sprintf("%v: %.4g, ",
				stats.charProfilesInitial[i].Base.Key.String(),
				float64(stats.charSubstatFinal[i][attributes.ER])*stats.substatValues[attributes.ER]*stats.charSubstatRarityMod[i],
			)
	}
	stats.optimizer.logger.Info(output)
}

func (stats *SubstatOptimizerDetails) findOptimalERforChars() {
	stats.simcfg.Settings.IgnoreBurstEnergy = true
	// キャラクターは最小ERから開始
	stats.simcfg.Characters = stats.charProfilesERBaseline

	seed := time.Now().UnixNano()
	a := optstats.NewEnergyAggBuffer(stats.simcfg)
	_, err := optstats.RunWithConfigCustomStats(context.TODO(), stats.cfg, stats.simcfg, stats.gcsl, stats.simopt, seed, optstats.OptimizerERStat, a.Add)
	if err != nil {
		stats.optimizer.logger.Fatal(err.Error())
	}
	a.Flush()
	for idxChar := range stats.charProfilesERBaseline {
		// erDiffは必要なER量
		erLen := len(a.AdditionalErNeeded[idxChar])
		if stats.optimizer.verbose {
			hist := fmtHist(a.ErNeeded[idxChar], float64(int(a.ErNeeded[idxChar][0]*10))/10.0, 0.05)
			stats.optimizer.logger.Infof("%v: ER Needed Distribution", stats.charProfilesInitial[idxChar].Base.Key.Pretty())
			for _, val := range hist {
				stats.optimizer.logger.Infoln(val)
			}
		}
		erDiff := percentile(a.AdditionalErNeeded[idxChar], 0.8)

		erSubVal := stats.substatValues[attributes.ER] * stats.charSubstatRarityMod[idxChar]

		// 最も近い整数のERサブ数を検索
		// TODO: 切り上げと四捨五入どちらが良い？バイアス付きの四捨五入かも？
		erSubs := int(math.Round(erDiff / erSubVal))

		// 雷電将軍は0から開始しないため、その分を差し引く必要がある
		erSubs = clamp[int](0, erSubs, stats.charSubstatLimits[idxChar][attributes.ER]-stats.charSubstatFinal[idxChar][attributes.ER])
		stats.charMaxExtraERSubs[idxChar] = math.Ceil(a.AdditionalErNeeded[idxChar][erLen-1]/erSubVal) - float64(stats.charSubstatFinal[idxChar][attributes.ER])
		stats.charProfilesCopy[idxChar] = stats.charProfilesERBaseline[idxChar].Clone()
		stats.charSubstatFinal[idxChar][attributes.ER] += erSubs
		stats.charProfilesCopy[idxChar].Stats[attributes.ER] += float64(erSubs) * erSubVal
	}
	stats.simcfg.Settings.IgnoreBurstEnergy = false
}
