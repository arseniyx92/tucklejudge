package layer

import (
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/maths"
)

type Layer interface {
	ForwardPropagation(input maths.Tensor) maths.Tensor
	BackwardPropagation(gradient maths.Tensor, lr float64) maths.Tensor
	LoadInfo(s string)
	SaveInfo() string

	OutputDims() []int

	//Copy() Layer
	//Mutate()
}
