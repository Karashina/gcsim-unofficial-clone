package simulator

import (
	"encoding/json"
	"strconv"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/eval"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulation"
	"google.golang.org/protobuf/types/known/structpb"
)

// GenerateSampleWithSeedは指定されたシードでデバッグ有効のシミュレーションを1回実行して出力する
// デバッグログを出力。最小/最大実行のデバッグ生成に使用
func GenerateSampleWithSeed(cfg string, seed uint64, opts Options) (*model.Sample, error) {
	simcfg, gcsl, err := Parse(cfg)
	if err != nil {
		return &model.Sample{}, err
	}

	c, err := simulation.NewCore(int64(seed), true, simcfg)
	if err != nil {
		return &model.Sample{}, err
	}
	eval, err := eval.NewEvaluator(gcsl, c)
	if err != nil {
		return nil, err
	}
	// 新しいシミュレーションを作成して実行
	s, err := simulation.New(simcfg, eval, c)
	if err != nil {
		return &model.Sample{}, err
	}
	_, err = s.Run()
	if err != nil {
		return &model.Sample{}, err
	}

	// ログをキャプチャ
	logs, err := c.Log.Dump()
	if err != nil {
		return &model.Sample{}, err
	}

	// TODO: Log.Dump()はデータをマーシャルすべきでない。JSONの中にJSON文字列を埋め込むのは良くない
	var events []map[string]interface{}
	if err := json.Unmarshal(logs, &events); err != nil {
		return &model.Sample{}, err
	}

	chars, err := GenerateCharacterDetails(simcfg)
	if err != nil {
		return &model.Sample{}, err
	}

	sample := &model.Sample{
		Config:           cfg,
		InitialCharacter: simcfg.InitialChar.String(),
		CharacterDetails: chars,
		Seed:             strconv.FormatUint(seed, 10),
		TargetDetails:    make([]*model.Enemy, len(simcfg.Targets)),
		Logs:             make([]*structpb.Struct, len(events)),
	}

	sample.TargetDetails = make([]*model.Enemy, len(simcfg.Targets))
	for i := range simcfg.Targets {
		target := &simcfg.Targets[i]
		resist := make(map[string]float64)
		for k, v := range target.Resist {
			resist[k.String()] = v
		}

		sample.TargetDetails[i] = &model.Enemy{
			Level:  int32(target.Level),
			HP:     target.HP,
			Resist: resist,
			Pos: &model.Coord{
				X: target.Pos.X,
				Y: target.Pos.Y,
				R: target.Pos.R,
			},
			ParticleDropThreshold: target.ParticleDropThreshold,
			ParticleDropCount:     target.ParticleDropCount,
			ParticleElement:       target.ParticleElement.String(),
			Name:                  target.MonsterName,
			Modified:              target.Modified,
		}
	}

	for i, event := range events {
		es, err := structpb.NewStruct(event)
		if err != nil {
			return &model.Sample{}, err
		}
		sample.Logs[i] = es
	}
	return sample, err
}

// GenerateRawDebugWithSeedは指定されたシードでデバッグ有効のシミュレーションを1回実行する
// 生のデバッグログバイト列（JSON配列）を返し、呼び出し側がNDJSONに変換可能。
func GenerateRawDebugWithSeed(cfg string, seed uint64, opts Options) ([]byte, error) {
	simcfg, gcsl, err := Parse(cfg)
	if err != nil {
		return nil, err
	}

	c, err := simulation.NewCore(int64(seed), true, simcfg)
	if err != nil {
		return nil, err
	}
	eval, err := eval.NewEvaluator(gcsl, c)
	if err != nil {
		return nil, err
	}
	// 新しいシミュレーションを作成して実行
	s, err := simulation.New(simcfg, eval, c)
	if err != nil {
		return nil, err
	}
	_, err = s.Run()
	if err != nil {
		return nil, err
	}

	// 生のJSON配列としてログをキャプチャ
	logs, err := c.Log.Dump()
	if err != nil {
		return nil, err
	}
	return logs, nil
}
