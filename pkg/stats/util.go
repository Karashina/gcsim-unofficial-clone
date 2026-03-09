package stats

import "math"

type WeightedStreamStats struct {
	Min, Max float64
	mean     float64
	vM2      float64
	wSum     uint
}

func (s *WeightedStreamStats) Add(x float64, w int) {
	if s.wSum == 0 {
		s.Min, s.Max = x, x
	} else {
		if x < s.Min {
			s.Min = x
		}
		if x > s.Max {
			s.Max = x
		}
	}
	s.wSum += uint(w)

	// オンライン重み付き標本頻度分散（Besselの補正付き）
	// West (1979) "Updating Mean and Variance Estimates: An Improved Method" に基づく
	delta := x - s.mean
	s.mean += delta * (float64(w) / float64(s.wSum))
	s.vM2 += float64(w) * delta * (x - s.mean)
}

func (s *WeightedStreamStats) Mean() float64 {
	return s.mean
}

func (s *WeightedStreamStats) Variance() float64 {
	return s.vM2 / float64(s.wSum-1)
}

func (s *WeightedStreamStats) StdDev() float64 {
	out := math.Sqrt(s.Variance())
	if math.IsNaN(out) {
		return 0
	}
	return out
}
