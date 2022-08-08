package layer

import (
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
)

type ConvolutionLayer struct {
	filters              maths.Tensor
	filterDimensionSizes []int
	ccMapSize            []int
	outputDimensions     []int
	recentInput          maths.Tensor

	iteration int

	jobs    chan job
	results chan result
}

func NewConvolutionLayer(filterDimensionSizes []int, depth int, inputDims []int) *ConvolutionLayer {
	conv := &ConvolutionLayer{}
	conv.filterDimensionSizes = filterDimensionSizes
	for len(conv.filterDimensionSizes) < len(inputDims) {
		conv.filterDimensionSizes = append(conv.filterDimensionSizes, 1)
	}

	//Calculate the size of the cross correlation map resultant from applying a given filter
	conv.ccMapSize = maths.SubtractIntSlices(inputDims, maths.AddIntToAll(conv.filterDimensionSizes, -1))

	conv.filters = *maths.NewTensor(append(conv.filterDimensionSizes, depth), nil)

	randLimits := math.Sqrt(2) / math.Sqrt(float64(maths.ProductIntSlice(inputDims)))
	conv.filters = *conv.filters.Randomize()
	conv.filters = *conv.filters.MulScalar(randLimits)

	conv.outputDimensions = append(conv.ccMapSize, depth)

	conv.jobs = make(chan job, 1000)
	conv.results = make(chan result, 1000)

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(conv.jobs, conv.results)
	}

	//n, err := conv.SaveFiltersAsImages("./filters")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Saved %d filters to images\n", n)
	return conv
}

func (conv *ConvolutionLayer) LoadInfo(info string) {
	conv.filters.LoadTensor(info)
}

func (conv *ConvolutionLayer) SaveInfo() string {
	return conv.filters.SaveTensor()
}

type job struct {
	i int
	f func(i int) float64
}

type result struct {
	i int
	v float64
}

func worker(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		results <- result{i: j.i, v: j.f(j.i)}
	}
}

// crossCorrelationMap creates a cross-correlation map based on a filter.
// ccMapSize is passed so we don't have to recalculate it every time.
// base is the input where the filter needs to be applied to.
func (c *ConvolutionLayer) crossCorrelationMap(base, filter *maths.Tensor, ccMapSize, padding []int) *maths.Tensor {
	ccMapValues := make([]float64, maths.ProductIntSlice(ccMapSize))

	numOps := 0
	// "slide" the filter over the base.
	// For every 'region' where the filter can come, calculate the inner product of that part of base with the filter.
	// assign this product to the cross-correlation map.
	rii := maths.NewRegionsIteratorIterator(base, filter.Dimensions(), padding)
	for rii.HasNext() {
		ccMapValues[numOps] = rii.Next().InnerProduct(filter)
		numOps++
	}

	// assert numOps != len(ccMapValues)
	if numOps != len(ccMapValues) {
		panic("number of InnerProduct operations is not equal to len(ccMapValues)")
	}

	// Return a tensor with the cross-correlation map
	return maths.NewTensor(ccMapSize, ccMapValues)
}

// crossCorrelationMap creates a cross-correlation map based on a filter.
// ccMapSize is passed so we don't have to recalculate it every time.
// base is the input where the filter needs to be applied to.
func (c *ConvolutionLayer) crossCorrelationMapParallel(base, filter *maths.Tensor, ccMapSize, padding []int) *maths.Tensor {
	ccMapValues := make([]float64, maths.ProductIntSlice(ccMapSize))

	// We append all regions to slice so we can process each region in parallel. Which might be faster if the regions
	// are large enough
	var regions []*maths.ValuesIterator
	rii := maths.NewRegionsIteratorIterator(base, filter.Dimensions(), padding)
	for rii.HasNext() {
		regions = append(regions, rii.Next())
	}

	// assert regions is not nil
	if regions == nil {
		panic("regions is nil, something went wrong")
	}

	// length of ccMapValues needs to be the same length of regions, else something went wrong
	if len(ccMapValues) != len(regions) {
		panic(fmt.Sprintf("len(ccMapValues)=%d != len(regions)=%d", len(ccMapValues), len(regions)))
	}

	// process each regions Inner Product in parallel using n = runtime.NumCPU() workers
	for i := range ccMapValues {
		// Send job to worker
		c.jobs <- job{i: i, f: func(i int) float64 { return regions[i].InnerProduct(filter) }}
	}
	for range ccMapValues {
		// receive result and set result in ccMapValues
		res := <-c.results
		ccMapValues[res.i] = res.v
	}

	// Return a tensor with the cross-correlation map
	return maths.NewTensor(ccMapSize, ccMapValues)
}

func (c *ConvolutionLayer) ForwardPropagation(input maths.Tensor) maths.Tensor {
	c.recentInput = input // might want to rewrite this because it blocks parallel batches
	var output *maths.Tensor

	// Iterate through each filter, appending result to output tensor.
	// A filter is a "sub-tensor" of the filters tensor.
	for iter := maths.NewRegionsIterator(&c.filters, c.filterDimensionSizes, []int{}); iter.HasNext(); {
		// Compute the cross-correlation map from the input tensor with a filter.
		// The size of the cross-correlation map has been computed before-hand.
		// No padding
		newMap := c.crossCorrelationMap(&input, iter.Next(), c.ccMapSize, []int{})
		if output == nil {
			output = newMap
		} else {
			output = output.AppendTensor(newMap, len(c.filters.Dimensions()))
		}
	}

	if output == nil {
		panic("ConvolutionLayer.ForwardPropagation did not work. output == nil")
	}
	return *output
}

func (c *ConvolutionLayer) BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor {
	var filterGradients *maths.Tensor
	inputGradients := c.recentInput.Zeroes()

	// We can think of gradient as a series of "gradient filters" which we apply to the recent input.
	// This is the size of each of those filters.
	outputGradientSize := gradient.FirstDimsCopy(len(gradient.Dimensions()) - 1)

	//Iterate through each output grad, appending result to filter grad tensor.
	//Simultaneously, iterate through our original filters, altering them via gradient descent
	iterGradient := maths.NewRegionsIterator(&gradient, outputGradientSize, []int{})
	iterFilter := maths.NewRegionsIterator(&c.filters, c.filterDimensionSizes, []int{})

	for iterGradient.HasNext() && iterFilter.HasNext() {
		outputLayer := iterGradient.Next()
		filterLayer := iterFilter.Next()

		// Calculate derivation filters
		newMap := c.crossCorrelationMap(&c.recentInput, outputLayer, c.filterDimensionSizes, []int{})
		// Append or assign new map to filter grad tensor
		if filterGradients == nil {
			filterGradients = newMap
		} else {
			filterGradients = filterGradients.AppendTensor(newMap, len(gradient.Dimensions()))
		}

		// Calculate derivation input
		// derivation input = derivation  outputs * flipped filters
		// To return the correct sized tensor, this requires some padding - which happens to be (filter sizes - 1)
		padding := maths.AddIntToAll(filterLayer.Dimensions(), -1)
		flippedFilter := filterLayer.Flip()
		currentInputGradient := c.crossCorrelationMap(outputLayer, flippedFilter, inputGradients.Dimensions(), padding)

		inputGradients = inputGradients.Add(currentInputGradient, 1)
	}

	if filterGradients == nil {
		panic("BackwardPropagation did not work correctly in convolution layer. filterGradients == nil")
	}

	// Gradient descent on filters
	c.filters = *c.filters.Add(filterGradients, -1*lr)

	// Save each filter as an image. This allows for visualisation of the changes to the filter
	//if c.iteration < 100 {
	//	if err := os.Mkdir(fmt.Sprintf("./filters/filters-iteration-%d", c.iteration), 0777); err != nil {
	//		fmt.Println(err)
	//	}
	//	_, err := c.SaveFiltersAsImages(fmt.Sprintf("./filters/filters-iteration-%d", c.iteration))
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}
	//c.iteration++
	return *inputGradients
}

func (c *ConvolutionLayer) OutputDims() []int { return c.outputDimensions }

// SaveFiltersAsImages saves the filters to images relative to 'path'
// returns the amount of images saved.
// saves as grayscale for now
func (c *ConvolutionLayer) SaveFiltersAsImages(path string) (int, error) {
	numFilters := 0
	for iter := maths.NewRegionsIterator(&c.filters, c.filterDimensionSizes, []int{}); iter.HasNext(); {
		t := iter.Next() // grab a filter

		pixels := make([]uint8, len(t.Values())) // allocate pixels

		for p := 0; p < len(pixels); p++ {
			c := 255 - uint8(t.At(p)*255)
			pixels[p] = c
		}
		gray := image.NewGray(image.Rect(0, 0, c.filterDimensionSizes[0], c.filterDimensionSizes[1]))

		pixelIndex := 0
		for x := 0; x < c.filterDimensionSizes[0]; x++ {
			for y := 0; y < c.filterDimensionSizes[1]; y++ {
				gray.SetGray(x, y, color.Gray{Y: pixels[pixelIndex]})
				pixelIndex++
			}
		}
		f, err := os.Create(fmt.Sprintf("%s/filter_%d.png", path, numFilters))
		if err != nil {
			return numFilters, err
		}

		if err := png.Encode(f, gray); err != nil {
			return numFilters, err
		}

		if err := f.Close(); err != nil {
			return numFilters, err
		}

		numFilters++
	}

	return numFilters, nil
}
