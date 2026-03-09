package worker

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/eval"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulation"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

type Pool struct {
	respCh  chan stats.Result
	errCh   chan error
	QueueCh chan Job
	StopCh  chan bool
}
type Job struct {
	Cfg     *info.ActionList
	Actions ast.Node
	Seed    int64
}

// New は新しいPoolを作成する。p.QueueChにジョブを送信できる
// p.StopChを閉じるとキュー内のジョブの実行を停止し、作業中の
// ワーカーはレスポンスを返さなくなる
func New(maxWorker int, respCh chan stats.Result, errCh chan error) *Pool {
	p := &Pool{
		respCh:  respCh,
		errCh:   errCh,
		QueueCh: make(chan Job),
		StopCh:  make(chan bool),
	}
	// ワーカーを作成
	for i := 0; i < maxWorker; i++ {
		go p.worker()
	}
	return p
}
func (p *Pool) worker() {
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
			res, err := s.Run()
			if err != nil {
				p.errCh <- err
				break
			}
			p.respCh <- res
		case _, ok := <-p.StopCh:
			if !ok {
				// stopping
				return
			}
		}
	}
}
