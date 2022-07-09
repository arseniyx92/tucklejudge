package AI

import (
	"bufio"
	"fmt"
	"github.com/Arafatk/glot"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Pictures will be 28x28 so object lenght is 784
const INF = 1e18
const OBJECT_LENGTH = 785
const DATA_SIZE = 42000 * COPIES
const FIB_SIZE = 46
const eps = 1e-9

// Changable constants
const COPIES = 1
const BUTCH_SIZE = 1000
const ITERATIONS = 301
const NUMBER_OF_TRAINING_SAMPLES = 30000 * COPIES
const ITERATIONS_TO_SAVE = INF
const ITERATIONS_TO_PRINT = 50
const ITERATIONS_TO_PLOT_SUPPLEMENTION = 1

var PLOT_ON = false

type Perceptron struct {
	w [OBJECT_LENGTH]float64
}

type Object struct {
	x     [OBJECT_LENGTH]int
	digit int
}

var trainObjsGlob []*Object
var testObjsGlob []*Object
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func parseObjectsFile() {
	trainObjsGlob = make([]*Object, NUMBER_OF_TRAINING_SAMPLES)
	testObjsGlob = make([]*Object, DATA_SIZE-NUMBER_OF_TRAINING_SAMPLES)
	f, err := os.Open("train.csv")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	for i := 0; i < DATA_SIZE; i += COPIES {
		// scanning image into y
		y := make([]int, OBJECT_LENGTH)
		scanner.Scan()
		s := strings.Split(scanner.Text(), ",")
		for j := 0; j < OBJECT_LENGTH; j++ {
			y[j], _ = strconv.Atoi(s[j])
		}

		// rotating and flipping image (x, y) -> (y, x)
		x := make([]int, OBJECT_LENGTH)
		x[0] = y[0]
		for j := 1; j < OBJECT_LENGTH; j++ {
			ind := j - 1
			posX := ind / 28
			posY := ind - 28*posX
			x[j] = y[1+28*posY+posX]
		}

		if i < NUMBER_OF_TRAINING_SAMPLES {
			trainObjsGlob[i] = generateObject(x, true)
			for k := 0; k < COPIES/2; k++ {
				trainObjsGlob[i] = rotateAlphaDeg(trainObjsGlob[i], -5)
			}
			for k := 1; k < COPIES; k++ {
				trainObjsGlob[i+k] = rotateAlphaDeg(trainObjsGlob[i+k-1], 5)
			}
		} else {
			testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES] = generateObject(x, true)
			for k := 0; k < COPIES/2; k++ {
				testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES] = rotateAlphaDeg(testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES], -5)
			}
			for k := 1; k < COPIES; k++ {
				testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k] = rotateAlphaDeg(testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k-1], 5)
			}
		}
	}
}

func degreeToRadian(deg float64) float64 {
	return deg * math.Pi / 180.
}

func rotateAlphaDeg(init *Object, deg float64) *Object {
	rad := degreeToRadian(deg)
	cos := math.Cos(rad)
	sin := math.Sin(rad)

	obj := Object{
		digit: init.digit,
	}
	for i := 1; i < OBJECT_LENGTH; i++ {
		x := (i - 1) / 28
		y := i - 1 - 28*x
		nx := int(math.Round(cos*float64(x) + sin*float64(y)))
		ny := int(math.Round(-sin*float64(x) + cos*float64(y)))
		ind := 1 + nx*28 + ny
		if ind >= 1 && ind < OBJECT_LENGTH {
			obj.x[i] = init.x[ind]
		}
	}
	//if obj.digit == 3 && deg == 20 {
	//	printObjectImage28x28(&obj)
	//	os.Exit(0)
	//}
	return &obj
}

func generateObject(v []int, threshold bool) *Object { // first value - digit, then goes 28x28 picture
	obj := Object{
		digit: v[0],
	}
	obj.x[0] = 1 // bias weight
	if threshold {
		colors := make([]int, 256)
		pref := make([]int, 256)
		mean := make([]int, 256)
		for i := 1; i < OBJECT_LENGTH; i++ {
			colors[v[i]]++
		}
		for i, v := range colors {
			pref[i] = v
			mean[i] = v * i
			if i != 0 {
				pref[i] += pref[i-1]
				mean[i] += mean[i-1]
			}
		}
		maxx := 0.
		bestInd := 0
		sz := 28 * 28
		for i, _ := range colors {
			if i == 255 {
				break
			}
			Wb := float64(pref[i]) / float64(sz)
			Ww := float64(pref[255]-pref[i]) / float64(sz)
			ub := float64(mean[i]) / float64(pref[i])
			uw := float64(mean[255]-mean[i]) / float64(pref[255]-pref[i])
			gamma := Wb * Ww * (uw - ub) * (uw - ub)
			if maxx < gamma {
				maxx = gamma
				bestInd = i
			}
		}
		for i := 1; i < OBJECT_LENGTH; i++ {
			if v[i] <= bestInd {
				obj.x[i] = 0
			} else {
				obj.x[i] = 255
			}
		}
	} else {
		for i := 1; i < OBJECT_LENGTH; i++ {
			obj.x[i] = v[i]
		}
	}
	return &obj
}

func printImage(filepath string, img image.Image) {
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}
	f, _ := os.Create(filepath)
	_ = png.Encode(f, img)
}

func printObjectImage28x28(obj *Object) {
	img := image.NewGray(image.Rect(0, 0, 28, 28))
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			img.SetGray(i, j, color.Gray{Y: uint8(obj.x[i*28+j])})
		}
	}
	printImage(fmt.Sprintf("pic%d.png", obj.digit), img)
}

func generatePerceptron(randChan chan int) *Perceptron {
	p := Perceptron{}
	// filling with random weights
	for i := 0; i < OBJECT_LENGTH; i++ {
		p.w[i] = rnd.Float64() * float64(<-randChan)
		if (<-randChan)%3 == 0 {
			p.w[i] *= -1
		}
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
		dotProduct += p.w[i] * float64(obj.x[i])
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
			suggestion += model.w[j] * float64(objs[i].x[j])
		}
		if (objs[i].digit == digit) != (suggestion >= 0.) {
			results = append(results, math.Abs(suggestion))
		}
	}
	sort.Float64s(results)
	//mistakes := countMistakes(model, trainObjsGlob, digit)

	// retrieving temporary alterations
	for i := 0; i < OBJECT_LENGTH; i++ {
		model.w[i] -= grad[i] * lr
	}

	// returning median value
	return results[(len(results)-1)/2] * float64(len(results))
	//return float64(len(results))
	//return float64(mistakes)
}

func learn(model *Perceptron, digit int) {
	currentTypeObjs := make([]*Object, 0)
	for i := 0; i < NUMBER_OF_TRAINING_SAMPLES; i++ {
		if trainObjsGlob[i].digit == digit {
			currentTypeObjs = append(currentTypeObjs, trainObjsGlob[i])
			//printObjectImage28x28(trainObjsGlob[i])
			//return
		}
	}

	var fibs [FIB_SIZE]float64
	fibs[0], fibs[1] = 0., 1.
	for i := 2; i < FIB_SIZE; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}

	var prevD []float64
	var sqrLenPrevGrad float64
	prevF := INF

	for iteration := 0; iteration < ITERATIONS; iteration++ {
		// getting a batch of objects
		objs := make([]*Object, BUTCH_SIZE)
		for i := 0; i < BUTCH_SIZE/2; i++ {
			objs[i] = currentTypeObjs[(BUTCH_SIZE*iteration+i)%len(currentTypeObjs)]
			//objs[i] = currentTypeObjs[rnd.Intn(len(currentTypeObjs))]
		}
		for i := BUTCH_SIZE / 2; i < BUTCH_SIZE; i++ {
			objs[i] = trainObjsGlob[(BUTCH_SIZE*iteration+i-BUTCH_SIZE/2)%len(trainObjsGlob)]
			//objs[i] = trainObjsGlob[rnd.Intn(len(trainObjsGlob))]
		}

		// generating answer for each butch object
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
					currentGrad[j] += sign * float64(objs[i].x[j])
				}
			}
		}
		if mistakes > 0 {
			for i := 0; i < OBJECT_LENGTH; i++ {
				currentGrad[i] /= float64(mistakes)
			}
		}

		// choosing best learning rate (step length)
		l, r := 0., .5
		m1, m2 := l+(r-l)*(fibs[FIB_SIZE-3]/fibs[FIB_SIZE-1]), l+(r-l)*(fibs[FIB_SIZE-2]/fibs[FIB_SIZE-1])
		f1 := evaluateResultCorrectness(objs, model, currentGrad, m1, digit)
		f2 := evaluateResultCorrectness(objs, model, currentGrad, m2, digit)
		for i := FIB_SIZE - 2; i > 1; i-- {
			if f1 >= f2 {
				l = m1
				m1 = m2
				f1 = f2
				m2 = l + (r-l)*(fibs[i-1]/fibs[i])
				f2 = evaluateResultCorrectness(objs, model, currentGrad, m2, digit)
			} else {
				r = m2
				m2 = m1
				f2 = f1
				m1 = l + (r-l)*(fibs[i-2]/fibs[i])
				f1 = evaluateResultCorrectness(objs, model, currentGrad, m1, digit)
			}
		}
		learningRate := l
		if f1 < prevF {
			sqrLenCurrentGrad := float64(0)
			for i := 0; i < OBJECT_LENGTH; i++ {
				currentGrad[i] += currentGrad[i] * learningRate
				sqrLenCurrentGrad += currentGrad[i] * currentGrad[i]
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
			prevF = evaluateResultCorrectness(objs, model, currentD, 1, digit)

			// apply alterations to the model
			for i := 0; i < OBJECT_LENGTH; i++ {
				model.w[i] += currentD[i]
			}
		}

		// logging
		if iteration%ITERATIONS_TO_PLOT_SUPPLEMENTION == 0 {
			addPerceptronResultsToPlot(model, digit)
		}
		if iteration%ITERATIONS_TO_PRINT == 0 {
			if countMistakes(model, trainObjsGlob, digit) < 5000 {
				saveFunc(model, digit)
				fmt.Println("NICE", countMistakes(model, trainObjsGlob, digit))
				os.Exit(0)
			}
			fmt.Println("Iteration", iteration, " of digit:", digit, ")     error value: ", countMistakes(model, trainObjsGlob, digit))
		}
		if iteration%ITERATIONS_TO_SAVE == 0 {
			saveFunc(model, digit)
		}
	}
}

func saveFunc(model *Perceptron, digit int) {
	os.WriteFile(fmt.Sprintf("perceptron%d.txt", digit), []byte(fmt.Sprintf("%v", model.w)), 0600)
	if PLOT_ON {
		err := createPlot(fmt.Sprintf("perceptron%d", digit))
		if err != nil {
			panic(err.Error())
		}
	}
}

func main() {
	parseObjectsFile()
	randChan := make(chan int)
	go func() {
		for {
			randChan <- rnd.Intn(1e4)
		}
	}()

	//train particular perceptron
	//PLOT_ON = true
	//for i := 0; i < 5; i++ {
	//	go trainParticularPerceptron(8, true, randChan)
	//}
	//for {
	//}

	//train all perceptrons
	//for digit := 0; digit <= 9; digit++ {
	//	go trainParticularPerceptron(digit, false, randChan)
	//}

	//check current perceptrons results
	//fmt.Println("MISTAKES: ")
	//for digit := 0; digit <= 9; digit++ {
	//	p := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	//	fmt.Printf("%d TRAIN: %d TEST: %d\n", digit, countMistakes(p, trainObjsGlob, digit), countMistakes(p, testObjsGlob, digit))
	//}

	// check accuracy
	digit := 8
	p := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	fmt.Println(getPerceptronAccuracy(p, testObjsGlob, digit))
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

func trainParticularPerceptron(digit int, createNew bool, randChan chan int) {
	for {
		p := generatePerceptron(randChan)
		if createNew == false {
			p = getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
		}
		learn(p, digit)
	}
}

func InitializePerceptronMesh() [10]*Perceptron {
	var perceptrons [10]*Perceptron
	for digit := 0; digit <= 9; digit++ {
		perceptrons[digit] = getPerceptronFromFile(fmt.Sprintf("perceptrons/perceptron%d.txt", digit))
	}
	return perceptrons
}

func GetPrediction(image [OBJECT_LENGTH]int, perceptrons [10]*Perceptron) int {
	obj := generateObject(image[:], true)
	maxProbability := 0.
	matchingDigit := 10
	for digit := 0; digit <= 9; digit++ {
		curProbability := perceptrons[digit].getSuggestion(obj)
		if curProbability > maxProbability {
			maxProbability = curProbability
			matchingDigit = digit
		}
	}
	return matchingDigit
}

func getPerceptronAccuracy(model *Perceptron, objs []*Object, digit int) float64 {
	mistakes := 0
	for _, obj := range objs {
		x := model.getSuggestion(obj)
		if (x >= 0) != (digit == obj.digit) {
			mistakes++
		}
	}
	return 1. - (float64(mistakes) / float64(len(objs)))
}

//MISTAKES:
//0 TRAIN: 1439 TEST: 649
//1 TRAIN: 956 TEST: 427
//2 TRAIN: 6551 TEST: 2718
//3 TRAIN: 7168 TEST: 2908
//4 TRAIN: 2511 TEST: 1076
//5 TRAIN: 8214 TEST: 3306
//6 TRAIN: 3201 TEST: 1335
//7 TRAIN: 3322 TEST: 1350
//8 TRAIN: 10889 TEST: 4440
//9 TRAIN: 6051 TEST: 2360
