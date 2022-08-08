package layer

import (
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"math"
)

type MaxPoolingLayer struct {
	strides []int
	sizes   []int

	outputTensor maths.Tensor
	inputTensor  maths.Tensor

	maxIndices []int
}

func (m *MaxPoolingLayer) LoadInfo(info string) {}

func (m *MaxPoolingLayer) SaveInfo() string {
	return ""
}

func NewMaxPoolingLayer(strides, sizes, inputDims []int) *MaxPoolingLayer {
	m := &MaxPoolingLayer{strides: strides, sizes: sizes}

	for len(m.strides) < len(inputDims) {
		m.strides = append(m.strides, 1)
	}

	for len(m.sizes) < len(inputDims) {
		m.sizes = append(m.sizes, 1)
	}

	m.inputTensor = *maths.NewTensor(inputDims, nil)

	outputDims := make([]int, len(inputDims))
	for i := 0; i < len(outputDims); i++ {
		outputDims[i] = int(math.Ceil((float64(inputDims[i]) - float64(m.sizes[i]) + 1) / float64(m.strides[i])))
	}

	m.outputTensor = *maths.NewTensor(outputDims, nil)
	m.maxIndices = make([]int, m.outputTensor.Len())
	return m
}

func (m *MaxPoolingLayer) ForwardPropagation(input maths.Tensor) maths.Tensor {
	//Apply max function across input

	// At the pooling layer, forward propagation results in an N×N pooling block being reduced to a
	// single value - value of the “winning unit”. Backpropagation of the pooling layer then computes the error
	// which is acquired by this single value “winning unit”.
	// To keep track of the “winning unit” its index noted during the forward pass and used for gradient routing
	// during backpropagation.

	for iter := maths.NewRegionsIteratorWithStrides(&input, m.sizes, []int{}, m.strides); iter.HasNext(); {
		nextRegion := iter.Next()
		maxIndex := nextRegion.MaxValueIndex()

		m.maxIndices[iter.CoordIterator.GetCurrentCount()-1] = maxIndex
		m.outputTensor.SetValue(iter.CoordIterator.GetCurrentCount()-1, nextRegion.At(maxIndex))
	}

	return m.outputTensor
}
func (m *MaxPoolingLayer) BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor {
	inputGradients := m.inputTensor.Zeroes() // Creates a new tensor with the same dimensions, but zero-valued

	// the error is just assigned to where it comes from - the “winning unit” because other units in the previous
	// layer’s pooling blocks did not contribute to it hence all the other assigned values of zero
	for iter := maths.NewRegionsIteratorWithStrides(inputGradients, m.sizes, []int{}, m.strides); iter.HasNext(); {
		iter.Next()

		maxIndex := m.maxIndices[iter.CoordIterator.GetCurrentCount()-1]
		maxCoords := maths.HornerToCoords(maxIndex, m.sizes)

		regionStart := iter.CoordIterator.GetCurrentCoords()
		coordsOfMax := maths.AddIntSlices(regionStart, maxCoords)

		inputGradients.Set(coordsOfMax, gradient.At(iter.CoordIterator.GetCurrentCount()-1))
	}

	return *inputGradients
}

func (m *MaxPoolingLayer) OutputDims() []int { return m.outputTensor.Dimensions() }
