package stat

import (
	"math"
	"slices"
)

type real interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Mean[T real](a []T) float64 {
	if len(a) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range a {
		sum += float64(v)
	}
	return sum / float64(len(a))
}

func Median[T real](a []T) float64 {
	n := len(a)
	if n == 0 {
		return 0
	}
	sorted := slices.Clone(a)
	slices.Sort(sorted)

	mid := n / 2
	if n%2 == 1 {
		return float64(sorted[mid])
	}
	return (float64(sorted[mid-1]) + float64(sorted[mid])) / 2.0
}

func Stdev[T real](a []T) float64 {
	n := len(a)
	if n < 2 {
		return 0
	}

	var mean, m2 float64
	for i, v := range a {
		x := float64(v)
		k := float64(i + 1)
		del := x - mean
		mean += del / k
		m2 += del * (x - mean)
	}
	variance := m2 / float64(n-1)
	return math.Sqrt(variance)
}
