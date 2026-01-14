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

// Additional runtime option to optimize substats according to KQM standards
func RunSubstatOptim(simopt simulator.Options, verbose bool, additionalOptions string) {
	// Each optimizer run should not be saving anything out for the GZIP
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

	// Parse and set all special sim options
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

	// Parse config
	cfg, err := simulator.ReadConfig(simopt.ConfigPath)
	if err != nil {
		sugarLog.Error(err)
		os.Exit(1)
	}

	clean, err := removeSubstatLines(cfg)
	if errors.Is(err, errInvalidStats) {
		// Provide detailed diagnostics to help identify which character(s) are missing
		// valid main stat rows (flower HP). This will list character names that
		// have no matching mainstat line (hp=4780 or hp=3571).
		charMatches := regexpLineCharname.FindAllStringSubmatch(cfg, -1)
		mainMatches := regexpLineMainstat.FindAllString(cfg, -1)

		hasMain := make(map[string]bool)
		for _, mm := range mainMatches {
			// Attempt to extract the character name from the main stat line
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

		// Fallback generic message
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

	// Sticks optimized substat string into config and output
	if simopt.ResultSaveToPath != "" {
		output = strings.TrimSpace(output) + "\n"
		// try creating file to write to
		err = os.WriteFile(simopt.ResultSaveToPath, []byte(output), 0o644)
		if err != nil {
			log.Panic(err)
		}
		sugarLog.Infof("Saved to the following location: %v", simopt.ResultSaveToPath)
	}
}

