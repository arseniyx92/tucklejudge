package AI

import (
	"bufio"
	"fmt"
	"github.com/Arafatk/glot"

	"github.com/Arafatk/glot"
	"math"
	"math/rand"
	"os"
	"sort" // should be changes to my own package
	"strconv"
	"strings"
	"time"
)

// Pictures will be 28x28 so object lenght is 784
const OBJECT_LENGTH = 785
const DATA_SIZE = 42000
const FIB_SIZE = 46
const eps = 1e-9

// Changable constants
const PLOT_ON = false
const BUTCH_SIZE = 100
const ITERATIONS = 1000001
const NUMBER_OF_TRAINING_SAMPLES = 30000
const ITERATIONS_TO_SAVE = 10000
const ITERATIONS_TO_PRINT = 500
const ITERATIONS_TO_PLOT_SUPPLEMENTION = 50

type Perceptron struct {
	w [OBJECT_LENGTH]float64
}

type Object struct {
	x     [OBJECT_LENGTH]float64
	digit int
}

var trainObjsGlob []*Object
var testObjsGlob []*Object
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func parseObjectsFile() {
	trainObjsGlob = make([]*Object, NUMBER_OF_TRAINING_SAMPLES)
	testObjsGlob = make([]*Object, DATA_SIZE-NUMBER_OF_TRAINING_SAMPLES-1)
	f, err := os.Open("train.csv")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for i := 0; i < DATA_SIZE; i++ {
		x := make([]int, OBJECT_LENGTH)
		scanner.Scan()
		if i == 0 {
			continue
		}
		s := strings.Split(scanner.Text(), ",")
		for j := 0; j < OBJECT_LENGTH; j++ {
			x[j], _ = strconv.Atoi(s[j])
		}
		if i < NUMBER_OF_TRAINING_SAMPLES+1 {
			trainObjsGlob[i-1] = generateObject(x)
		} else {
			testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES-1] = generateObject(x)
		}
	}
}

func generateObject(v []int) *Object { // first value - digit, then goes 28x28 picture
	obj := Object{
		digit: v[0],
	}
	obj.x[0] = 1
	for i := 1; i < OBJECT_LENGTH; i++ {
		obj.x[i] = float64(v[i])
	}
	return &obj
}

func generatePerceptron() *Perceptron {
	p := Perceptron{}
	for i := 0; i < OBJECT_LENGTH; i++ {
		p.w[i] = rnd.Float64() * float64(rnd.Intn(1e4))
	}
	return &p
}

func getPerceptronFromFile(filepath string) *Perceptron {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	str := scanner.Text() + " "

	p := Perceptron{}
	indexInW := 0
	currentNumber := ""
	for i := 0; i < len(str); i++ {
		if str[i] == '[' || str[i] == ']' {
			continue
		}
		if str[i] == ' ' {
			p.w[indexInW], err = strconv.ParseFloat(currentNumber, 64)
			if err != nil {
				panic(err.Error())
			}
			indexInW++
			currentNumber = ""
		} else {
			currentNumber += string(str[i])
		}
	}
	return &p
}

func (p *Perceptron) getSuggestion(obj *Object) float64 {
	dotProduct := float64(0)
	for i := 0; i < OBJECT_LENGTH; i++ {
		dotProduct += p.w[i] * obj.x[i]
	}
	return dotProduct
}

func countMistakes(p *Perceptron, objs []*Object, digit int) int {
	mistakes := 0
	for i := 0; i < len(objs); i++ {
		suggestion := p.getSuggestion(objs[i])
		if (objs[i].digit == digit) != (suggestion >= 0) {
			mistakes++
		}
	}
	return mistakes
}

func evaluateResultCorrectness(objs []*Object, model *Perceptron, grad []float64, lr float64, digit int) float64 {
	// applying temporary alterations
	for i := 0; i < OBJECT_LENGTH; i++ {
		model.w[i] += grad[i] * lr
	}

	var results = []float64{0}
	for i := 0; i < BUTCH_SIZE; i++ {
		suggestion := 0.
		for j := 0; j < OBJECT_LENGTH; j++ {
			suggestion += model.w[j] * objs[i].x[j]
		}
		if (objs[i].digit == digit) != (suggestion >= 0.) {
			results = append(results, math.Abs(suggestion))
		}
	}
	sort.Float64s(results)

	// retrieving temporary alterations
	for i := 0; i < OBJECT_LENGTH; i++ {
		model.w[i] -= grad[i] * lr
	}

	// returning median value
	return results[(len(results)-1)/2] * float64(len(results))
}

func learn(model *Perceptron, digit int) {
	currentTypeObjs := make([]*Object, 0)
	for i := 0; i < NUMBER_OF_TRAINING_SAMPLES; i++ {
		if trainObjsGlob[i].digit == digit {
			currentTypeObjs = append(currentTypeObjs, trainObjsGlob[i])
		}
	}

	var fibs [FIB_SIZE]float64
	fibs[0], fibs[1] = 0., 1.
	for i := 2; i < FIB_SIZE; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}

	var prevD []float64
	var sqrLenPrevGrad float64

	for iteration := 0; iteration < ITERATIONS; iteration++ {
		// getting a batch of objects
		objs := make([]*Object, BUTCH_SIZE)
		for i := 0; i < BUTCH_SIZE/2; i++ {
			objs[i] = trainObjsGlob[rnd.Intn(len(trainObjsGlob))]
		}
		for i := BUTCH_SIZE / 2; i < BUTCH_SIZE; i++ {
			objs[i] = currentTypeObjs[rnd.Intn(len(currentTypeObjs))]
		}
		rnd.Shuffle(len(objs), func(i, j int) { objs[i], objs[j] = objs[j], objs[i] })
		// generating answer for each one
		mistakes := 0
		currentGrad := make([]float64, OBJECT_LENGTH)
		for i := 0; i < BUTCH_SIZE; i++ {
			suggestion := model.getSuggestion(objs[i])
			if (objs[i].digit == digit) != (suggestion >= 0.) {
				mistakes++
				sign := float64(1)
				if suggestion >= 0 {
					sign = float64(-1)
				}
				for j := 0; j < OBJECT_LENGTH; j++ {
					currentGrad[j] += sign * objs[i].x[j]
				}
			}
		}
		sqrLenCurrentGrad := float64(0)
		if mistakes > 0 {
			for i := 0; i < OBJECT_LENGTH; i++ {
				currentGrad[i] /= float64(mistakes)
				sqrLenCurrentGrad += currentGrad[i] * currentGrad[i]
			}
		}

		// creating gradient by using Fletcher Rives method
		var currentD []float64
		if iteration == 0 {
			currentD = currentGrad
		} else {
			betta := 0.
			if sqrLenPrevGrad > eps || sqrLenPrevGrad < -eps {
				betta = sqrLenCurrentGrad / sqrLenPrevGrad
			}

			currentD = make([]float64, OBJECT_LENGTH)
			for i := 0; i < OBJECT_LENGTH; i++ {
				currentD[i] = currentGrad[i] + betta*prevD[i]
			}
		}
		prevD = currentD
		sqrLenPrevGrad = sqrLenCurrentGrad

		// choosing best learning rate (step length)
		l, r := 0., .5
		pi := fibs[FIB_SIZE-2] / fibs[FIB_SIZE-1]
		m1, m2 := l+(r-l)*pi, r-(r-l)*pi
		f1 := evaluateResultCorrectness(objs, model, currentD, m1, digit)
		f2 := evaluateResultCorrectness(objs, model, currentD, m2, digit)
		for i := FIB_SIZE - 3; i > 0; i-- {
			pi = fibs[i] / fibs[i+1]
			if f1 > f2 {
				l = m1
				m1 = m2
				f1 = f2
				m2 = r - (r-l)*pi
				f2 = evaluateResultCorrectness(objs, model, currentD, m2, digit)
			} else {
				r = m2
				m2 = m1
				f2 = f1
				m1 = l + (r-l)*pi
				f1 = evaluateResultCorrectness(objs, model, currentD, m1, digit)
			}
		}
		learningRate := l

		// apply alterations to the model
		for i := 0; i < OBJECT_LENGTH; i++ {
			model.w[i] += currentD[i] * learningRate
		}
		if iteration%ITERATIONS_TO_PLOT_SUPPLEMENTION == 0 {
			addPerceptronResultsToPlot(model, digit)
		}
		if iteration%ITERATIONS_TO_PRINT == 0 {
			fmt.Println("Iteration ", iteration, ") error value: ", f1)
		}
		if iteration%ITERATIONS_TO_SAVE == 0 {
			os.WriteFile(fmt.Sprintf("perceptron%d.txt", digit), []byte(fmt.Sprintf("%v", model.w)), 0600)
			if PLOT_ON {
				err := createPlot(fmt.Sprintf("perceptron%d", digit))
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}
}

func main() {
	parseObjectsFile()
	for digit := 0; digit <= 9; digit++ {
		go trainParticularPerceptron(digit, false)
	}
	for {
	}
	//fmt.Println("MISTAKES: ")
	//for digit := 0; digit <= 9; digit++ {
	//	p := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	//	fmt.Printf("%d TRAIN: %d TEST: %d\n", digit, countMistakes(p, trainObjsGlob, digit), countMistakes(p, testObjsGlob, digit))
	//}
}

var testingDataCorrectnessPlot []float64
var trainingDataCorrectnessPlot []float64

func addPerceptronResultsToPlot(p *Perceptron, digit int) {
	trainingDataCorrectnessPlot = append(trainingDataCorrectnessPlot, 1.-float64(countMistakes(p, trainObjsGlob, digit))/float64(len(trainObjsGlob)))
	testingDataCorrectnessPlot = append(testingDataCorrectnessPlot, 1.-float64(countMistakes(p, testObjsGlob, digit))/float64(len(testObjsGlob)))
}

func createPlot(modelName string) error {
	plot, _ := glot.NewPlot(2, false, false)
	err := plot.AddPointGroup(modelName+" - Testing data results", "lines", testingDataCorrectnessPlot)
	if err != nil {
		return err
	}
	err = plot.AddPointGroup(modelName+" - Training data results", "lines", trainingDataCorrectnessPlot)
	if err != nil {
		return err
	}
	return plot.SavePlot("accuracy.png")
}

func trainParticularPerceptron(digit int, createNew bool) {
	p := generatePerceptron()
	if createNew == false {
		p = getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	}
	learn(p, digit)
}

func getPrediction(p *Perceptron, image [OBJECT_LENGTH+1]int) int {
	obj := generateObject(image[:])
	maxProbability := 0.
	matchingDigit := -1
	for digit := 0; digit <= 9; digit++ {
		p := getPerceptronFromFile(fmt.Sprintf("coreRecognitionAlgorithm/perceptrons/perceptron%d.txt", digit))
		curProbability := p.getSuggestion(obj)
		if curProbability > maxProbability {
			maxProbability = curProbability
			matchingDigit = digit
		}
	}
	return matchingDigit
}
