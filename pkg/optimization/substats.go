package optimization

import (
	"fmt"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
	"go.uber.org/zap"
)

type SubstatOptimizer struct {
	logger     *zap.SugaredLogger
	optionsMap map[string]float64
	details    *SubstatOptimizerDetails
	verbose    bool
}

func NewSubstatOptimizer(optionsMap map[string]float64, sugarLog *zap.SugaredLogger, verbose bool) *SubstatOptimizer {
	o := SubstatOptimizer{}
	o.optionsMap = optionsMap
	o.logger = sugarLog
	o.verbose = verbose
	return &o
}

// サブステータス最適化の戦略は現時点では非常にシンプル:
// これは完全に最適ではない - コード内の他のコメントを参照
// 1) ユーザーがチーム、武器、聖遺物セット/メインステータス、ローテーションを設定
// 2) それらを前提に、各キャラクターについて、DPS平均/標準偏差を実質的に最大化するERサブステータス値を選択する
// （高ER値にはペナルティが課される）
//   - 戦略は各キャラクターのERサブステータス値に対する単純なグリッドサーチ
//   - ERサブステータス値は探索を容易にするため2刻みで設定
//
// 3) ER値が決まったら、"勾配降下法"（実際にはそうではない）で他のサブステータスを最適化する
func (o *SubstatOptimizer) Run(cfg string, simopt simulator.Options, simcfg *info.ActionList, gcsl ast.Node) {
	simcfg.Settings.Iterations = int(o.optionsMap["sim_iter"])
	// 最適化では統計情報の収集は不要なので無効化する
	simcfg.Settings.CollectStats = []string{""}

	o.details = NewSubstatOptimizerDetails(
		o,
		cfg,
		simopt,
		simcfg,
		gcsl,
		int(o.optionsMap["indiv_liquid_cap"]),
		int(o.optionsMap["total_liquid_substats"]),
		int(o.optionsMap["fixed_substats_count"]),
	)

	o.details.setStatLimits()

	o.details.setInitialSubstats(o.details.fixedSubstatCount)
	o.logger.Info("Starting ER Optimization...")

	for i := range o.details.charProfilesERBaseline {
		o.details.charProfilesCopy[i] = o.details.charProfilesERBaseline[i].Clone()
	}

	// TODO: ERのみ計算する設定を追加すべきかもしれない？
	o.details.optimizeERSubstats()

	o.logger.Info("Calculating optimal DMG substat distribution...")
	debugLogs := o.details.optimizeNonERSubstats()
	for _, debugLog := range debugLogs {
		o.logger.Info(debugLog)
	}

	for i := 0.0; i < o.optionsMap["fine_tune"]; i++ {
		o.logger.Info("Fine tuning optimal ER vs DMG substat distribution...")
		debugLogs = o.details.optimizeERAndDMGSubstats()
		for _, debugLog := range debugLogs {
			o.logger.Info(debugLog)
		}
	}
}

// 最終出力
// 相対的にそれほど時間はかからないので、常に処理を実行する...
func (o *SubstatOptimizer) PrettyPrint(output string, statsFinal *SubstatOptimizerDetails) string {
	charNames := make(map[keys.Char]string)
	o.logger.Info("Final config substat strings:")

	for _, match := range regexpLineCharname.FindAllStringSubmatch(output, -1) {
		charKey := shortcut.CharNameToKey[match[1]]
		charNames[charKey] = match[1]
	}

	for idxChar := range statsFinal.charProfilesInitial {
		finalString := fmt.Sprintf("%v add stats", charNames[statsFinal.charProfilesInitial[idxChar].Base.Key])

		for idxSubstat, value := range statsFinal.substatValues {
			if value <= 0 {
				continue
			}
			value *= statsFinal.charSubstatRarityMod[idxChar]
			if o.optionsMap["show_substat_scalars"] > 0 {
				finalString += fmt.Sprintf(" %v=%.6g*%v", attributes.StatTypeString[idxSubstat], value, float64(statsFinal.fixedSubstatCount+statsFinal.charSubstatFinal[idxChar][idxSubstat]))
			} else {
				finalString += fmt.Sprintf(" %v=%.6g", attributes.StatTypeString[idxSubstat], value*float64(statsFinal.fixedSubstatCount+statsFinal.charSubstatFinal[idxChar][idxSubstat]))
			}
		}

		fmt.Println(finalString + ";")

		output = replaceSimOutputForChar(charNames[statsFinal.charProfilesInitial[idxChar].Base.Key], output, finalString)
	}

	return output
}

func NewSubstatOptimizerDetails(
	optimizer *SubstatOptimizer,
	cfg string,
	simopt simulator.Options,
	simcfg *info.ActionList,
	gcsl ast.Node,
	indivLiquidCap int,
	totalLiquidSubstats int,
	fixedSubstatCount int,
) *SubstatOptimizerDetails {
	s := SubstatOptimizerDetails{}
	s.optimizer = optimizer
	s.cfg = cfg
	s.simopt = simopt
	s.simcfg = simcfg
	s.fixedSubstatCount = fixedSubstatCount
	s.indivSubstatLiquidCap = indivLiquidCap
	s.totalLiquidSubstats = totalLiquidSubstats

	s.artifactSets4Star = []keys.Set{
		keys.ResolutionOfSojourner,
		keys.TinyMiracle,
		keys.Berserker,
		keys.Instructor,
		keys.TheExile,
		keys.DefendersWill,
		keys.BraveHeart,
		keys.MartialArtist,
		keys.Gambler,
		keys.Scholar,
		keys.PrayersForWisdom,
		keys.PrayersForDestiny,
		keys.PrayersForIllumination,
		keys.PrayersToSpringtime,
	}

	s.substatValues = make([]float64, attributes.EndStatType)
	s.mainstatValues = make([]float64, attributes.EndStatType)

	// TODO: この値の設定方法は本当に最善なのか、何か見落としがあるのでは..？
	s.substatValues[attributes.ATKP] = 0.0496
	s.substatValues[attributes.CR] = 0.0331
	s.substatValues[attributes.CD] = 0.0662
	s.substatValues[attributes.EM] = 19.82
	s.substatValues[attributes.ER] = 0.0551
	s.substatValues[attributes.HPP] = 0.0496
	s.substatValues[attributes.DEFP] = 0.062
	s.substatValues[attributes.ATK] = 16.54
	s.substatValues[attributes.DEF] = 19.68
	s.substatValues[attributes.HP] = 253.94

	s.mainstatValues[attributes.HP] = 4780
	s.mainstatValues[attributes.ATK] = 311
	s.mainstatValues[attributes.ATKP] = 0.466
	s.mainstatValues[attributes.CR] = 0.311
	s.mainstatValues[attributes.CD] = 0.622
	s.mainstatValues[attributes.EM] = 186.5
	s.mainstatValues[attributes.ER] = 0.518
	s.mainstatValues[attributes.HPP] = 0.466
	s.mainstatValues[attributes.DEFP] = 0.583
	s.mainstatValues[attributes.PyroP] = 0.466
	s.mainstatValues[attributes.HydroP] = 0.466
	s.mainstatValues[attributes.CryoP] = 0.466
	s.mainstatValues[attributes.ElectroP] = 0.466
	s.mainstatValues[attributes.AnemoP] = 0.466
	s.mainstatValues[attributes.GeoP] = 0.466
	s.mainstatValues[attributes.DendroP] = 0.466
	s.mainstatValues[attributes.PhyP] = 0.583
	s.mainstatValues[attributes.Heal] = 0.359

	s.mainstatTol = 0.005       // 現在のメインステータス許容誤差は0.5%
	s.fourstarMod = 0.746514762 // 5★メインステータスを4★に変換する平均係数

	// [キャラクター][サブステータス数] を保持する最終出力配列
	s.charSubstatFinal = make([][]int, len(simcfg.Characters))
	for i := range simcfg.Characters {
		s.charSubstatFinal[i] = make([]int, attributes.EndStatType)
	}
	s.charMaxExtraERSubs = make([]float64, len(simcfg.Characters))
	s.charSubstatLimits = make([][]int, len(simcfg.Characters))
	s.charSubstatRarityMod = make([]float64, len(simcfg.Characters))
	s.charProfilesInitial = make([]info.CharacterProfile, len(simcfg.Characters))
	s.charTotalLiquidSubstats = make([]int, len(simcfg.Characters))

	// 最適化のため、これらのキャラクターのエネルギー計算で例外処理が必要
	s.charWithFavonius = make([]bool, len(simcfg.Characters))
	// 初期状態を設定するため全キャラクターに最大ERを付与
	s.charProfilesERBaseline = make([]info.CharacterProfile, len(simcfg.Characters))
	s.charProfilesCopy = make([]info.CharacterProfile, len(simcfg.Characters))
	s.gcsl = gcsl

	s.charRelevantSubstats = make([][]attributes.Stat, len(simcfg.Characters))
	for i := range simcfg.Characters {
		// ERは専用のERステップがあるため省略する。
		s.charRelevantSubstats[i] = []attributes.Stat{
			attributes.HPP,
			attributes.HP,
			attributes.DEFP,
			attributes.DEF,
			attributes.ATKP,
			attributes.ATK,
			attributes.CR,
			attributes.CD,
			attributes.EM,
		}
	}

	return &s
}

// ステータスが低すぎる場合は-1、許容範囲内なら0、高すぎる場合は1を返す
func (stats *SubstatOptimizerDetails) isMainStatInTolerance(idxChar, idxStat, fourStarCount, fiveStarCount int) int {
	lower := stats.mainstatValues[idxStat] * (1 - stats.mainstatTol) * (float64(fiveStarCount) + stats.fourstarMod*float64(fourStarCount))
	upper := stats.mainstatValues[idxStat] * (1 + stats.mainstatTol) * (float64(fiveStarCount) + stats.fourstarMod*float64(fourStarCount))
	val := stats.simcfg.Characters[idxChar].Stats[idxStat]
	switch {
	case val < lower:
		return -1
	case val > upper:
		return 1
	default:
		return 0
	}
}

// possibleMainstatCount[i][0] * 4★値 + possibleMainstatCount[i][1] * 5★値 で使用する場合、この配列は昇順になる
var possibleMainstatCount = [][]int{{1, 0}, {0, 1}, {2, 0}, {1, 1}, {0, 2}, {3, 0}, {2, 1}, {1, 2}, {0, 3}}

// メインステータスに基づいてサブステータス数の上限を取得し、4★セットの状態も判定する
// TODO: 時計/冠/杯のステータス要件にセットを適合させ、DMG%やCritやERの重複を防止する
func (stats *SubstatOptimizerDetails) setStatLimits() {
	for i := range stats.simcfg.Characters {
		stats.setStatLimitsPerChar(i)
	}
}

func (stats *SubstatOptimizerDetails) setStatLimitsPerChar(i int) {
	char := &stats.simcfg.Characters[i]
	fourStarCount := 0
	for set, cnt := range char.Sets {
		for _, fourStar := range stats.artifactSets4Star {
			if set == fourStar {
				fourStarCount += cnt
			}
		}
	}

	stats.charSubstatLimits[i] = make([]int, attributes.EndStatType)

	fourStarMainsCount := 0
	fiveStarMainsCount := 0
	for idxStat := range stats.mainstatValues {
		main4, main5 := stats.setStatLimitsPerCharMainStat(i, idxStat, fourStarCount > 0)
		fourStarMainsCount += main4
		fiveStarMainsCount += main5
	}
	if fourStarMainsCount != fourStarCount {
		stats.optimizer.logger.Warn(char.Base.Key, " has ", fourStarMainsCount, "x 4* mainstats but expected ", fourStarCount)
	}
	if fourStarMainsCount+fiveStarMainsCount != 5 {
		stats.optimizer.logger.Warn(char.Base.Key, " has ", fourStarMainsCount+fiveStarMainsCount, "x mainstats but expected 5")
	}

	// TODO: 2を4★ごとのユーザー設定可能な削減値に置き換える
	stats.charTotalLiquidSubstats[i] = max(stats.totalLiquidSubstats-2*fourStarCount, 0)

	// 全体のレアリティ倍率は4★1つにつき0.04ずつ低下する
	stats.charSubstatRarityMod[i] = 1.0 - 0.04*float64(fourStarCount)
}

// 指定されたステータスに対する（4★メイン数, 5★メイン数）を返す。合計が適合しない場合、
// この関数はxとyがキャラクターのステータス値を超えない最大値となる(x,y)を返す
func (stats *SubstatOptimizerDetails) setStatLimitsPerCharMainStat(i, idxStat int, checkFourStars bool) (int, int) {
	stat := stats.mainstatValues[idxStat]
	char := &stats.simcfg.Characters[i]
	if stat == 0 {
		return 0, 0
	}
	if char.Stats[idxStat] == 0 {
		stats.charSubstatLimits[i][idxStat] = stats.indivSubstatLiquidCap
		return 0, 0
	}

	var main4, main5 int

	// count[0] は4★メインの数
	// count[1] は5★メインの数
	for _, count := range possibleMainstatCount {
		if !checkFourStars && count[0] > 0 {
			continue
		}
		inTol := stats.isMainStatInTolerance(i, idxStat, count[0], count[1])
		if inTol == 0 {
			// 現在、サブステータスごとの上限は4★メインに対して調整されていない
			stats.charSubstatLimits[i][idxStat] = stats.indivSubstatLiquidCap - (stats.fixedSubstatCount * (count[0] + count[1]))
			return count[0], count[1]
		}

		// possibleMainstatCount 配列は昇順である。
		// キャラクターのステータスが低すぎるため、ループを早期終了できる
		if inTol < 0 {
			break
		}
		// main4 と main5 がキャラクターのステータス値を超えない最大値となるよう更新する
		main4 = count[0]
		main5 = count[1]
	}

	val := char.Stats[idxStat]
	msgEnd := " is not a valid multiple of 5* mainstats"
	if checkFourStars {
		msgEnd = " is not a valid sum of 4* or 5* mainstats"
	}
	stats.optimizer.logger.Warn(char.Base.Key, " mainstat ", attributes.Stat(idxStat), "=", val, msgEnd)

	stats.charSubstatLimits[i][idxStat] = stats.indivSubstatLiquidCap - (stats.fixedSubstatCount * (main5 + main4))
	return main4, main5
}

// サブステータス数を見やすく出力するヘルパー関数。float配列を受け取る類似関数から流用
func PrettyPrintStatsCounts(statsCounts []int) string {
	var sb strings.Builder
	sb.WriteString("Liquid Substat Counts: ")
	for i, v := range statsCounts {
		if v > 0 {
			sb.WriteString(attributes.StatTypeString[i])
			sb.WriteString(": ")
			sb.WriteString(fmt.Sprintf("%v", v))
			sb.WriteString(" ")
		}
	}
	return strings.Trim(sb.String(), " ")
}
