package metrics

import (
	"fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"math"
)

type LossFunction interface {
	CalculateLoss(target, predicted []float64) maths.Tensor
	CalculateLossDerivative(target, predicted []float64) maths.Tensor
}

type CrossEntropyLoss struct{}

func (c *CrossEntropyLoss) CalculateLoss(target, predicted []float64) maths.Tensor {
	lossValues := make([]float64, len(target))
	for i := 0; i < len(lossValues); i++ {
		if target[i] == 1 {
			lossValues[i] = -1.0 * math.Log(predicted[i])
		} else {
			lossValues[i] = -1.0 * math.Log(1-predicted[i])
		}
	}
	return *maths.NewTensor([]int{len(lossValues)}, lossValues)
}

func (c *CrossEntropyLoss) CalculateLossDerivative(target, predicted []float64) maths.Tensor {
	lossDerivatives := make([]float64, len(target))

	for i := 0; i < len(lossDerivatives); i++ {
		if target[i] == 1 {
			lossDerivatives[i] = -1.0 / predicted[i]
		} else {
			lossDerivatives[i] = 1.0 / (1 - predicted[i])
		}
	}
	return *maths.NewTensor([]int{len(lossDerivatives)}, lossDerivatives)
}
