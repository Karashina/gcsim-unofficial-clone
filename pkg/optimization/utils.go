package optimization

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ソートされたインデックスも取得できるsort.Sliceの薄いラッパー
type Slice struct {
	slice sort.Float64Slice
	idx   []int
}

func (s Slice) Len() int {
	return len(s.slice)
}

func (s Slice) Less(i, j int) bool {
	return s.slice[i] < s.slice[j]
}

func (s Slice) Swap(i, j int) {
	s.slice.Swap(i, j)
	s.idx[i], s.idx[j] = s.idx[j], s.idx[i]
}

func newSlice(n ...float64) *Slice {
	var nDup []float64
	nDup = append(nDup, n...)

	s := &Slice{
		slice: sort.Float64Slice(nDup),
		idx:   make([]int, len(nDup)),
	}
	for i := range s.idx {
		s.idx[i] = i
	}
	return s
}

func percentile[T comparable](arr []T, percentile float64) T {
	return arr[int(math.Floor(float64(len(arr))*percentile))]
}

func mean(arr []float64) float64 {
	if len(arr) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range arr {
		sum += v
	}

	return sum / float64(len(arr))
}

func clamp[T Ordered](minVal, val, maxVal T) T {
	return max(min(val, maxVal), minVal)
}

func (stats *SubstatOptimizerDetails) getCharSubstatTotal(idxChar int) int {
	sum := 0
	for _, count := range stats.charSubstatFinal[idxChar] {
		sum += count
	}
	return sum
}

func fmtHist(sortedArr []float64, start, binSize float64) []string {
	valPerBlock := 1.0 / 350.0 * 8.0 * 1.0 // 1 iteration = 1/8th of a block

	output := make([]string, 0)
	binMin := make([]float64, 0)
	binMax := make([]float64, 0)
	binCount := make([]float64, 0)

	for i := start; i <= sortedArr[len(sortedArr)-1]; i += binSize {
		output = append(output, "")
		binMin = append(binMin, i)
		binCount = append(binCount, 0)
		binMax = append(binMax, i+binSize)
	}

	currBin := 0
	for _, val := range sortedArr {
		for val > binMax[currBin] {
			currBin++
		}
		binCount[currBin]++
	}

	for i := range binCount {
		binCount[i] /= float64(len(sortedArr))
	}
	// イテレーション数が不足しているビンを先頭と末尾から除去
	for i := range binCount {
		if int(binCount[i]/valPerBlock) == 0 && int(math.Round(math.Mod(binCount[i], valPerBlock)*8)) == 0 {
			continue
		}
		if i != 0 {
			output = output[i:]
			binMin = binMin[i:]
			binMax = binMax[i:]
			binCount = binCount[i:]
		}
		break
	}

	for i := len(binCount) - 1; i >= 0; i-- {
		if int(binCount[i]/valPerBlock) == 0 && int(math.Round(math.Mod(binCount[i], valPerBlock)*8)) == 0 {
			continue
		}
		output = output[:i+1]
		binMin = binMin[:i+1]
		binMax = binMax[:i+1]
		binCount = binCount[:i+1]
		break
	}

	for i := range output {
		// ASCIIブロック要素は8単位のチャンクなので、8分の何が必要かを計算する
		// https://en.wikipedia.org/wiki/Block_Elements
		barChunks := int(math.Round(binCount[i] / valPerBlock))
		bar := strings.Repeat("█", barChunks)

		// 現在のPowerShellデフォルトフォントは部分ブロック要素をサポートしていない
		// Windows Terminalのデフォルトフォント Cascadia Mono はサポートしている
		// TODO: Windows 11をターゲットにしたらこれを復活させ、barChunksのmath.Roundを削除する
		// （W11のデフォルトコンソールはTerminalのため）
		// rem := int(math.Round(math.Mod(binCount[i], valPerBlock) / valPerBlock * 8))
		// if rem > 0 {
		// 	bar += fmt.Sprint(string(rune('█' + 8 - rem)))
		// }
		output[i] = fmt.Sprintf(" %.0f-%.0f  |%s", binMin[i]*100, binMax[i]*100, bar)
	}

	return output
}
