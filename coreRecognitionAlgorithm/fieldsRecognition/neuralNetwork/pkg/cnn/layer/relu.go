package layer

import (
	"fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"math"
)

type ReLULayer struct {
	outputDims  []int
	recentInput maths.Tensor
}

func (o *ReLULayer) LoadInfo(info string) {}

func (o *ReLULayer) SaveInfo() string {
	return ""
}

func NewReLULayer(inputDims []int) *ReLULayer {
	return &ReLULayer{
		outputDims: inputDims}
}

func (o *ReLULayer) derivatives(input maths.Tensor) *maths.Tensor {
	output := input.Zeroes()

	for i := 0; i < input.Len(); i++ {
		if input.At(i) > 0 {
			output.SetValue(i, 1)
		} else {
			output.SetValue(i, 0)
		}
	}

	return output
}

func (o *ReLULayer) ForwardPropagation(input maths.Tensor) maths.Tensor {
	o.recentInput = input
	output := input.Zeroes()
	for i := 0; i < input.Len(); i++ {
		output.SetValue(i, math.Max(input.At(i), 0))
	}
	return *output
}
func (o *ReLULayer) BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor {
	return *gradient.MulElem(o.derivatives(o.recentInput))
}

func (o *ReLULayer) OutputDims() []int {
	return o.outputDims
}
