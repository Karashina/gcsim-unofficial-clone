package glog

// Event はログに記録される1つのイベントを表す
type Event interface {
	LogSource() Source                            // このログイベントの種類（キャラクター、シム、ダメージ等）を返す
	StartFrame() int                              // このイベントが開始したフレームを返す
	Src() int                                     // このイベントをトリガーしたキャラクターのインデックス。キャラクター以外の場合は-1
	WriteBuildMsg(keyAndVal ...interface{}) Event // イベントに追加のキーと値のペアを書き込む
	Write(key string, val interface{}) Event      // イベントに追加のキーと値のペアを書き込む
	SetEnded(f int) Event
}

// Logger は LogEvent を記録する
type Logger interface {
	// NewEvent(msg string, typ Source, srcChar int, keysAndValues ...interface{}) Event
	NewEvent(msg string, typ Source, srcChar int) Event
	NewEventBuildMsg(typ Source, srcChar int, msg ...string) Event
	Dump() ([]byte, error) // ログされた全イベントを追加順のJSON文字列の配列として出力する
}
