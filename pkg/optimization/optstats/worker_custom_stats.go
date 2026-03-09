package optstats

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/eval"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulation"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

type PoolCustomStats[T any] struct {
	respCh   chan stats.Result
	errCh    chan error
	QueueCh  chan JobCustomStats[T]
	customCh chan T
	StopCh   chan bool
}

type JobCustomStats[T any] struct {
	Cfg     *info.ActionList
	Actions ast.Node
	Seed    int64
	Cstat   NewStatsFuncCustomStats[T]
}

// Newは新しいPoolを作成する。p.QueueChに送信することでジョブを投入できる。
// p.StopChを閉じると、プールはキュー内のジョブの実行を停止し、
// 実行中のワーカーはレスポンスを返さなくなる
func WorkerNewWithCustomStats[T any](maxWorker int, respCh chan stats.Result, errCh chan error, customCh chan T) *PoolCustomStats[T] {
	p := &PoolCustomStats[T]{
		respCh:   respCh,
		errCh:    errCh,
		customCh: customCh,
		QueueCh:  make(chan JobCustomStats[T]),
		StopCh:   make(chan bool),
	}
	// ワーカーを作成
	for i := 0; i < maxWorker; i++ {
		go p.worker()
	}
	return p
}

func (p *PoolCustomStats[T]) worker() {
	for {
		select {
		case job := <-p.QueueCh:
			// fmt.Printf("got job: %s\n", job.Cfg.PrettyPrint())
			c, err := simulation.NewCore(job.Seed, false, job.Cfg)
			if err != nil {
				p.errCh <- err
				break
			}
			eval, err := eval.NewEvaluator(job.Actions, c)
			if err != nil {
				p.errCh <- err
				break
			}
			s, err := simulation.New(job.Cfg, eval, c)
			if err != nil {
				p.errCh <- err
				break
			}
			t, err := job.Cstat(c)
			if err != nil {
				p.errCh <- err
				break
			}
			res, err := s.Run()
			if err != nil {
				p.errCh <- err
				break
			}
			p.customCh <- t.Flush(c)
			p.respCh <- res

		case _, ok := <-p.StopCh:
			if !ok {
				// stopping
				return
			}
		}
	}
}
