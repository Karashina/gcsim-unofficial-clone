package optstats

import "github.com/Karashina/gcsim-unofficial-clone/pkg/core"

type CollectorCustomStats[T any] interface {
	Flush(core *core.Core) T
}

type NewStatsFuncCustomStats[T any] func(core *core.Core) (CollectorCustomStats[T], error)

