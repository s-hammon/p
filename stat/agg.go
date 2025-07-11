package stat

import "slices"

type real interface {
	int8 | int16 | int32 | int64 | float32 | float64
}

func Mean[T real](a []T) T {
	var sum T
	if len(a) == 0 {
		return sum
	}
	for _, num := range a {
		sum += num
	}
	return sum / T(len(a))
}

func Median[T real](a []T) T {
	var median T
	n := len(a)
	if n == 0 {
		return median
	}
	sorted := slices.Clone(a)
	slices.Sort(sorted)

	mid := n / 2
	if n%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}
