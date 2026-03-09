package task

// TODO: delay<=0 の挙動が一貫していない
// TODO: 全タスクを単一ハンドラーに統合することを検討
// 現在のタスク実行順序: (enemy1, enemy2, ...), (char1, char2, ...), (core tasks)
// 置き換えるにはcoreタスクキューがタスク位置の更新をサポートする必要がある。
// また実行順序も考慮が必要。現在 QueueCharTask/QueueEnemyTask でキューされたものは
// 常にcoreタスクキューの全エントリより先に実行される。この順序に依存する実装がある場合、
// 追加の問題が発生する。

import "container/heap"

type minHeap []task

type task struct {
	executeBy int
	f         func()
	id        int
}

type Handler struct {
	f       *int
	tasks   *minHeap
	counter int
}

type Tasker interface {
	Add(f func(), delay int)
}

func New(f *int) *Handler {
	return &Handler{
		f:     f,
		tasks: &minHeap{},
	}
}

func (s *Handler) Run() {
	for s.tasks.Len() > 0 && s.tasks.Peek().executeBy <= *s.f {
		heap.Pop(s.tasks).(task).f()
	}
}

func (s *Handler) Add(f func(), delay int) {
	heap.Push(s.tasks, task{
		executeBy: *s.f + delay,
		f:         f,
		id:        s.counter,
	})
	s.counter += 1
}

// 最小ヒープ関数

func (h minHeap) Len() int {
	return len(h)
}

func (h minHeap) Less(i, j int) bool {
	return h[i].executeBy < h[j].executeBy || (h[i].executeBy == h[j].executeBy && h[i].id < h[j].id)
}

func (h minHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(task))
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h minHeap) Peek() task {
	return h[0]
}
