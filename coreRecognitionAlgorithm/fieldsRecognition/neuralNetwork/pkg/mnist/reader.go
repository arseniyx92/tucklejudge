package mnist

import (
	"encoding/binary"
	"errors"
	"fieldsRecognition/neuralNetwork/pkg/cnn/maths"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
)

func ReadGrayImages(path string, limit int) ([]maths.Tensor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	defer f.Close()

	var header struct{ Magic, N, Rows, Cols int32 }
	if err := binary.Read(f, binary.BigEndian, &header); err != nil {
		return nil, errors.New("bad header")
	}
	if header.Magic != 2051 {
		return nil, errors.New("wrong magic number in header")
	}
	bytes := make([]byte, header.N*header.Rows*header.Cols)
	if _, err = io.ReadFull(f, bytes); err != nil {
		return nil, fmt.Errorf("%w, could not read full", err)
	}

	if limit > int(header.N) {
		return nil, fmt.Errorf("limit is larger than the amount of images in the dataset")
	}
	{
		byteIndex := (28 * 28) * 10

		bounds := image.Rect(0, 0, 28, 28)
		gray := image.NewGray(bounds)
		for x := 0; x < int(header.Rows); x++ {
			for y := 0; y < int(header.Cols); y++ {
				gray.SetGray(x, y, color.Gray{Y: bytes[byteIndex]})
				byteIndex++
			}
		}

		f, err := os.Create("output.png")
		if err != nil {
			log.Fatal(err)
		}
		png.Encode(f, gray)
		f.Close()
	}

	byteIndex := 0

	images := make([]maths.Tensor, limit)
	for i := 0; i < limit; i++ {
		pixels := make([]float64, header.Cols*header.Rows)
		values := make([]int, header.Cols*header.Rows)
		for j := 0; j < len(pixels); j++ {
			values[j] = int(bytes[byteIndex])
			byteIndex++
		}
		values = applyFancyAlterations(values)
		for j := 0; j < len(pixels); j++ {
			pixels[j] = (128.0 - float64(values[j])) / 255.0
		}
		images[i] = *maths.NewTensor([]int{int(header.Cols), int(header.Rows)}, pixels)
	}

	return images, nil
}

// arseniyx92's fields recognition integration
func applyFancyAlterations(init []int) []int {
	img := make([]int, 28*28)
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			img[i*28+j] = init[i+28*j]
		}
	}
	otsuThreshold(img)
	img = augmentWhiteFiguresThickness(img)
	img = augmentWhiteFiguresThickness(img)
	saveOnlyMainComponent(img)
	gaussianBlur(img, 2)
	centralizeMainComponent(img)
	saveOnlyMainComponent(img)
	return img
}

func augmentWhiteFiguresThickness(init []int) []int {
	img := make([]int, len(init))
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			img[x*28+y] = init[x*28+y]
			if init[x*28+y] == 255 {
				for dir := 0; dir < 4; dir++ {
					nx := x + dirX[dir]
					ny := y + dirY[dir]
					if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 {
						img[nx*28+ny] = 255
					}
				}
			}
		}
	}
	return img
}

func otsuThreshold(img []int) {
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
			colors[img[i*28+j]]++
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
			if img[i*28+j] <= bestInd {
				img[i*28+j] = 0
			} else {
				img[i*28+j] = 255
			}
		}
	}
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
			if used[i-lx][j-ly] != -1 || img[i*28+j] == 0 {
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
					if nx >= lx && nx <= rx && ny >= ly && ny <= ry && used[nx-lx][ny-ly] == -1 && img[nx*28+ny] > 0 {
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
				img[pos.first*28+pos.second] = 0
			}
		}
	}
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

func min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func gaussianBlur(img []int, kernel int) {
	lx := 0
	ly := 0
	rx := 27
	ry := 27
	pref := make([][]int, rx-lx+1)
	pref[0] = make([]int, ry-ly+1)
	pref[0][0] = img[0]
	for j := ly + 1; j <= ry; j++ {
		pref[0][j] = pref[0][j-1] + int(img[j])
	}
	for i := lx + 1; i <= rx; i++ {
		pref[i] = make([]int, ry-ly+1)
		pref[i][0] = pref[i-1][0] + int(img[i*28])
		for j := ly + 1; j <= ry; j++ {
			pref[i][j] = pref[i-1][j] + pref[i][j-1] - pref[i-1][j-1] + img[i*28+j]
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
			img[i*28+j] = int(float64(sum) / float64(cnt))
		}
	}
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
	}
	boundsX := IntPair{28, 0}
	boundsY := IntPair{28, 0}
	centerX := 0
	centerY := 0
	comp := components[0]
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
	scale := max(float64(boundsY.second-boundsY.first)/20., float64(boundsX.second-boundsX.first)/18.)
	boundsX.first = int(float64(boundsX.first)/scale) - centerX
	boundsX.second = int(float64(boundsX.second)/scale) - centerX
	boundsY.first = int(float64(boundsY.first)/scale) - centerY
	boundsY.second = int(float64(boundsY.second)/scale) - centerY
	for i := -28; i < 28; i++ {
		for j := -28; j < 28; j++ {
			shiftX := 14
			nx := i + shiftX + (boundsX.second-boundsX.first)/2 - boundsX.second
			ny := j + 24 - boundsY.second
			x := int(math.Round(scale * float64(i+centerX)))
			y := int(math.Round(scale * float64(j+centerY)))
			if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 && x >= 0 && x < 28 && y >= 0 && y < 28 {
				img[nx*28+ny] = cpy[x*28+y]
			}
		}
	}
}

func ReadLabels(path string, limit int) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	defer f.Close()

	var header struct{ Magic, N int32 }
	if err := binary.Read(f, binary.BigEndian, &header); err != nil {
		return nil, errors.New("bad header")
	}
	if header.Magic != 2049 {
		return nil, errors.New("wrong magic number in header")
	}

	bytes := make([]byte, header.N)
	if _, err = io.ReadFull(f, bytes); err != nil {
		return nil, fmt.Errorf("%w, could not read full", err)
	}
	if limit > int(header.N) {
		return nil, fmt.Errorf("limit is larger than the amount of labels in the dataset")
	}

	labels := make([]int, limit)
	for i := 0; i < limit; i++ {
		labels[i] = int(bytes[i])
	}

	return labels, nil
}

func LabelsToTensors(labels []int) []maths.Tensor {
	tensors := make([]maths.Tensor, len(labels))
	for i := 0; i < len(labels); i++ {
		values := make([]float64, 10)
		values[labels[i]] = 1
		t := maths.NewTensor([]int{10}, values) // 1d tensor of 10 values for 0-9
		tensors[i] = *t
	}
	return tensors
}
