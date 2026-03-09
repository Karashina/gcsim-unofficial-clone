// package queue provide a universal way of handling queuing and executing tasks
package queue

// TODO: このタスクキュー全体を置き換え、coreのタスクキューを代わりに使用すべき
// 置き換えるには、coreのタスクキューがキュー内のタスクの位置を更新する機能をサポートする必要がある
// （タスクの位置更新）
// また、順序も考慮する必要がある。現在、QueueCharTaskでキューに入れられたものは全て
// coreタスクキューの全エントリの前に実行される。この順序に依存する実装がある場合、
// 追加の問題が発生する。
type Task struct {
	F     func()
	Delay int
}

func Add(slice *[]Task, f func(), delay int) {
	if delay == 0 {
		f()
		return
	}
	*slice = append(*slice, Task{
		F:     f,
		Delay: delay,
	})
}

func Run(slice *[]Task, currentTime int) {
	n := 0
	for i := 0; i < len(*slice); i++ {
		if (*slice)[i].Delay <= currentTime {
			(*slice)[i].F()
		} else {
			(*slice)[n] = (*slice)[i]
			n++
		}
	}
	*slice = (*slice)[:n]
}
