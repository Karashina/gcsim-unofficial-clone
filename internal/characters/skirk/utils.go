package skirk

import "fmt"

/// RingQueue https://github.com/logdyhq/logdy-core/blob/main/ring/ring.go を基に改変

type RingQueue[T any] struct {
	data   []T  // ジェネリック型Tのコンテナデータ
	isFull bool // キューが満杯か空かを区別するフラグ
	start  int  // 開始インデックス（含む、最初の要素）
	end    int  // 終了インデックス（含まない、最後の要素の次）
}

func NewRingQueue[T any](capacity int64) RingQueue[T] {
	return RingQueue[T]{
		data:   make([]T, capacity),
		isFull: false,
		start:  0,
		end:    0,
	}
}

func (r *RingQueue[T]) String() string {
	return fmt.Sprintf(
		"[RingQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *RingQueue[T]) Push(elem T) error {
	if r.isFull {
		return fmt.Errorf("out of bounds push, container is full")
	}
	r.pushUnchecked(elem)

	return nil
}

func (r *RingQueue[T]) PushOverwrite(elem T) {
	if r.isFull {
		r.Pop()
		r.pushUnchecked(elem)
		return
	}
	r.pushUnchecked(elem)
}

func (r *RingQueue[T]) pushUnchecked(elem T) {
	r.data[r.end] = elem              // 空きスペースに新しい要素を配置
	r.end = (r.end + 1) % len(r.data) // 容量のモジュロで終了位置を前進
	r.isFull = r.end == r.start       // 満杯かどうかチェック
}

func (r *RingQueue[T]) Pop() (T, error) {
	var res T // 型に応じた「ゼロ」要素
	if r.IsEmpty() {
		return res, fmt.Errorf("empty queue")
	}

	res = r.data[r.start]                 // キューの最初の要素をコピー
	r.start = (r.start + 1) % len(r.data) // キューの開始位置を移動
	r.isFull = false                      // 要素を削除しているため、満杯にはならない

	return res, nil
}

func (r *RingQueue[T]) IsEmpty() bool {
	if !r.isFull && r.start == r.end {
		return true
	}
	return false
}

func (r *RingQueue[T]) IsFull() bool {
	return r.isFull
}

func (r *RingQueue[T]) Len() int {
	if r.isFull {
		return len(r.data)
	}
	return (r.end - r.start + len(r.data)) % len(r.data)
}

func (r *RingQueue[T]) Index(ind int) (T, error) {
	var res T // 型に応じた「ゼロ」要素
	if ind >= r.Len() {
		return res, fmt.Errorf("Index out of bound")
	}
	return r.data[(r.start+ind)%len(r.data)], nil
}

func (r *RingQueue[T]) Clear() {
	r.start = r.end
	r.isFull = false
}

func (r *RingQueue[T]) Count(filter func(x T) bool) int {
	count := 0
	for i := range r.Len() {
		val, _ := r.Index(i)
		if filter(val) {
			count++
		}
	}
	return count
}
