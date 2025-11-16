package agg

import (
	"sync"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

type Config struct {
	Name string
	New  NewAggFunc
}

type Aggregator interface {
	Add(result stats.Result)
	// TODO: Merge(other Aggregator) Aggregator for multi-threaded aggregations (optional optimization)
	Flush(result *model.SimulationStatistics)
}

type NewAggFunc func(cfg *info.ActionList) (Aggregator, error)

var (
	mu          sync.Mutex
	aggregators = map[string]Config{}
)

func Register(cfg Config) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := aggregators[cfg.Name]; ok {
		panic("duplicate aggregator registered: " + cfg.Name)
	}
	aggregators[cfg.Name] = cfg
}

func Aggregators() map[string]Config {
	return aggregators
}

