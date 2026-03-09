package glog

import (
	"log"
	"strings"

	easyjson "github.com/mailru/easyjson"
)

// Debugw用
// Warnw用

//nolint:staticcheck // staticcheckはnocopyがjsonではなくeasyjsonからのものであると認識できない: https://github.com/dominikh/go-tools/issues/836
type keyVal struct {
	Key string      `json:"key,nocopy"`
	Val interface{} `json:"val"`
}

//nolint:staticcheck // staticcheckはnocopyがjsonではなくeasyjsonからのものであると認識できない: https://github.com/dominikh/go-tools/issues/836
//easyjson:json
type LogEvent struct {
	Event     Source                 `json:"event"`
	Frame     int                    `json:"frame"`
	Ended     int                    `json:"ended"`
	CharIndex int                    `json:"char_index"`
	Msg       string                 `json:"msg,nocopy"`
	Logs      map[string]interface{} `json:"logs"`
	Ordering  map[string]int         `json:"ordering"`
	counter   int
}

//easyjson:json
type EventArr []*LogEvent

func (e *LogEvent) Write(key string, value interface{}) Event {
	e.Logs[key] = value
	e.Ordering[key] = e.counter
	e.counter++

	return e
}

func (e *LogEvent) WriteBuildMsg(keysAndValues ...interface{}) Event {
	// 偶数であるべき
	var key string
	var ok bool
	for i := 0; i < len(keysAndValues); i++ {
		key, ok = keysAndValues[i].(string)
		if !ok {
			log.Panicf("invalid key %v, expected type to be string", keysAndValues[i].(string))
		}
		// 対応する値があることを確認
		i++
		if i == len(keysAndValues) {
			log.Panicf("expected an associated value after key %v, got nothing", key)
		}
		// e.Logs = append(e.Logs, keyVal{
		// 	Key: key,
		// 	Val: keysAndValues[i],
		// })
		e.Logs[key] = keysAndValues[i]
		e.Ordering[key] = e.counter
		e.counter++
	}
	return e
}

func (e *LogEvent) SetEnded(end int) Event {
	e.Ended = end
	return e
}

func (e *LogEvent) LogSource() Source { return e.Event }
func (e *LogEvent) StartFrame() int   { return e.Frame }
func (e *LogEvent) Src() int          { return e.CharIndex }

type Ctrl struct {
	// 発生順序を追跡するために配列に保存
	// events []*Event
	events map[int]*LogEvent
	count  int
	f      *int
}

func New(f *int, size int) Logger {
	ctrl := &Ctrl{
		events: make(map[int]*LogEvent),
		f:      f,
	}
	return ctrl
}

func (c *Ctrl) Dump() ([]byte, error) {
	r := make(EventArr, 0, c.count)
	for i := 0; i < c.count; i++ {
		v, ok := c.events[i]
		if ok {
			r = append(r, v)
		}
	}
	return easyjson.Marshal(r)
}

func (c *Ctrl) NewEventBuildMsg(typ Source, srcChar int, msg ...string) Event {
	if len(msg) == 0 {
		panic("no msg provided")
	}
	var sb strings.Builder
	for _, v := range msg {
		sb.WriteString(v)
	}
	return c.NewEvent(sb.String(), typ, srcChar)
}

func (c *Ctrl) NewEvent(msg string, typ Source, srcChar int) Event {
	e := &LogEvent{
		Msg:       msg,
		Frame:     *c.f,
		Ended:     *c.f,
		Event:     typ,
		CharIndex: srcChar,
		Logs:      make(map[string]interface{}), // デフォルトに+5、追加のキーが必要になる場合に備えて
		Ordering:  make(map[string]int),
	}
	// c.events = append(c.events, e)
	c.events[c.count] = e
	c.count++
	return e
}
