package layer

import (
	"fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"fmt"
	"strconv"
	"strings"
)

type FullyConnectedLayer struct {
	weights maths.Tensor
	biases  []float64

	inputDims  []int
	outputDims []int

	recentOutput []float64
	recentInput  maths.Tensor
}

func NewFullyConnectedLayer(outputLength int, inputDims []int) *FullyConnectedLayer {
	dense := &FullyConnectedLayer{}
	dense.inputDims = inputDims
	dense.recentInput = *maths.NewTensor(inputDims, nil)
	dense.recentOutput = make([]float64, outputLength)
	dense.outputDims = []int{outputLength}

	dense.weights = *maths.NewTensor(append(inputDims, outputLength), nil)
	dense.weights = *dense.weights.Randomize()

	dense.biases = make([]float64, outputLength)

	return dense
}

func (d *FullyConnectedLayer) LoadInfo(info string) {
	for i, str := range strings.Split(info, "\n") {
		if i == 0 {
			d.weights.LoadTensor(str)
		} else {
			for j, y := range strings.Split(str, " ") {
				x, err := strconv.ParseFloat(y, 64)
				if err != nil {
					panic(err)
				}
				d.biases[j] = x
			}
		}
	}
}

func (d *FullyConnectedLayer) SaveInfo() string {
	info := d.weights.SaveTensor() + "\n"
	for i, b := range d.biases {
		info += fmt.Sprintf("%g", b)
		if i != len(d.biases)-1 {
			info += " "
		}
	}
	return info
}

func (d *FullyConnectedLayer) ForwardPropagation(input maths.Tensor) maths.Tensor {
	d.recentInput = input

	i := maths.NewRegionsIterator(&d.weights, d.inputDims, []int{})
	for i.HasNext() {
		d.recentOutput[i.CoordIterator.GetCurrentCount()] = i.Next().InnerProduct(&input)
	}
	d.recentOutput = func(l, r []float64) []float64 {
		ret := make([]float64, len(l))
		for i := 0; i < len(ret); i++ {
			ret[i] = l[i] + r[i]
		}
		return ret
	}(d.recentOutput, d.biases)

	return *maths.NewTensor([]int{len(d.recentOutput)}, d.recentOutput)
}

func (d *FullyConnectedLayer) BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor {
	var weightsGradient *maths.Tensor
	for i := 0; i < len(gradient.Values()); i++ {
		newGrads := d.recentInput.MulScalar(gradient.Values()[i])
		if weightsGradient == nil {
			weightsGradient = newGrads
		} else {
			weightsGradient = weightsGradient.AppendTensor(newGrads, len(d.weights.Dimensions()))
		}
	}

	inputGradient := maths.NewTensor(d.inputDims, nil)
	j := maths.NewRegionsIterator(&d.weights, d.inputDims, []int{})
	for j.HasNext() {
		newGrads := j.Next().MulScalar(gradient.Values()[j.CoordIterator.GetCurrentCount()-1])
		inputGradient = inputGradient.Add(newGrads, 1)
	}

	d.weights = *d.weights.Add(weightsGradient, -1.0*lr)
	d.biases = maths.AddFloat64Slices(d.biases, maths.MulFloat64ToSlice(gradient.Values(), -1.0*lr))

	return *inputGradient
}

func (d *FullyConnectedLayer) OutputDims() []int { return d.outputDims }
