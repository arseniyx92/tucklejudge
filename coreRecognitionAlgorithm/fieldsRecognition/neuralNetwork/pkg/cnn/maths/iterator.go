package maths

import (
	"math"
)

type CoordIteratorInterface interface {
	GetCurrentCoords() []int
	GetCurrentCount() int
	HasNext() bool
	Next() []int
}
type CoordIterator struct {
	startCoords, finalCoords, currentCoords []int
	currentCount                            int
	finalIndex                              int
}

func NewCoordIterator(corner1, corner2 []int) *CoordIterator {
	it := &CoordIterator{
		startCoords:   make([]int, len(corner1)),
		currentCoords: make([]int, len(corner1)),
		finalCoords:   make([]int, len(corner2)),
		finalIndex:    1,
	}

	//Ensure start is at minimum coords within region, end is at maximum
	for i := 0; i < len(it.startCoords); i++ {
		it.startCoords[i] = int(math.Min(float64(corner1[i]), float64(corner2[i])))
		it.currentCoords[i] = it.startCoords[i]
		it.finalCoords[i] = int(math.Max(float64(corner1[i]), float64(corner2[i]))) - 1

		//The total number of integer coordinates within the region (not considering stride)
		it.finalIndex *= (it.finalCoords[i] - it.startCoords[i]) + 2
	}
	//At each iteration, we increment the coordinate before returning it. This means that our starting position has to be 1 stride less than the first coord.
	it.currentCoords[0]--
	return it
}

func (it *CoordIterator) GetCurrentCoords() []int { return it.currentCoords }
func (it *CoordIterator) GetCurrentCount() int    { return it.currentCount }
func (it *CoordIterator) HasNext() bool           { return it.currentCount < it.finalIndex }
func (it *CoordIterator) Next() []int {
	for i := 0; i < len(it.currentCoords); i++ {
		if it.currentCoords[i] > it.finalCoords[i] {
			it.currentCount += it.currentCoords[i] - it.finalCoords[i] - 1
			it.currentCoords[i] = it.startCoords[i]
		} else {
			it.currentCoords[i] = it.currentCoords[i] + 1
			it.currentCount++
			return it.currentCoords
		}
	}
	panic("Tried to access coordinate beyond boundary")
}

type StridingCoordIterator struct {
	strides, corner1, currentCoords []int
	coordIter                       *CoordIterator
}

func NewStridingCoordIterator(corner1, corner2, strides []int) *StridingCoordIterator {
	it := &StridingCoordIterator{corner1: corner1, strides: strides}
	// (Difference between corner2 and corner1) / strides
	adjustedDifference := DivideIntSlices(SubtractIntSlices(corner2, corner1), strides)
	it.coordIter = NewCoordIterator(make([]int, len(corner1)), adjustedDifference)
	return it
}

func (it *StridingCoordIterator) GetCurrentCoords() []int { return it.currentCoords }
func (it *StridingCoordIterator) GetCurrentCount() int    { return it.coordIter.GetCurrentCount() }
func (it *StridingCoordIterator) HasNext() bool           { return it.coordIter.HasNext() }
func (it *StridingCoordIterator) Next() []int {
	it.currentCoords = AddIntSlices(MulIntSlices(it.coordIter.Next(), it.strides), it.corner1)
	return it.currentCoords
}

// RegionsIterator Iterates through all possible regions (of a certain size) that can be made from this tensor.
// Can be instantiated with or without strides.
type RegionsIterator struct {
	regionSizes  []int
	bottomCorner []int
	topCorner    []int

	CoordIterator CoordIteratorInterface

	tensor *Tensor // Reference to tensor so we can perform operations on it
}

func (it *RegionsIterator) setup(regionSizes, padding []int) {
	regionSizeCopy := regionSizes

	//Ensure regions are the same size by appending dimensions of length 1
	if len(regionSizeCopy) < len(it.tensor.dimension) {
		regionSizeCopy = make([]int, len(it.tensor.dimension))
		for i := 0; i < len(regionSizes); i++ {
			regionSizeCopy[i] = regionSizes[i]
		}
		for i := len(regionSizes); i < len(it.tensor.dimension); i++ {
			// Pad with 1's
			regionSizeCopy[i] = 1
		}
	}

	it.regionSizes = AddIntToAll(regionSizeCopy, -1)
	it.bottomCorner = make([]int, len(it.tensor.dimension))
	it.bottomCorner = SubtractIntSlices(it.bottomCorner, padding)
	//Top corner is limited by size of region
	it.topCorner = SubtractIntSlices(it.tensor.dimension, regionSizeCopy)
	//And extended by padding
	it.topCorner = AddIntSlices(it.topCorner, padding)

}

func NewRegionsIterator(t *Tensor, regionSizes, padding []int) *RegionsIterator {
	it := &RegionsIterator{tensor: t}
	it.setup(regionSizes, padding)

	it.CoordIterator = NewCoordIterator(it.bottomCorner, it.topCorner)
	return it
}

func NewRegionsIteratorWithStrides(t *Tensor, regionSizes, padding, strides []int) *RegionsIterator {
	it := &RegionsIterator{tensor: t}
	it.setup(regionSizes, padding)
	it.CoordIterator = NewStridingCoordIterator(it.bottomCorner, it.topCorner, strides)
	return it
}

func (it *RegionsIterator) HasNext() bool {
	return it.CoordIterator.HasNext()
}

func (it *RegionsIterator) Next() *Tensor {
	regionBottomCorner := it.CoordIterator.Next()
	regionTopCorner := AddIntSlices(regionBottomCorner, it.regionSizes)
	return it.tensor.Region(regionBottomCorner, regionTopCorner)
}

// RegionsIteratorIteratorIterates through all possible regions (of a certain size) that can be made from this tensor.
// While RegionsIterator returns Tensors, this iterator returns Iterators.
// Can be instantiated with or without strides.
type RegionsIteratorIterator struct {
	iter *RegionsIterator
}

func NewRegionsIteratorIterator(tensor *Tensor, regionSizes, padding []int) *RegionsIteratorIterator {
	it := &RegionsIteratorIterator{iter: NewRegionsIterator(tensor, regionSizes, padding)}
	return it
}

func (it *RegionsIteratorIterator) HasNext() bool { return it.iter.HasNext() }
func (it *RegionsIteratorIterator) Next() *ValuesIterator {
	regionBottomCorner := it.iter.CoordIterator.Next()
	regionTopCorner := AddIntSlices(regionBottomCorner, it.iter.regionSizes)

	return NewValuesIterator(it.iter.tensor, NewCoordIterator(regionBottomCorner, regionTopCorner))
}

// ValuesIterator iterates through values in the tensor.
// Primarily used to improve performance on functions that demand large regions to be extracted.
type ValuesIterator struct {
	iter   *CoordIterator
	tensor *Tensor // reference to the tensor on which the iterator works
}

func NewValuesIterator(tensor *Tensor, iter *CoordIterator) *ValuesIterator {
	return &ValuesIterator{tensor: tensor, iter: iter}
}

func (it *ValuesIterator) HasNext() bool { return it.iter.HasNext() }
func (it *ValuesIterator) Next() float64 { return it.tensor.AtCoords(it.iter.Next()) }

// InnerProduct Multiply each element of base with a corresponding element of other, then sum these values.
func (it *ValuesIterator) InnerProduct(t *Tensor) float64 {
	result := 0.0
	for i := 0; i < len(t.values); i++ {
		result += it.Next() * t.values[i]
	}
	return result
}
