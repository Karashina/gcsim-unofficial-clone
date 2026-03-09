package agg

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/aclements/go-moremath/stats"
)

// カウントがss.Countより小さい場合は何も起こらない。
// それ以外の場合、ss.Countがcountと等しくなるまで0がssに追加される。
func PadStreamStatToCount(ss *stats.StreamStats, count uint) {
	if count < ss.Count {
		return
	}
	for range count - ss.Count {
		ss.Add(0)
	}
}

func ToDescriptiveStats(ss *stats.StreamStats) *model.DescriptiveStats {
	sd := ss.StdDev()
	if math.IsNaN(sd) {
		sd = 0
	}

	mean := ss.Mean()
	return &model.DescriptiveStats{
		Min:  &ss.Min,
		Max:  &ss.Max,
		Mean: &mean,
		SD:   &sd,
	}
}

func ToOverviewStats(input *stats.Sample) *model.OverviewStats {
	input.Sorted = false
	input.Sort()

	minval, maxval := input.Bounds()
	std := input.StdDev()
	if math.IsNaN(std) {
		std = 0
	}

	out := model.OverviewStats{
		SD:   &std,
		Min:  &minval,
		Max:  &maxval,
		Mean: Ptr(input.Mean()),
		Q1:   Ptr(input.Quantile(0.25)),
		Q2:   Ptr(input.Quantile(0.5)),
		Q3:   Ptr(input.Quantile(0.75)),
	}

	// Scottの正規参照ルール
	h := (3.49 * std) / (math.Pow(float64(len(input.Xs)), 1.0/3.0))
	if h == 0.0 || maxval == minval {
		hist := make([]uint32, 1)
		hist[0] = uint32(len(input.Xs))
		out.Hist = hist
	} else {
		nbins := int(math.Ceil((maxval - minval) / h))
		hist := NewLinearHist(minval, maxval, nbins)
		for _, x := range input.Xs {
			hist.Add(x)
		}
		low, bins, high := hist.Counts()
		bins[0] += low
		bins[len(bins)-1] += high
		out.Hist = bins
	}

	return &out
}

// go-moremathから取得。protoの型互換性のため再実装が必要
type LinearHist struct {
	min, max  float64
	delta     float64 // 1/bin width (to avoid division in hot path)
	low, high uint32
	bins      []uint32
}

// NewLinearHistは均一サイズのnbins個のビンを持つ空のヒストグラムを返す。
// ビンは[min, max]の範囲にまたがる。
func NewLinearHist(minval, maxval float64, nbins int) *LinearHist {
	delta := float64(nbins) / (maxval - minval)
	return &LinearHist{minval, maxval, delta, 0, 0, make([]uint32, nbins)}
}

func (h *LinearHist) bin(x float64) int {
	return int(h.delta * (x - h.min))
}

func (h *LinearHist) Add(x float64) {
	switch bin := h.bin(x); {
	case bin < 0:
		h.low++
	case bin >= len(h.bins):
		h.high++
	default:
		h.bins[bin]++
	}
}

func (h *LinearHist) Counts() (uint32, []uint32, uint32) {
	return h.low, h.bins, h.high
}

func (h *LinearHist) BinToValue(bin float64) float64 {
	return h.min + bin/h.delta
}

func Ptr[T any](v T) *T {
	return &v
}

// メタデータとダメージ集計のユーティリティ

const FloatEqDelta = 0.00001

// 事前ソートされた値のスライスを受け取り、パーセンタイルのインデックスを返す
func GetPercentileIndexes[T any](a []T) (int, int) {
	l := len(a)
	if l == 1 {
		return 1, 0
	}
	if l%2 == 0 {
		return l / 2, l / 2
	}
	return (l - 1) / 2, (l-1)/2 + 1
}

// 事前ソートされた値のスライスを受け取り、中央値の要素を返す
func Median[T any](a []T) T {
	l := len(a)

	if l == 0 {
		var empty T
		return empty
	}
	// 配列の長さが偶数の場合、中央値はa[l/2]とa[l/2+1]の間
	// 使用された要素が必要なため、a[l/2]で十分近い
	return a[l/2]
}
