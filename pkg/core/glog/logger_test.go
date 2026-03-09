package glog

import (
	"testing"

	easyjson "github.com/mailru/easyjson"
)

func TestEventWriteKeyOnlyPanic(t *testing.T) {
	e := &LogEvent{
		Msg:       "test",
		Frame:     1,
		Event:     LogCharacterEvent,
		CharIndex: 0,
		Logs:      map[string]interface{}{},
	}
	// 書き込みテスト
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// これは panic するべき
	e.WriteBuildMsg("keyonly")
}

func TestEventWriteNonStringKeyPanic(t *testing.T) {
	e := &LogEvent{
		Msg:       "test",
		Frame:     1,
		Event:     LogCharacterEvent,
		CharIndex: 0,
		Logs:      map[string]interface{}{},
	}
	// 書き込みテスト
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// これは panic するべき
	e.WriteBuildMsg(1)
}

func TestEventWriteKeyVal(t *testing.T) {
	e := &LogEvent{
		Msg:       "test",
		Frame:     1,
		Event:     LogCharacterEvent,
		CharIndex: 0,
		Logs:      map[string]interface{}{},
		Ordering:  make(map[string]int),
	}

	// panic せず正常に動作するべき
	// e.Write("stuff", 1, "goes", true, "here", "two")
	e.Write("stuff", 1).
		Write("goes", true).
		Write("here", "two")
}

func BenchmarkEasyJSONSerialization(b *testing.B) {
	// 90秒間で1フレームあたり約2行のデバッグを生成
	// 各行は約10フィールド
	// つまり10800イベント
	count := 10800
	var testdata EventArr
	testdata = make([]*LogEvent, 0, count)
	for i := 0; i < count; i++ {
		e := &LogEvent{
			Msg:       "test",
			Frame:     1,
			Event:     LogCharacterEvent,
			CharIndex: 0,
			Logs:      map[string]interface{}{},
			Ordering:  make(map[string]int),
		}
		e.Write("a", 1).
			Write("b", true).
			Write("c", "stuff").
			Write("e", 123).
			Write("f", "boo").
			Write("g", 111)
		testdata = append(testdata, e)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		easyjson.Marshal(testdata)
	}
}

type testChain struct{}

func (t *testChain) Write(key, val interface{}) *testChain { return t }

type testVariadic struct{}

func (t *testVariadic) Write(kv ...interface{}) {}

func BenchmarkChainCalls(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// 10個のkvペアを書き込み
		var x testChain
		x.Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val").
			Write("key", "val")
	}
}

func BenchmarkChainVariadic(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// 10個のkvペアを書き込み
		var x testVariadic
		x.Write(
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
			"key", "val",
		)
	}
}
