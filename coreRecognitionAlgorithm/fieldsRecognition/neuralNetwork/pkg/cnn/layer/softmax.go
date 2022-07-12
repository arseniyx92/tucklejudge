package layer

import (
	"fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"math"
)

type SoftmaxLayer struct {
	outputDims  []int
	recentInput maths.Tensor
}

func (o *SoftmaxLayer) LoadInfo(info string) {}

func (o *SoftmaxLayer) SaveInfo() string {
	return ""
}

func NewSoftmaxLayer(inputDims []int) *SoftmaxLayer {
	return &SoftmaxLayer{
		outputDims: inputDims}
}

func (o *SoftmaxLayer) ForwardPropagation(input maths.Tensor) maths.Tensor {
	o.recentInput = input

	output := input.Zeroes()
	expSum := 0.0

	for i := 0; i < input.Len(); i++ {
		output.SetValue(i, math.Exp(input.At(i)))
		expSum += output.At(i)
	}

	output.Apply(func(val float64, idx int) float64 {
		return val / expSum
	})

	return *output
}
func (o *SoftmaxLayer) BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor {
	return *gradient.MulElem(o.derivatives(o.recentInput))
}

func (o *SoftmaxLayer) OutputDims() []int {
	return o.outputDims
}

func (o *SoftmaxLayer) derivatives(input maths.Tensor) *maths.Tensor {
	output := make([]float64, input.Len())
	expSum := 0.0

	for i := 0; i < input.Len(); i++ {
		output[i] = math.Exp(input.At(i))
		expSum += output[i]
	}

	for i := 0; i < input.Len(); i++ {
		output[i] *= expSum - output[i]
	}
	return maths.NewTensor(input.Dimensions(), maths.DivideFloat64SliceByFloat64(output, expSum*expSum))
}
