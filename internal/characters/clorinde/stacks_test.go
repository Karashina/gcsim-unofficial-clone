package clorinde

import "testing"

func TestStackTracker(t *testing.T) {
	var f int
	tq := testqueue{
		frame: &f,
	}

	st := newStackTracker(3, tq.queue, &f)

	// 最大3まで追加
	st.Add(10) // f=10で期限切れ
	if st.Count() != 1 {
		t.Errorf("expecting 1 stack, got %v", st.Count())
	}
	f++
	st.Add(10) // f=11で期限切れ
	if st.Count() != 2 {
		t.Errorf("expecting 2 stack, got %v", st.Count())
	}
	f++
	st.Add(10) // f=12で期限切れ
	if st.Count() != 3 {
		t.Errorf("expecting 3 stack, got %v", st.Count())
	}
	// この追加はf=0のものを置き換えるべき
	f++
	st.Add(10) // f=13で期限切れ
	if st.Count() != 3 {
		t.Errorf("expecting 3 stack, got %v", st.Count())
	}
	if *st.stacks[0] != f {
		t.Errorf("expecting src at idx 0 to be %v, got %v", f, *st.stacks[0])
	}
	f = 12
	tq.clear(f)
	if st.Count() != 1 {
		t.Errorf("expecting 1 stack at f = %v, got %v", f, st.Count())
	}
	f = 14
	tq.clear(f)
	if st.Count() != 0 {
		t.Errorf("expecting 0 stack at f = %v, got %v", f, st.Count())
	}
}

type testqueue struct {
	queued []func()
	delay  []int
	frame  *int
}

func (t *testqueue) queue(f func(), delay int) {
	t.queued = append(t.queued, f)
	t.delay = append(t.delay, *t.frame+delay)
}

func (t *testqueue) clear(f int) {
	n := 0
	for i, v := range t.queued {
		if t.delay[i] <= f {
			v()
			continue
		}
		t.queued[n] = v
		t.delay[n] = t.delay[i]
	}
}
