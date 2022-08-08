package AI

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
	x     []int
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

var dd = 0

func generateObject(v []int, threshold bool) *Object { // first value - digit, then goes 28x28 picture
	obj := Object{
		digit: v[0],
		x:     make([]int, OBJECT_LENGTH),
	}
	obj.x[0] = 1 // bias weight
	for i := 1; i < OBJECT_LENGTH; i++ {
		obj.x[i] = v[i]
	}
	otsuThreshold(&obj)

	// augmenting white pixels count
	{
		getPureWhites := func() int {
			whiteCount := 0
			cpy := copyObject(obj)
			centralizeMainComponent(cpy.x)
			gaussianBlur(&cpy, 2)
			for i := 0; i < 28; i++ {
				for j := 0; j < 28; j++ {
					if cpy.x[i*28+j] == 255 {
						whiteCount++
					}
				}
			}
			return whiteCount
		}
		//for getPureWhites() > 200 {
		//	obj.x = diminishWhiteFiguresThickness(obj.x)
		//}
		for getPureWhites() < 100 {
			obj.x = augmentWhiteFiguresThickness(obj.x)
		}
	}
	centralizeMainComponent(obj.x)
	gaussianBlur(&obj, 2)
	//if obj.digit == 0 && dd == 2 {
	//	printObjectImage28x28(&obj)
	//	os.Exit(0)
	//} else if obj.digit == 0 {
	//	dd++
	//}
	return &obj
}

func copyObject(obj Object) Object {
	cpy := Object{
		digit: obj.digit,
		x:     make([]int, len(obj.x)),
	}
	for i, _ := range obj.x {
		cpy.x[i] = obj.x[i]
	}
	return cpy
}

type IntPair struct {
	first, second int
}

type Queue[T any] struct {
	stackIN  []T
	stackOUT []T
}

type ErrorEmptyQueue int

func (err ErrorEmptyQueue) Error() string {
	return "Queue os empty for deleting or checking last/first element"
}

func (q *Queue[T]) Size() int {
	return len(q.stackIN) + len(q.stackOUT)
}

func (q *Queue[T]) Push(val T) {
	q.stackIN = append(q.stackIN, val)
}

func (q *Queue[T]) Pop() error {
	if q.Size() == 0 {
		return ErrorEmptyQueue(0)
	}
	if len(q.stackOUT) == 0 {
		for len(q.stackIN) > 0 {
			q.stackOUT = append(q.stackOUT, q.stackIN[len(q.stackIN)-1])
			q.stackIN = q.stackIN[:len(q.stackIN)-1]
		}
	}
	q.stackOUT = q.stackOUT[:len(q.stackOUT)-1]
	return nil
}

func (q *Queue[T]) Back() (T, error) {
	if q.Size() == 0 {
		var result T
		return result, ErrorEmptyQueue(0)
	}
	if len(q.stackIN) == 0 {
		return q.stackOUT[0], nil
	} else {
		return q.stackIN[len(q.stackIN)-1], nil
	}
}

func (q *Queue[T]) Front() (T, error) {
	if q.Size() == 0 {
		var result T
		return result, ErrorEmptyQueue(0)
	}
	if len(q.stackOUT) == 0 {
		return q.stackIN[0], nil
	} else {
		return q.stackOUT[len(q.stackOUT)-1], nil
	}
}

func getWhiteComponents(img []int) (components [][]IntPair) {
	lx := 0
	ly := 0
	rx := 27
	ry := 27

	// examining black components
	var q Queue[IntPair]
	dirX := []int{-1, 0, 0, 1, -1, -1, 1, 1}
	dirY := []int{0, -1, 1, 0, -1, 1, -1, 1}
	used := make([][]int, rx-lx+1)
	for i := 0; i < rx-lx+1; i++ {
		used[i] = make([]int, ry-ly+1)
		for j := 0; j < ry-ly+1; j++ {
			used[i][j] = -1
		}
	}
	for j := ly; j <= ry; j++ {
		for i := lx; i <= rx; i++ {
			if used[i-lx][j-ly] != -1 || img[1+i*28+j] == 0 {
				continue
			}
			comp := len(components)
			components = append(components, []IntPair{})
			components[comp] = append(components[comp], IntPair{i, j})
			used[i-lx][j-ly] = comp
			q.Push(IntPair{i, j})
			for q.Size() > 0 {
				pos, _ := q.Front()
				q.Pop()
				for dir := 0; dir < 8; dir++ {
					nx := pos.first + dirX[dir]
					ny := pos.second + dirY[dir]
					if nx >= lx && nx <= rx && ny >= ly && ny <= ry && used[nx-lx][ny-ly] == -1 && img[1+nx*28+ny] > 0 {
						used[nx-lx][ny-ly] = comp
						q.Push(IntPair{nx, ny})
						components[comp] = append(components[comp], IntPair{nx, ny})
					}
				}
			}
		}
	}
	return components
}

func saveOnlyMainComponent(img []int) {
	components := getWhiteComponents(img)
	maxSize := 0
	for _, comp := range components {
		if maxSize < len(comp) {
			maxSize = len(comp)
		}
	}
	for _, comp := range components {
		if maxSize != len(comp) {
			for _, pos := range comp {
				img[1+pos.first*28+pos.second] = 0
			}
		}
	}
}

func diminishWhiteFiguresThickness(init []int) []int {
	img := make([]int, len(init))
	img[0] = init[0]
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			img[1+x*28+y] = init[1+x*28+y]
			if init[1+x*28+y] == 255 {
				cnt := 0
				for dir := 0; dir < 4; dir++ {
					nx := x + dirX[dir]
					ny := y + dirY[dir]
					if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 && init[1+nx*28+ny] == 255 {
						cnt++
					}
				}
				if cnt < 4 {
					img[1+x*28+y] = 0
				}
			}
		}
	}
	return img
}

func augmentWhiteFiguresThickness(init []int) []int {
	img := make([]int, len(init))
	img[0] = init[0]
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			img[1+x*28+y] = init[1+x*28+y]
			if init[1+x*28+y] == 255 {
				for dir := 0; dir < 4; dir++ {
					nx := x + dirX[dir]
					ny := y + dirY[dir]
					if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 {
						img[1+nx*28+ny] = 255
					}
				}
			}
		}
	}
	return img
}

func centralizeMainComponent(img []int) {
	components := getWhiteComponents(img)
	if len(components) > 1 {
		for i, comp := range components {
			if i != 0 {
				components[0] = append(components[0], comp...)
			}
		}
	} else if len(components) == 0 {
		return
	}
	cpy := make([]int, len(img))
	for i, _ := range img {
		cpy[i] = img[i]
		img[i] = 0
	}
	img[0] = cpy[0]
	boundsX := IntPair{28, 0}
	boundsY := IntPair{28, 0}
	comp := components[0]
	for _, pos := range comp {
		boundsX.first = min(boundsX.first, pos.first)
		boundsX.second = max(boundsX.second, pos.first)
		boundsY.first = min(boundsY.first, pos.second)
		boundsY.second = max(boundsY.second, pos.second)
	}
	center := IntPair{0, 0}
	result := int(1e18)
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			sum := 0
			for _, pos := range comp {
				sum += (pos.first-i)*(pos.first-i) + (pos.second-j)*(pos.second-j)
			}
			if sum < result {
				result = sum
				center = IntPair{i, j}
			}
		}
	}
	centerX := float64(center.first)
	centerY := float64(center.second)
	scale := max(float64(boundsY.second-boundsY.first)/28., float64(boundsX.second-boundsX.first)/30.)
	for i := -28; i < 28; i++ {
		for j := -28; j < 28; j++ {
			x := int(float64(i)*scale + centerX)
			y := int(float64(j)*scale + centerY)
			nx := i + 13
			ny := j + 26 + int((centerY-float64(boundsY.second))/scale)
			if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 && x >= 0 && x < 28 && y >= 0 && y < 28 {
				img[1+nx*28+ny] = cpy[1+x*28+y]
			}
		}
	}
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

func countMistakes(p *Perceptron, objs []*Object, digit int) (int, int) {
	counterOfCurrentDigit := 0
	totalDataSize := 0
	mistakes := 0
	for i := 0; i < len(objs); i++ {
		if objs[i].digit == digit {
			counterOfCurrentDigit++
			totalDataSize++
			suggestion := p.getSuggestion(objs[i])
			if (objs[i].digit == digit) != (suggestion >= 0) {
				mistakes++
			}
		}
	}
	for i := 0; i < len(objs); i++ {
		if objs[i].digit != digit {
			counterOfCurrentDigit--
			if counterOfCurrentDigit < 0 {
				break
			}
			totalDataSize++
			suggestion := p.getSuggestion(objs[i])
			if (objs[i].digit == digit) != (suggestion >= 0) {
				mistakes++
			}
		}
	}
	return mistakes, totalDataSize
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
	if iteration%3 == 0 {
		currentError, _ := countMistakes(model, trainObjsGlob, digit)
		if currentError < minError {
			saveFunc(model, digit)
			minError = currentError
			if iteration > 50 {
				printPercepronsImage(digit)
			}
			fmt.Println("NICE", currentError)
			//os.Exit(0)
		}
	}
	if PLOT_ON && iteration%ITERATIONS_TO_PLOT_SUPPLEMENTION == 0 {
		addPerceptronResultsToPlot(model, digit)
	}
	if iteration%ITERATIONS_TO_PRINT == 0 {
		currentError, objectsDataSize := countMistakes(model, trainObjsGlob, digit)
		fmt.Println("Iteration", iteration, " of digit:", digit, ")     error value: ", float64(currentError)/float64(objectsDataSize))
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

func printPercepronsImage(digit int) {
	model := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	obj := &Object{
		x: make([]int, OBJECT_LENGTH),
	}
	for i := 0; i < OBJECT_LENGTH; i++ {
		if model.w[i] >= 0 {
			obj.x[i] = min(255, int(100+model.w[i]/20.))
		} else {
			obj.x[i] = max(0, int(100+model.w[i]/20.))
		}
	}
	obj.digit = 22
	printObjectImage28x28(obj)
}

func main() {
	//train particular perceptron
	//PLOT_ON = true
	digit := 8 // todo train 5,6,7,8,9

	if false {
		parseObjectsFile()
		randChan := make(chan int)
		go func() {
			for {
				randChan <- rnd.Intn(10)
			}
		}()
		for i := 0; i < 5; i++ {
			go trainParticularPerceptron(digit, true, false, randChan)
		}
		for {
		}
	}

	printPercepronsImage(digit)

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
	//parseObjectsFile()
	//p := getPerceptronFromFile(fmt.Sprintf("perceptron%d.txt", digit))
	//fmt.Println(getPerceptronAccuracy(p, trainObjsGlob, digit))
}

var testingDataCorrectnessPlot []float64
var trainingDataCorrectnessPlot []float64

func addPerceptronResultsToPlot(p *Perceptron, digit int) {
	mistakes, dataSize := countMistakes(p, trainObjsGlob, digit)
	trainingDataCorrectnessPlot = append(trainingDataCorrectnessPlot, 1.-float64(mistakes)/float64(dataSize))
	mistakes, dataSize = countMistakes(p, testObjsGlob, digit)
	testingDataCorrectnessPlot = append(testingDataCorrectnessPlot, 1.-float64(mistakes)/float64(dataSize))
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
		perceptrons[digit] = getPerceptronFromFile(fmt.Sprintf("fieldsRecognition/perceptrons/perceptron%d.txt", digit))
	}
	return perceptrons
}

func GetPrediction(init []int, perceptrons [10]*Perceptron) []int {
	if init == nil {
		return []int{10}
	}
	obj := &Object{
		x: append([]int{1}, init...),
	}

	probabilities := make([]float64, 10)
	for digit := 0; digit < 10; digit++ {
		probabilities[digit] = perceptrons[digit].getSuggestion(obj)
	}
	digits := make([]int, 10)
	for i := range digits {
		digits[i] = i
	}
	sort.Slice(digits, func(i, j int) bool {
		return probabilities[digits[i]] > probabilities[digits[j]]
	})
	return digits
}

type Pair[T constraints.Ordered, C constraints.Ordered] struct {
	first  T
	second C
}

func GetAnalyticsPrediction(dims, img []int) []int {
	if img == nil {
		return []int{10}
	}
	// figuring out number of white zones and computing convolutional map
	whiteZonesArr := make([]int, 0)
	whitePixels := 0.
	convMap := make([]float64, 4)
	used := make([]bool, len(img))
	for i := 0; i < dims[0]; i++ {
		for j := 0; j < dims[1]; j++ {
			ind := dims[1]*i + j
			// conv map
			if img[ind] > 0 {
				convPos := 0
				if i > dims[0]/2 {
					convPos += 2
				}
				if j > dims[1]/2 {
					convPos++
				}
				convMap[convPos]++
				whitePixels++
			} else if !used[ind] { // white zones
				zoneSize := 1
				used[ind] = true
				var q Queue[IntPair]
				q.Push(IntPair{i, j})
				dirX := []int{-1, 0, 0, 1, -1, -1, 1, 1}
				dirY := []int{0, -1, 1, 0, -1, 1, -1, 1}
				for q.Size() > 0 {
					pos, _ := q.Front()
					x, y := pos.first, pos.second
					q.Pop()
					for dir := 0; dir < 8; dir++ {
						nx := x + dirX[dir]
						ny := y + dirY[dir]
						if nx >= 0 && ny >= 0 && nx < dims[0] && ny < dims[1] && !used[nx*dims[1]+ny] && img[nx*dims[1]+ny] == 0 {
							zoneSize++
							used[nx*dims[1]+ny] = true
							q.Push(IntPair{nx, ny})
						}
					}
				}
				whiteZonesArr = append(whiteZonesArr, zoneSize)
			}
		}
	}
	whiteZones := 0
	for _, sz := range whiteZonesArr {
		if sz > 1 {
			whiteZones++
		}
	}
	whitePixels = max(whitePixels, 1.) // here error has been silenced
	for i := range convMap {
		convMap[i] /= whitePixels
	}
	// getting prediction array from convolved map[4]
	generic := make([][]float64, 10)
	generic[0] = []float64{0.25, 0.25, 0.25, 0.25}
	generic[1] = []float64{0.2, 0.5, 0.3, 0}
	generic[2] = []float64{0.2, 0.3, 0.19, 0.31}
	generic[3] = []float64{0.15, 0.35, 0.35, 0.15}
	generic[4] = []float64{0.33, 0.33, 0.33, 0}
	generic[5] = []float64{0.36, 0.15, 0.35, 0.14}
	generic[6] = []float64{0.29, 0.12, 0.29, 0.3}
	generic[7] = []float64{0.19, 0.49, 0.02, 0.3}
	generic[8] = []float64{0.25, 0.25, 0.25, 0.25}
	generic[9] = []float64{0.29, 0.3, 0.29, 0.12}
	predictions := make([]Pair[float64, int], 10)
	for i := 0; i < 10; i++ {
		predictions[i] = Pair[float64, int]{0., i}
		for j := range convMap {
			predictions[i].first += math.Abs(generic[i][j] - convMap[j])
		}
	}
	sort.Slice(predictions, func(i, j int) bool {
		if predictions[i].first < predictions[j].first {
			return true
		}
		return false
	})
	// figuring out the most likable result (and returning a list with sorted unlikely predictions)
	// and transforming predictions to the []int form
	results := make([]int, 0)
	if whiteZones == 1 {
		for _, cur := range predictions {
			if cur.second != 6 && cur.second != 9 && cur.second != 0 && cur.second != 8 {
				results = append(results, cur.second)
			}
		}
		for _, cur := range predictions {
			if cur.second == 6 || cur.second == 9 || cur.second == 0 {
				results = append(results, cur.second)
			}
		}
		results = append(results, 8)
	} else if whiteZones == 2 {
		for _, cur := range predictions {
			if cur.second == 6 || cur.second == 9 || cur.second == 0 {
				results = append(results, cur.second)
			}
		}
		results = append(results, 8)
	} else if whiteZones == 3 {
		results = append(results, 8)
	} else {
		results = make([]int, 10)
		for i := range predictions {
			results[i] = predictions[i].second
		}
	}
	return results
}

func getPerceptronAccuracy(model *Perceptron, objs []*Object, digit int) float64 {
	mistakes := 0
	count := 0
	for _, obj := range objs {
		x := model.getSuggestion(obj)
		if digit != obj.digit {
			count++
			if (x >= 0) != (digit == obj.digit) {
				mistakes++
			}
		}
	}
	return 1. - (float64(mistakes) / float64(count))
}
