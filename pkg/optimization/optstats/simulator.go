package optstats

import (
	"context"
	"math/rand"
	"slices"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/agg"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulator"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

// パース済み設定とカスタムスタットコレクター、アグリゲーターを使用してシミュレーションを実行する
// TODO: cfg文字列はActionList内に含めるべき
// TODO: 無限ループを避けるためにコンテキストを追加する必要がある
func RunWithConfigCustomStats[T any](ctx context.Context, cfg string, simcfg *info.ActionList, gcsl ast.Node, opts simulator.Options, seed int64, cstat NewStatsFuncCustomStats[T], cagg func(T)) (*model.SimulationResult, error) {
	// アグリゲーターを初期化
	var aggregators []agg.Aggregator
	for _, aggregator := range agg.Aggregators() {
		enabled := simcfg.Settings.CollectStats
		if len(enabled) > 0 && !slices.Contains(enabled, aggregator.Name) {
			continue
		}
		a, err := aggregator.New(simcfg)
		if err != nil {
			return &model.SimulationResult{}, err
		}
		aggregators = append(aggregators, a)
	}

	// プールをセットアップ
	respCh := make(chan stats.Result)
	errCh := make(chan error)
	customCh := make(chan T)
	pool := WorkerNewWithCustomStats(simcfg.Settings.NumberOfWorkers, respCh, errCh, customCh)
	pool.StopCh = make(chan bool)

	// キューに入れた合計がイテレーション数未満である限りジョブを入れるgoルーチンを起動
	// キューが一杯になるとブロックする
	go func() {
		src := rand.NewSource(seed)
		// 全てのシードを作成
		wip := 0
		for wip < simcfg.Settings.Iterations {
			pool.QueueCh <- JobCustomStats[T]{
				Cfg:     simcfg.Copy(),
				Actions: gcsl.Copy(),
				Seed:    src.Int63(),
				Cstat:   cstat,
			}
			wip++
		}
	}()

	defer close(pool.StopCh)

	// respChを読み取り、wip == イテレーション数になるまで新しいジョブをキューに入れる
	count := 0
	for count < simcfg.Settings.Iterations {
		select {
		case result := <-customCh:
			cagg(result)
		case result := <-respCh:
			for _, a := range aggregators {
				a.Add(result)
			}
			count += 1
		case err := <-errCh:
			// エラー発生
			return &model.SimulationResult{}, err
		case <-ctx.Done():
			return &model.SimulationResult{}, ctx.Err()
		}
	}

	result, err := simulator.GenerateResult(cfg, simcfg)
	if err != nil {
		return result, err
	}

	// 最終アグリゲーション結果を生成
	stats := &model.SimulationStatistics{}
	for _, a := range aggregators {
		a.Flush(stats)
	}
	result.Statistics = stats

	return result, nil
}
