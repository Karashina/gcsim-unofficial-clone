package simulator

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/agg"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/parser"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/result"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/worker"
)

// Options はシムの実行設定を定義する（デバッグモード等）
type Options struct {
	ResultSaveToPath string // 結果ファイルの保存パス（拡張子除く）。空の場合はファイルに保存しない
	GZIPResult       bool   // 結果ファイルをgzip圧縮するか。ResultSaveToPathが空でない場合のみ有効
	ConfigPath       string // 読み込む設定ファイルのパス
}

var (
	sha1ver   string
	buildTime string
	modified  bool
)

func init() {
	info, _ := debug.ReadBuildInfo()
	for _, bs := range info.Settings {
		if bs.Key == "vcs.revision" {
			sha1ver = bs.Value
		}
		if bs.Key == "vcs.time" {
			buildTime = bs.Value
		}
		if bs.Key == "vcs.modified" {
			bv, _ := strconv.ParseBool(bs.Value)
			modified = bv
		}
	}
}

func Version() string {
	return sha1ver
}

func Parse(cfg string) (*info.ActionList, ast.Node, error) {
	parser := parser.New(cfg)
	simcfg, gcsl, err := parser.Parse()
	if err != nil {
		return &info.ActionList{}, nil, err
	}

	// 他のエラーもチェック
	if len(simcfg.Errors) != 0 {
		fmt.Println("The config has the following errors: ")
		errMsgs := ""
		for _, v := range simcfg.Errors {
			errMsg := fmt.Sprintf("\t%v\n", v)
			fmt.Println(errMsg)
			errMsgs += errMsg
		}
		return &info.ActionList{}, nil, errors.New(errMsgs)
	}

	return simcfg, gcsl, nil
}

// Run はシミュレーションを指定回数実行する
func Run(ctx context.Context, opts Options) (*model.SimulationResult, error) {
	start := time.Now()

	cfg, err := ReadConfig(opts.ConfigPath)
	if err != nil {
		return &model.SimulationResult{}, err
	}

	simcfg, gcsl, err := Parse(cfg)
	if err != nil {
		return &model.SimulationResult{}, err
	}

	return RunWithConfig(ctx, cfg, simcfg, gcsl, opts, start)
}

// RunWithConfig はパース済みの設定でシミュレーションを実行する
// TODO: cfg 文字列は ActionList に含めるべき
// TODO: 無限ループを避けるためにコンテキストを追加する必要がある
func RunWithConfig(ctx context.Context, cfg string, simcfg *info.ActionList, gcsl ast.Node, opts Options, start time.Time) (*model.SimulationResult, error) {
	// アグリゲータを初期化
	var aggregators []agg.Aggregator
	for _, aggregator := range agg.Aggregators() {
		enabled := simcfg.Settings.CollectStats
		if len(enabled) > 0 && !slices.Contains(enabled, aggregator.Name) {
			continue
		}
		a, err := aggregator.New(simcfg)
		if err != nil {
			return &model.SimulationResult{}, err
		}
		aggregators = append(aggregators, a)
	}

	// プールをセットアップ
	respCh := make(chan stats.Result)
	errCh := make(chan error)
	pool := worker.New(simcfg.Settings.NumberOfWorkers, respCh, errCh)
	pool.StopCh = make(chan bool)

	// ジョブをキューに入れるgoroutineを起動。キューが一杯になるとブロックされる
	go func() {
		// 全シードを作成
		wip := 0
		for wip < simcfg.Settings.Iterations {
			pool.QueueCh <- worker.Job{
				Cfg:     simcfg.Copy(),
				Actions: gcsl.Copy(),
				Seed:    CryptoRandSeed(),
			}
			wip++
		}
	}()

	defer close(pool.StopCh)

	// respChを読み取り、全イテレーション完了まで待機
	count := 0
	for count < simcfg.Settings.Iterations {
		select {
		case result := <-respCh:
			for _, a := range aggregators {
				a.Add(result)
			}
			count += 1
		case err := <-errCh:
			// エラーが発生
			return &model.SimulationResult{}, err
		case <-ctx.Done():
			return &model.SimulationResult{}, ctx.Err()
		}
	}

	result, err := GenerateResult(cfg, simcfg)
	if err != nil {
		return result, err
	}

	// 最終アグリゲート結果を生成
	stats := &model.SimulationStatistics{}
	for _, a := range aggregators {
		a.Flush(stats)
	}
	result.Statistics = stats

	return result, nil
}

// GenerateResult はイテレーションに依存しない結果を生成する（イテレーション数によって出力は変わらない）
func GenerateResult(cfg string, simcfg *info.ActionList) (*model.SimulationResult, error) {
	out := &model.SimulationResult{
		// これは常にUIのビューアーアップグレードダイアログと同期する必要がある
		// 結果スキーマが変更された場合のみスキーマを変更する。AGG結果の変更も含む
		// SemVer 仕様:
		//    Major: 新スキーマに後方互換性がない場合に上げ、minorを0にリセット
		//        例 - 重要なカラムの位置変更、大規模リファクタリング
		//    Minor: 新スキーマに後方互換性がある場合に上げる
		//        例 - UIの新グラフ用データ追加。データがなくてもUIは機能する
		// バージョンを上げるとUIが全ての古いシムを古いとフラグする
		SchemaVersion: &model.Version{Major: "4", Minor: "2"}, // UIバージョンと同期すること ui/packages/ui/src/Pages/Viewer/UpgradeDialog.tsx
		SimVersion:    &sha1ver,
		BuildDate:     buildTime,
		Modified:      &modified,
		KeyType:       "NONE",
		SimulatorSettings: &model.SimulatorSettings{
			Duration:          simcfg.Settings.Duration,
			DamageMode:        simcfg.Settings.DamageMode,
			EnableHitlag:      simcfg.Settings.EnableHitlag,
			DefHalt:           simcfg.Settings.DefHalt,
			IgnoreBurstEnergy: simcfg.Settings.IgnoreBurstEnergy,
			NumberOfWorkers:   uint32(simcfg.Settings.NumberOfWorkers),
			Iterations:        uint32(simcfg.Settings.Iterations),
			Delays: &model.Delays{
				Skill:  int32(simcfg.Settings.Delays.Skill),
				Burst:  int32(simcfg.Settings.Delays.Burst),
				Attack: int32(simcfg.Settings.Delays.Attack),
				Charge: int32(simcfg.Settings.Delays.Charge),
				Aim:    int32(simcfg.Settings.Delays.Aim),
				Dash:   int32(simcfg.Settings.Delays.Dash),
				Jump:   int32(simcfg.Settings.Delays.Jump),
				Swap:   int32(simcfg.Settings.Delays.Swap),
			},
		},
		EnergySettings: &model.EnergySettings{
			Active:         simcfg.EnergySettings.Active,
			Once:           simcfg.EnergySettings.Once,
			Start:          int32(simcfg.EnergySettings.Start),
			End:            int32(simcfg.EnergySettings.End),
			Amount:         int32(simcfg.EnergySettings.Amount),
			LastEnergyDrop: int32(simcfg.EnergySettings.LastEnergyDrop),
		},
		Config:           cfg,
		SampleSeed:       strconv.FormatUint(uint64(CryptoRandSeed()), 10),
		InitialCharacter: simcfg.InitialChar.String(),
		TargetDetails:    make([]*model.Enemy, len(simcfg.Targets)),
		PlayerPosition: &model.Coord{
			X: simcfg.InitialPlayerPos.X,
			Y: simcfg.InitialPlayerPos.Y,
			R: simcfg.InitialPlayerPos.R,
		},
	}

	for i := range simcfg.Targets {
		target := &simcfg.Targets[i]
		resist := make(map[string]float64)
		for k, v := range target.Resist {
			resist[k.String()] = v
		}

		out.TargetDetails[i] = &model.Enemy{
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

	if simcfg.Settings.DamageMode {
		out.Mode = model.SimMode_TTK_MODE
	}

	charDetails, err := GenerateCharacterDetails(simcfg)
	if err != nil {
		return out, err
	}
	out.CharacterDetails = charDetails

	for i := range simcfg.Characters {
		if !result.IsCharacterComplete(simcfg.Characters[i].Base.Key) {
			out.IncompleteCharacters = append(out.IncompleteCharacters, simcfg.Characters[i].Base.Key.String())
		}
	}

	return out, nil
}

// CryptoRandSeed は暗号学的乱数を使用してランダムシードを生成する
func CryptoRandSeed() int64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}

var reImport = regexp.MustCompile(`(?m)^import "(.+)"$`)

// ReadConfig は指定パスの設定を読み込む。import文も解決する。
func ReadConfig(fpath string) (string, error) {
	src, err := os.ReadFile(fpath)
	if err != nil {
		return "", err
	}

	// importをチェック
	var data strings.Builder

	rows := strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")
	for _, row := range rows {
		match := reImport.FindStringSubmatch(row)
		if match != nil {
			// importを読み込む
			p := path.Join(path.Dir(fpath), match[1])
			src, err = os.ReadFile(p)
			if err != nil {
				return "", err
			}

			data.Write(src)
		} else {
			data.WriteString(row)
			data.WriteString("\n")
		}
	}

	return data.String(), nil
}
