package main

import (
	"bufio"
	"fmt"
	"github.com/Arafatk/glot"
	"golang.org/x/exp/constraints"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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

const BUTCH_SIZE = 2000
const STRIKE_TO_FINISH = 301
const ITERATIONS = 3001

const NUMBER_OF_TRAINING_SAMPLES = 30000 * COPIES
const ITERATIONS_TO_SAVE = 100
const ITERATIONS_TO_PRINT = 50
const ITERATIONS_TO_PLOT_SUPPLEMENTION = 1

var minError int = INF

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
var mu sync.Mutex

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
				trainObjsGlob[i] = moveUp(trainObjsGlob[i], 1)
			}
			for k := 1; k < COPIES; k++ {
				trainObjsGlob[i+k] = moveUp(trainObjsGlob[i+k-1], -1)
			}
			//for k := 0; k < COPIES; k++ {
			//	trainObjsGlob[i+k] = maxPooling(trainObjsGlob[i+k], 2)
			//}
		} else {
			testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES] = generateObject(x, true)
			for k := 0; k < COPIES/2; k++ {
				testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES] = moveUp(testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES], 1)
			}
			for k := 1; k < COPIES; k++ {
				testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k] = moveUp(testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k-1], -1)
			}
			//for k := 0; k < COPIES; k++ {
			//	testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k] = maxPooling(testObjsGlob[i-NUMBER_OF_TRAINING_SAMPLES+k], 2)
			//}
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

func moveUp(init *Object, pixels int) *Object {
	obj := Object{
		digit: init.digit,
	}
	for i := 1; i < OBJECT_LENGTH; i++ {
		x := (i - 1) / 28
		y := i - 1 - 28*x
		nx := x
		ny := y + pixels
		ind := 1 + nx*28 + ny
		if ind >= 1 && ind < OBJECT_LENGTH {
			obj.x[i] = init.x[ind]
		}
	}
	return &obj
}

func generateObject(v []int, threshold bool) *Object { // first value - digit, then goes 28x28 picture
	obj := Object{
		digit: v[0],
	}
	obj.x[0] = 1 // bias weight
	if threshold && false {
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
	centralizeMainComponent(&obj)
	gaussianBlur(&obj, 2)
	//if obj.digit == 5 {
	//	printObjectImage28x28(&obj)
	//	os.Exit(0)
	//}
	return &obj
}

type IntPair struct {
	first, second int
}

func centralizeMainComponent(init *Object) {
	obj := &Object{
		digit: init.digit,
		x:     init.x,
	}
	otsuThreshold(obj)
	comp := make([]IntPair, 0)
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			if obj.x[1+i*28+j] == 255 {
				comp = append(comp, IntPair{i, j})
			}
		}
	}
	boundsX := IntPair{28, 0}
	boundsY := IntPair{28, 0}
	centerX := 0
	centerY := 0
	for _, pos := range comp {
		boundsX.first = min(boundsX.first, pos.first)
		boundsX.second = max(boundsX.second, pos.first)
		boundsY.first = min(boundsY.first, pos.second)
		boundsY.second = max(boundsY.second, pos.second)
		centerX += pos.first
		centerY += pos.second
	}
	centerX /= len(comp)
	centerY /= len(comp)
	scale := max(float64(boundsY.second-boundsY.first)/18., float64(boundsX.second-boundsX.first)/16.)
	boundsX.first = int(float64(boundsX.first)/scale) - centerX
	boundsX.second = int(float64(boundsX.second)/scale) - centerX
	boundsY.first = int(float64(boundsY.first)/scale) - centerY
	boundsY.second = int(float64(boundsY.second)/scale) - centerY

	for i := -28; i < 28; i++ {
		for j := -28; j < 28; j++ {
			nx := i + 14 + (boundsX.second-boundsX.first)/2 - boundsX.second
			ny := j + 24 - boundsY.second
			x := int(math.Round(scale * float64(i+centerX)))
			y := int(math.Round(scale * float64(j+centerY)))
			if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 && x >= 0 && x < 28 && y >= 0 && y < 28 {
				init.x[1+nx*28+ny] = obj.x[1+x*28+y]
			}
		}
	}

	//// shift down main component
	//for j := 27; j >= 0; j-- {
	//	cnt := 0
	//	for i := 0; i < 28; i++ {
	//		cnt += int(img.GrayAt(i, j).Y)
	//	}
	//	if cnt > 0 {
	//		shift := j - 27
	//		cpy := copyGray(img)
	//		for x := 0; x < 28; x++ {
	//			for y := 0; y < 28; y++ {
	//				img.SetGray(x, y, cpy.GrayAt(x, y+shift))
	//			}
	//		}
	//		break
	//	}
	//}
}

func otsuThreshold(obj *Object) {
	lx := 0
	ly := 0
	rx := 27
	ry := 27
	sz := (rx - lx + 1) * (ry - ly + 1)
	colors := make([]int, 256)
	pref := make([]int, 256)
	mean := make([]int, 256)
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			colors[obj.x[1+i*28+j]]++
		}
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
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			if obj.x[1+i*28+j] <= bestInd {
				obj.x[1+i*28+j] = 0
			} else {
				obj.x[1+i*28+j] = 255
			}
		}
	}
}

func gaussianBlur(obj *Object, kernel int) {
	lx := 0
	rx := 27
	ly := 0
	ry := 27
	pref := make([][]int, rx-lx+1)
	pref[0] = make([]int, ry-ly+1)
	pref[0][0] = obj.x[1]
	for j := ly + 1; j <= ry; j++ {
		pref[0][j] = pref[0][j-1] + obj.x[1+j]
	}
	for i := lx + 1; i <= rx; i++ {
		pref[i] = make([]int, ry-ly+1)
		pref[i][0] = pref[i-1][0] + obj.x[1+i*28]
		for j := ly + 1; j <= ry; j++ {
			pref[i][j] = pref[i-1][j] + pref[i][j-1] - pref[i-1][j-1] + obj.x[1+i*28+j]
		}
	}
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			maxX := min(rx, i+kernel-1)
			maxY := min(ry, j+kernel-1)
			minX := max(0, i-kernel+1)
			minY := max(0, j-kernel+1)
			sum := pref[maxX][maxY]
			cnt := (maxX - minX + 1) * (maxY - minY + 1)
			if minX != 0 && minY != 0 {
				sum += -pref[maxX][minY-1] - pref[minX-1][maxY] + pref[minX-1][minY-1]
			} else if minX != 0 {
				sum -= pref[minX-1][maxY]
			} else if minY != 0 {
				sum -= pref[maxX][minY-1]
			}
			obj.x[1+i*28+j] = int(float64(sum) / float64(cnt))
		}
	}
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
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
			img.SetGray(i, j, color.Gray{Y: uint8(obj.x[1+i*28+j])})
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

func generateEmptyPerceptron() *Perceptron {
	p := Perceptron{}
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

func activationFunction(x float64) int {
	if x >= 0 {
		return 1
	}
	return 0
}

func maxPooling(init *Object, kernel int) *Object {
	obj := Object{
		digit: init.digit,
	}
	for si := 0; si < 28; si += kernel {
		for sj := 0; sj < 28; sj += kernel {
			maxx := 0
			for i := si; i < si+kernel; i++ {
				for j := sj; j < sj+kernel; j++ {
					if maxx < init.x[1+28*i+j] {
						maxx = init.x[1+28*i+j]
					}
				}
			}
			obj.x[1+28*si/kernel+sj/kernel] = maxx
		}
	}
	//printObjectImage28x28(&obj)
	//os.Exit(0)
	return &obj
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

func countMistakesDistinguish(p *Perceptron, objs []*Object, digit0, digit1 int) int {
	mistakes := 0
	for _, obj := range objs {
		if (p.getSuggestion(obj) == 0) != (obj.digit == digit0) {
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

func evaluateResultDistinguishCorrectness(objs []*Object, model *Perceptron, grad []float64, lr float64, digit0, digit1 int) int {
	for i := 0; i < OBJECT_LENGTH; i++ {
		model.w[i] += grad[i] * lr
	}
	mistakes := countMistakesDistinguish(model, objs, digit0, digit1)
	for i := 0; i < OBJECT_LENGTH; i++ {
		model.w[i] -= grad[i] * lr
	}
	return mistakes
}

func getParticularDigitObjectButch(objsGlob []*Object, digit int) (currentTypeObjs []*Object) {
	for i := 0; i < len(objsGlob); i++ {
		if objsGlob[i].digit == digit {
			currentTypeObjs = append(currentTypeObjs, objsGlob[i])
		}
	}
	return currentTypeObjs
}

func logging(iteration int, model *Perceptron, digit int) {
	if iteration%3 == 0 && iteration > 150 {
		currentError := countMistakes(model, trainObjsGlob, digit)
		if currentError < minError {
			saveFunc(model, digit)
			minError = currentError
			fmt.Println("NICE", currentError)
			//os.Exit(0)
		}
	}
	if PLOT_ON && iteration%ITERATIONS_TO_PLOT_SUPPLEMENTION == 0 {
		addPerceptronResultsToPlot(model, digit)
	}
	if iteration%ITERATIONS_TO_PRINT == 0 {
		currentError := countMistakes(model, trainObjsGlob, digit)
		fmt.Println("Iteration", iteration, " of digit:", digit, ")     error value: ", float64(currentError)/float64(len(trainObjsGlob)))
	}
	if PLOT_ON && iteration%ITERATIONS_TO_SAVE == 0 {
		saveFunc(model, digit)
	}
}

func loggingDistinguisher(model *Perceptron, iteration, digit0, digit1 int, objs []*Object) {
	if iteration%ITERATIONS_TO_PRINT == 0 {
		fmt.Println("Iteration", iteration, " of digits:", digit0, ",", digit1, ")     error value: ", float64(countMistakesDistinguish(model, objs, digit0, digit1))/float64(len(objs)))
	}
}

func learn(model *Perceptron, digit int) {
	gnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	currentTypeObjs := getParticularDigitObjectButch(trainObjsGlob, digit)

	var fibs [FIB_SIZE]float64
	fibs[0], fibs[1] = 0., 1.
	for i := 2; i < FIB_SIZE; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}

	var prevD []float64
	var sqrLenPrevGrad float64
	prevF := INF
	failureCounter := 0

	for iteration := 0; iteration < ITERATIONS; iteration++ {
		// getting a batch of objects
		objs := make([]*Object, BUTCH_SIZE)
		for i := 0; i < BUTCH_SIZE/2; i++ {
			//objs[i] = currentTypeObjs[(BUTCH_SIZE*iteration+i)%len(currentTypeObjs)]
			objs[i] = currentTypeObjs[gnd.Intn(len(currentTypeObjs))]
		}
		for i := BUTCH_SIZE / 2; i < BUTCH_SIZE; i++ {
			//objs[i] = trainObjsGlob[(BUTCH_SIZE*iteration+i-BUTCH_SIZE/2)%len(trainObjsGlob)]
			objs[i] = trainObjsGlob[gnd.Intn(len(trainObjsGlob))]
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
					currentGrad[j] += sign * 2 * float64(objs[i].x[j])
				}
			}
		}
		if mistakes > 0 {
			for i := 0; i < OBJECT_LENGTH; i++ {
				currentGrad[i] /= float64(mistakes)
			}
		}

		// choosing best learning rate (step length)
		l, r := 0., .05
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
			failureCounter = 0
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
		} else {
			failureCounter++
			if failureCounter == STRIKE_TO_FINISH {
				return
			}
		}

		// logging
		logging(iteration, model, digit)
	}
}

func learnToDistinguish(model *Perceptron, digit0, digit1 int) {
	var fibs [FIB_SIZE]float64
	fibs[0], fibs[1] = 0., 1.
	for i := 2; i < FIB_SIZE; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}

	objs0 := getParticularDigitObjectButch(trainObjsGlob, digit0)
	objs1 := getParticularDigitObjectButch(trainObjsGlob, digit1)
	objs0 = append(objs0, objs1...)
	objs := objs0
	mu.Lock()
	rnd.Shuffle(len(objs), func(i, j int) {
		objs[i], objs[j] = objs[j], objs[i]
	})
	mu.Unlock()

	for iteration := 0; iteration < ITERATIONS; iteration++ {
		// computing gradient
		grad := make([]float64, OBJECT_LENGTH)
		for i := 0; i < BUTCH_SIZE; i++ {
			obj := objs[(iteration*BUTCH_SIZE+i)%len(objs)]
			if (model.getSuggestion(obj) == 0) == (obj.digit == digit0) {
				continue
			}
			sign := 1
			if obj.digit == digit0 {
				sign = -1
			}
			for j := 0; j < OBJECT_LENGTH; j++ {
				grad[j] += float64(sign * 2 * obj.x[j])
			}
		}
		// choosing best learning rate (step length)
		l, r := 0., .005
		m1, m2 := l+(r-l)*(fibs[FIB_SIZE-3]/fibs[FIB_SIZE-1]), l+(r-l)*(fibs[FIB_SIZE-2]/fibs[FIB_SIZE-1])
		f1 := evaluateResultDistinguishCorrectness(objs, model, grad, m1, digit0, digit1)
		f2 := evaluateResultDistinguishCorrectness(objs, model, grad, m2, digit0, digit1)
		for i := FIB_SIZE - 2; i > 1; i-- {
			if f1 >= f2 {
				l = m1
				m1 = m2
				f1 = f2
				m2 = l + (r-l)*(fibs[i-1]/fibs[i])
				f2 = evaluateResultDistinguishCorrectness(objs, model, grad, m2, digit0, digit1)
			} else {
				r = m2
				m2 = m1
				f2 = f1
				m1 = l + (r-l)*(fibs[i-2]/fibs[i])
				f1 = evaluateResultDistinguishCorrectness(objs, model, grad, m1, digit0, digit1)
			}
		}
		// changing weights of the perceptron
		for j := 0; j < OBJECT_LENGTH; j++ {
			model.w[j] += 0.001 * grad[j]
		}
		loggingDistinguisher(model, iteration, digit0, digit1, objs)
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
			randChan <- rnd.Intn(10)
		}
	}()

	//train particular perceptron
	//PLOT_ON = true
	digit := 5

	if false {
		for i := 0; i < 5; i++ {
			go trainParticularPerceptron(digit, true, false, randChan)
		}
		for {
		}
	}

	model := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	obj := &Object{}
	for i := 0; i < OBJECT_LENGTH; i++ {
		if model.w[i] >= 0 {
			obj.x[i] = int(100 + model.w[i]/10.)
		} else {
			obj.x[i] = int(100 + model.w[i]/10.)
		}
	}
	obj.digit = 21
	printObjectImage28x28(obj)

	// new concept
	//PLOT_ON = true
	//model := generateEmptyPerceptron()
	//learnToDistinguish(model, 3, 5)
	//obj := &Object{}
	//for i := 0; i < OBJECT_LENGTH; i++ {
	//	if model.w[i] >= 0 {
	//		obj.x[i] = int(100 + model.w[i]/10.)
	//	} else {
	//		obj.x[i] = int(100 + model.w[i]/10.)
	//	}
	//}
	//os.WriteFile(fmt.Sprintf("perceptron35.txt"), []byte(fmt.Sprintf("%v", model.w)), 0600)
	//obj.digit = 228
	//printObjectImage28x28(obj)

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
	//digit := 0
	//p := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	//fmt.Println(getPerceptronAccuracy(p, trainObjsGlob, digit))
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

func trainParticularPerceptron(digit int, createNew, empty bool, randChan chan int) {
	for {
		p := generatePerceptron(randChan)
		if empty {
			p = generateEmptyPerceptron()
		}
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

//func GetPrediction(image [OBJECT_LENGTH]int, perceptrons [10]*Perceptron) int {
//	obj := generateObject(image[:], true)
//	maxProbability := 0
//	matchingDigit := 10
//	for digit := 0; digit <= 9; digit++ {
//		curProbability := perceptrons[digit].getSuggestion(obj)
//		if curProbability > maxProbability {
//			maxProbability = curProbability
//			matchingDigit = digit
//		}
//	}
//	return matchingDigit
//}

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
