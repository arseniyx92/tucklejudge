package maths

import "math"

func AddIntSlices(l, r []int) []int {
	var max []int
	var min []int
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]int, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = max[i] + min[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}

func IntSliceCopyOf(original []int, newLength int) []int {
	copy := make([]int, newLength)

	for i := 0; i < int(math.Min(float64(len(original)), float64(newLength))); i++ {
		copy[i] = original[i]
	}

	return copy
}

func AddIntToAll(l []int, r int) []int {
	ret := make([]int, len(l))
	for i := 0; i < len(ret); i++ {
		ret[i] = l[i] + r
	}
	return ret
}
func SubtractIntSlices(l, r []int) []int {
	var max []int
	var min []int
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]int, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] - r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}
func DivideIntSlices(l, r []int) []int {
	var max []int
	var min []int
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]int, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] / r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}
func MulIntSlices(l, r []int) []int {
	var max []int
	var min []int
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]int, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] * r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}

func ProductIntSlice(arr []int) int {
	product := 1
	for i := 0; i < len(arr); i++ {
		product *= arr[i]
	}
	return product
}

func AddFloat64Slices(l, r []float64) []float64 {
	var max []float64
	var min []float64
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]float64, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] + r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}
func AddFloat64ToSlice(l []float64, r float64) []float64 {
	ret := make([]float64, len(l))
	for i := 0; i < len(ret); i++ {
		ret[i] = l[i] + r
	}
	return ret
}
func MulFloat64Slices(l, r []float64) []float64 {
	var max []float64
	var min []float64
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]float64, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] * r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}
func MulFloat64ToSlice(l []float64, r float64) []float64 {
	ret := make([]float64, len(l))
	for i := 0; i < len(ret); i++ {
		ret[i] = l[i] * r
	}
	return ret
}

func DivideFloat64Slices(l, r []float64) []float64 {
	var max []float64
	var min []float64
	if len(l) > len(r) {
		max = l
		min = r
	} else {
		max = r
		min = l
	}

	ret := make([]float64, len(max))
	for i := 0; i < len(min); i++ {
		ret[i] = l[i] / r[i]
	}
	for i := len(min); i < len(max); i++ {
		ret[i] = max[i]
	}
	return ret
}
func DivideFloat64SliceByFloat64(l []float64, r float64) []float64 {
	ret := make([]float64, len(l))
	for i := 0; i < len(ret); i++ {
		ret[i] = l[i] / r
	}
	return ret
}

func ProductFloat64Slice(arr []float64) float64 {
	product := 1.0
	for i := 0; i < len(arr); i++ {
		product *= arr[i]
	}
	return product
}

func SumFloat64Slice(arr []float64) float64 {
	sum := 0.0
	for i := 0; i < len(arr); i++ {
		sum += arr[i]
	}
	return sum
}

func FindMaxIndexFloat64Slice(arr []float64) int {
	highest := math.MaxFloat64 * -1
	highestIndex := -1

	for idx, val := range arr {
		if val > highest {
			highest = val
			highestIndex = idx
		}
	}

	return highestIndex
}

func FindMaxValueFloat64Slice(arr []float64) float64 {
	highest := math.MaxFloat64 * -1
	for _, val := range arr {
		if val > highest {
			highest = val
		}
	}
	return highest
}
