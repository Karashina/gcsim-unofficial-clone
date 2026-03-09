// Package simulation は1回のシミュレーションを実行するための機能を提供する
package simulation

import (
	"slices"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

type Simulation struct {
	// f    int
	preActionDelay int
	C              *core.Core
	// アクションリスト関連
	cfg           *info.ActionList
	queue         []*action.Eval
	eval          action.Evaluator
	noMoreActions bool
	collectors    []stats.Collector

	// 前回のアクション、その使用フレーム、および他のチェーンアクションの
	// 最早使用可能フレームを追跡
}

/**

Simulation should maintain the following:
- queue (apl vs sl)
- frame count? pass it down to core instead of core maintaining
- random damage events
- energy events
- team: this should be a separate package which handles loading the characters, weapons, artifact sets, resonance etc..

**/

func New(cfg *info.ActionList, eval action.Evaluator, c *core.Core) (*Simulation, error) {
	var err error
	s := &Simulation{}
	s.cfg = cfg
	// fmt.Printf("cfg: %+v\n", cfg)
	s.C = c

	err = SetupTargetsInCore(c, geometry.Point{X: cfg.InitialPlayerPos.X, Y: cfg.InitialPlayerPos.Y}, cfg.InitialPlayerPos.R, cfg.Targets)
	if err != nil {
		return nil, err
	}

	err = SetupCharactersInCore(c, cfg.Characters, cfg.InitialChar)
	if err != nil {
		return nil, err
	}

	SetupResonance(c)

	SetupMisc(c)

	err = s.C.Init()
	if err != nil {
		return nil, err
	}

	// 夜魂はキャラクターのステータス初期化後に必要
	setupNightsoulBurst(c)

	for _, collector := range stats.Collectors() {
		enabled := cfg.Settings.CollectStats
		if len(enabled) > 0 && !slices.Contains(enabled, collector.Name) {
			continue
		}
		stat, err := collector.New(s.C)
		if err != nil {
			return nil, err
		}
		s.collectors = append(s.collectors, stat)
	}

	// デバッグログ出力のために呼び出す
	if s.C.Combat.Debug {
		s.CharacterDetails()
	}

	s.eval = eval

	return s, nil
}
