package optimization

import (
	"errors"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/parser"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
)

// KQM基準に従ってサブステータスを最適化する追加ランタイムオプション
func RunSubstatOptim(simopt simulator.Options, verbose bool, additionalOptions string) {
	// 各オプティマイザー実行はGZIPに何も保存すべきではない
	simopt.GZIPResult = false

	optionsMap := map[string]float64{
		"total_liquid_substats": 20,
		"indiv_liquid_cap":      10,
		"fixed_substats_count":  2,
		"verbose":               0,
		"fine_tune":             1,
		"show_substat_scalars":  1,
	}

	if verbose {
		optionsMap["verbose"] = 1
	}

	// 特殊シミュレーションオプションをパースして設定
	var sugarLog *zap.SugaredLogger
	if additionalOptions != "" {
		optionsMap, err := parseOptimizerCfg(additionalOptions, optionsMap)
		sugarLog = newLogger(optionsMap["verbose"] == 1)
		if err != nil {
			sugarLog.Panic(err.Error())
		}
	} else {
		sugarLog = newLogger(optionsMap["verbose"] == 1)
	}

	// 設定をパース
	cfg, err := simulator.ReadConfig(simopt.ConfigPath)
	if err != nil {
		sugarLog.Error(err)
		os.Exit(1)
	}

	clean, err := removeSubstatLines(cfg)
	if errors.Is(err, errInvalidStats) {
		// どのキャラクターに有効なメインステータス行（花HP）が欠けているか特定するための詳細な診断情報を提供。
		// 対応するメインステータス行（hp=4780またはhp=3571）がないキャラクター名をリスト表示する。
		charMatches := regexpLineCharname.FindAllStringSubmatch(cfg, -1)
		mainMatches := regexpLineMainstat.FindAllString(cfg, -1)

		hasMain := make(map[string]bool)
		for _, mm := range mainMatches {
			// メインステータス行からキャラクター名の抽出を試みる
			sub := regexpLineCharname.FindStringSubmatch(mm)
			if len(sub) > 1 {
				hasMain[sub[1]] = true
			}
		}

		var missing []string
		for _, cm := range charMatches {
			if len(cm) > 1 {
				name := cm[1]
				if !hasMain[name] {
					missing = append(missing, name)
				}
			}
		}

		if len(missing) > 0 {
			sugarLog.Panicf("Error: Could not identify valid main artifact stat rows for the following characters (missing flower HP main stat lines): %v\n5* flowers must have 4780 HP, and 4* flowers must have 3571 HP.", missing)
		}

		// フォールバックの汎用メッセージ
		sugarLog.Panic("Error: Could not identify valid main artifact stat rows for all characters based on flower HP values.\n5* flowers must have 4780 HP, and 4* flowers must have 3571 HP.")
		os.Exit(1)
	}

	if err != nil {
		sugarLog.Warn(err.Error())
	}

	parser := parser.New(clean)
	simcfg, gcsl, err := parser.Parse()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	optimizer := NewSubstatOptimizer(optionsMap, sugarLog, verbose)
	optimizer.Run(cfg, simopt, simcfg, gcsl)
	output := optimizer.PrettyPrint(clean, optimizer.details)

	// 最適化されたサブステータス文字列を設定に挿入して出力
	if simopt.ResultSaveToPath != "" {
		output = strings.TrimSpace(output) + "\n"
		// 書き込み先ファイルの作成を試行
		err = os.WriteFile(simopt.ResultSaveToPath, []byte(output), 0o644)
		if err != nil {
			log.Panic(err)
		}
		sugarLog.Infof("Saved to the following location: %v", simopt.ResultSaveToPath)
	}
}
