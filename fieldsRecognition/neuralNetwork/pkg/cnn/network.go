package cnn

import (
	"bufio"
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/layer"
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"tucklejudge/fieldsRecognition/neuralNetwork/pkg/cnn/metrics"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
)

type Network struct {
	layers       []layer.Layer
	inputDims    []int
	learningRate float64
	loss         metrics.LossFunction
}

func New(inputDims []int, learningRate float64, loss metrics.LossFunction) *Network {
	return &Network{
		inputDims:    inputDims,
		learningRate: learningRate,
		layers:       []layer.Layer{},
		loss:         loss}
}

func (n *Network) SaveNetworkToFolder() {
	for i := 0; i < len(n.layers); i++ {
		os.WriteFile(fmt.Sprintf("fieldsRecognition/networkWeights/%d.txt", i), []byte(n.layers[i].SaveInfo()), 0600)
	}
}

func (n *Network) LoadConvolutionLayer(filterDimensions []int, filterCount int) *Network {
	n.AddConvolutionLayer(filterDimensions, filterCount)
	f, err := os.Open(fmt.Sprintf("fieldsRecognition/networkWeights/%d.txt", len(n.layers)-1))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	n.layers[len(n.layers)-1].LoadInfo(scanner.Text())
	return n
}

func (n *Network) LoadFullyConnectedLayer(outputLength int) *Network {
	n.AddFullyConnectedLayer(outputLength)
	b, err := ioutil.ReadFile(fmt.Sprintf("fieldsRecognition/networkWeights/%d.txt", len(n.layers)-1)) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	n.layers[len(n.layers)-1].LoadInfo(string(b))
	return n
}

func (n *Network) SetLearningRate(rate float64) {
	n.learningRate = rate
}

func (n *Network) LearningRate() float64 { return n.learningRate }

func (n *Network) AddConvolutionLayer(filterDimensions []int, filterCount int) *Network {
	var dims []int
	if len(n.layers) == 0 {
		dims = n.inputDims
	} else {
		dims = n.layers[len(n.layers)-1].OutputDims()
	}
	n.layers = append(n.layers, layer.NewConvolutionLayer(filterDimensions, filterCount, dims))
	return n
}

func (n *Network) AddMaxPoolingLayer(stride int, dimensions []int) *Network {
	var dims []int
	if len(n.layers) == 0 {
		dims = n.inputDims
	} else {
		dims = n.layers[len(n.layers)-1].OutputDims()
	}
	strides := make([]int, len(dimensions))
	for i := 0; i < len(strides); i++ {
		strides[i] = stride
	}

	n.layers = append(n.layers, layer.NewMaxPoolingLayer(strides, dimensions, dims))
	return n
}

func (n *Network) AddFullyConnectedLayer(outputLength int) *Network {
	var dims []int
	if len(n.layers) == 0 {
		dims = n.inputDims
	} else {
		dims = n.layers[len(n.layers)-1].OutputDims()
	}
	n.layers = append(n.layers, layer.NewFullyConnectedLayer(outputLength, dims))
	return n
}

func (n *Network) AddReLULayer() *Network {
	var dims []int
	if len(n.layers) == 0 {
		dims = n.inputDims
	} else {
		dims = n.layers[len(n.layers)-1].OutputDims()
	}
	n.layers = append(n.layers, layer.NewReLULayer(dims))
	return n
}

func (n *Network) AddSoftmaxLayer() *Network {
	var dims []int
	if len(n.layers) == 0 {
		dims = n.inputDims
	} else {
		dims = n.layers[len(n.layers)-1].OutputDims()
	}

	n.layers = append(n.layers, layer.NewSoftmaxLayer(dims))
	return n
}

// Fit will train the CNN. inputs are the inputs, labels are the labels.
// epochs are the amount of times the network is fitted
// if valInputs and valLabels != nil a validation step is ran on that data after each epoch
// batchSize is the size of every propagation batch.
// if verbose then logging is enabled and is written to to stdout with fmt
// every 'logRate' of iterations a message is written when verbose == true
// onBatchDone is a callback that is called every time a batch is done. This can be used to reduce the learning rate for example
func (n *Network) Fit(inputs, labels, valInputs, valLabels []maths.Tensor, epochs int, batchSize int, verbose bool, logRate int, onEpochDone func()) {
	fmt.Println("Fit: ignoring batch size")
	if len(labels) != len(inputs) {
		panic("length of labels is not equal to length of inputs")
	}

	for epoch := 0; epoch < epochs; epoch++ {
		fmt.Printf("Starting epoch: %d\n", epoch)
		rand.Shuffle(len(inputs), func(i, j int) {
			inputs[i], inputs[j] = inputs[j], inputs[i]
			labels[i], labels[j] = labels[j], labels[i]
		})
		averageLoss := 0.0
		accuracy := 0.0
		for i := range inputs {
			// train the network
			output := n.forward(inputs[i])
			// Use the loss as input for the backpropagation
			n.backward(n.loss.CalculateLossDerivative(labels[i].Values(), output.Values()))

			if verbose {
				loss := n.loss.CalculateLoss(labels[i].Values(), output.Values())
				averageLoss += maths.SumFloat64Slice(loss.Values())
				if maths.FindMaxIndexFloat64Slice(labels[i].Values()) == maths.FindMaxIndexFloat64Slice(output.Values()) {
					accuracy++
				}
				if i%logRate == 0 {
					fmt.Printf("Input: %d / %d, average loss for last %d iterations was %f\n", i, len(inputs), logRate, averageLoss/float64(logRate))
					fmt.Printf("Accuracy for the last %d iterations was %.2f\n", logRate, accuracy/float64(logRate))
					fmt.Printf("Using learning rate of %f\n", n.learningRate)
					averageLoss = 0
					accuracy = 0
				}
			}
		}
		if valLabels != nil && valInputs != nil {
			n.Validate(valInputs, valLabels)
		}
		fmt.Printf("Completed epoch: %d\n", epoch)
		if onEpochDone != nil {
			onEpochDone()
		}
		n.SaveNetworkToFolder()
	}
}

func (n *Network) Validate(inputs []maths.Tensor, labels []maths.Tensor) {
	if len(labels) != len(inputs) {
		panic("length of labels is not equal to length of inputs")
	}
	fmt.Printf("Validating network with %d inputs...\n", len(inputs))

	averageLoss := 0.0
	accuracy := 0.0
	for i := range inputs {
		// train the network
		output := n.forward(inputs[i])
		loss := n.loss.CalculateLoss(labels[i].Values(), output.Values())
		averageLoss += maths.SumFloat64Slice(loss.Values())
		if maths.FindMaxIndexFloat64Slice(labels[i].Values()) == maths.FindMaxIndexFloat64Slice(output.Values()) {
			accuracy++
		}
	}
	fmt.Printf("Validation average loss: %f\n", averageLoss/float64(len(inputs)))
	fmt.Printf("Validation accuracy: %.2f\n", accuracy/float64(len(inputs)))
}

// Returns a slice of probabilities
func (n *Network) Predict(input maths.Tensor) []float64 {
	output := n.forward(input)
	return output.Values()
}

func (n *Network) GetDigitPredictionFromImageArray(dims, img []int) []int {
	if img == nil {
		return []int{10}
	}
	input := ArrImageToTensor(dims, img)
	predictions := n.Predict(input)
	digits := make([]int, 10)
	for i := range digits {
		digits[i] = i
	}
	sort.Slice(digits, func(i, j int) bool {
		return predictions[digits[i]] > predictions[digits[j]]
	})
	return digits
}

func ArrImageToTensor(dims, img []int) maths.Tensor {
	values := make([]float64, dims[0]*dims[1])
	for i := 0; i < dims[0]; i++ {
		for j := 0; j < dims[1]; j++ {
			values[i+j*dims[0]] = (128. - float64(img[i*dims[1]+j])) / 255.
		}
	}
	printObjectImage28x28(values)
	return *maths.NewTensor(dims, values)
}

func printObjectImage28x28(x []float64) {
	img := image.NewGray(image.Rect(0, 0, 28, 28))
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			img.SetGray(i, j, color.Gray{Y: uint8(-(255.*x[i*28+j] - 128.))})
		}
	}
	printImage(fmt.Sprintf("lol.png"), img)
}

func printImage(filepath string, img image.Image) {
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}
	f, _ := os.Create(filepath)
	_ = png.Encode(f, img)
}

// Returns the highest index from the prediction
func (n *Network) PredictIndex(input maths.Tensor) int {
	output := n.forward(input)
	return maths.FindMaxIndexFloat64Slice(output.Values())
}

func (n *Network) forward(input maths.Tensor) maths.Tensor {
	output := input

	for _, l := range n.layers {
		output = l.ForwardPropagation(output)
	}

	return output
}

func (n *Network) backward(outputGradient maths.Tensor) maths.Tensor {
	inputGradient := outputGradient

	for i := len(n.layers) - 1; i >= 0; i-- {
		inputGradient = n.layers[i].BackwardPropagation(inputGradient, n.learningRate)

	}
	return inputGradient
}

//// Copy is a deep copy of the network, this is useful for GA's when mutating the network
//func (n *Network) Copy() *Network {
//
//	layers := make([]layer.Layer, len(n.layers))
//	for i := 0; i < len(layers); i++ {
//		layers[i] = n.layers[i].Copy()
//	}
//
//	network := &Network{
//		layers:       layers,
//		inputDims:    n.inputDims,
//		learningRate: n.learningRate,
//		loss:         n.loss,
//	}
//	return network
//}
//
//func (n *Network) Mutate() {
//
//}
