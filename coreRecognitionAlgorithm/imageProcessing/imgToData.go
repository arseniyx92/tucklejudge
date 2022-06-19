package main

import (
	"errors"
	"fmt"
	"github.com/gen2brain/go-fitz"
	"golang.org/x/exp/constraints"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"os"
)

const pixel = 1
const MINIMUM_COMPONENT_SIZE = 50

type IntPair struct {
	first, second int
}

type ErrorEmptyQueue int

func (err ErrorEmptyQueue) Error() string {
	return "Queue os empty for deleting or checking last/first element"
}

type Queue[T any] struct {
	stackIN  []T
	stackOUT []T
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

func printImage(filepath string, img image.Image) {
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}
	f, _ := os.Create(filepath)
	_ = png.Encode(f, img)
}

func getImagesFromPdf(filepath string) ([]image.Image, error) {
	doc, err := fitz.New(filepath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	images := make([]image.Image, doc.NumPage())
	for i := 0; i < doc.NumPage(); i++ {
		images[i], err = doc.Image(i)
		if err != nil {
			return nil, err
		}
	}
	return images, nil
}

func getImageFromFile(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func imageToGrayScale(img image.Image) *image.Gray {
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	return grayImg
}

func grayImageRevealSharpness(init *image.Gray) *image.Gray {
	gray := image.NewGray(init.Bounds())
	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			gray.Set(x, y, color.Gray{Y: init.GrayAt(x, y).Y * 200})
		}
	}
	return gray
}

func copyGray(init *image.Gray) *image.Gray {
	gray := image.NewGray(init.Bounds())
	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			gray.Set(x, y, color.Gray{Y: init.GrayAt(x, y).Y})
		}
	}
	return gray
}

func IntAbs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
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

func augmentContrast(img *image.Gray) {
	mean := 0
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			mean += int(img.GrayAt(i, j).Y)
		}
	}
	mean /= img.Bounds().Size().X * img.Bounds().Size().Y

	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			if img.GrayAt(i, j).Y < uint8(mean/2) {
				img.SetGray(i, j, color.Gray{Y: uint8(max(int(img.GrayAt(i, j).Y)-mean/2, 0))})
			} else {
				img.SetGray(i, j, color.Gray{Y: uint8(min(int(img.GrayAt(i, j).Y)+mean, 255))})
			}
		}
	}
}

func divideOnComponents(img *image.Gray, capacity int, colorGradientPermissionFactor float64) (components [][]IntPair) {
	used := make([][]int, img.Bounds().Dx()+1)
	for i := 0; i < img.Bounds().Dx()+1; i++ {
		used[i] = make([]int, img.Bounds().Dy()+1)
		for j := 0; j < img.Bounds().Dy()+1; j++ {
			used[i][j] = -1
		}
	}

	var compColorIntensity []float64

	var dirsX = []int{1, 0, 0, -1, 1, 1, -1, -1}
	var dirsY = []int{0, 1, -1, 0, 1, -1, 1, -1}

	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			if used[i][j] != -1 {
				continue
			}
			pressure := make(map[IntPair]int)
			q := Queue[IntPair]{}
			q.Push(IntPair{i, j})
			curComponent := len(components)
			used[i][j] = curComponent
			components = append(components, []IntPair{IntPair{i, j}})
			compColorIntensity = append(compColorIntensity, float64(img.GrayAt(i, j).Y))
			for q.Size() > 0 {
				cord, _ := q.Front()
				_ = q.Pop()
				for dir := 0; dir < 8; dir++ {
					nx, ny := cord.first+dirsX[dir], cord.second+dirsY[dir]
					if nx >= img.Bounds().Min.X && nx <= img.Bounds().Max.X && ny >= img.Bounds().Min.Y && ny <= img.Bounds().Max.Y && used[nx][ny] == -1 && math.Abs(float64(img.GrayAt(nx, ny).Y)-compColorIntensity[curComponent]/float64(len(components[curComponent]))) < colorGradientPermissionFactor {
						if len(components[curComponent]) < MINIMUM_COMPONENT_SIZE || 1+pressure[IntPair{nx, ny}] >= capacity {
							q.Push(IntPair{nx, ny})
							used[nx][ny] = curComponent
							compColorIntensity[curComponent] += float64(img.GrayAt(nx, ny).Y)
							components[curComponent] = append(components[curComponent], IntPair{nx, ny})
						} else {
							pressure[IntPair{nx, ny}]++
						}
					}
				}
			}
		}
	}
	return components
}

func SquaresRecognition(initColored image.Image) (squares [][]IntPair) {
	init := imageToGrayScale(initColored)
	augmentContrast(init)

	imageSIZE := init.Bounds().Size().X * init.Bounds().Size().Y
	components := divideOnComponents(init, 2*pixel, 10.*pixel)

	// preprocessing gray image to highlight corners
	img := image.NewGray(init.Bounds())
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			img.SetGray(i, j, color.Gray{Y: 0})
		}
	}
	for _, comp := range components {
		if len(comp) < int(float64(imageSIZE)*0.001) || len(comp) > int(float64(imageSIZE)*0.01) {
			continue
		}
		for _, pair := range comp {
			img.Set(pair.first, pair.second, color.Gray{Y: 255})
		}
	}

	components = divideOnComponents(img, 2*pixel, 10.*pixel)
	printImage("naked.png", img)

	// finding squares
	coloredPic := image.NewRGBA(img.Bounds())
	for _, comp := range components {
		var minX, maxX float64 = 1e9, 0
		var minY, maxY float64 = 1e9, 0
		for _, pix := range comp {
			minX = min(minX, float64(pix.first))
			maxX = max(maxX, float64(pix.first))
			minY = min(minY, float64(pix.second))
			maxY = max(maxY, float64(pix.second))
		}
		d := (maxX - minX + maxY - minY) / 2.
		S := d * d
		if math.Abs(S-float64(len(comp))) < float64(max(S, float64(len(comp))))*0.3 && len(comp) < int(float64(imageSIZE)*0.01) && len(comp) > int(float64(imageSIZE)*0.001) {
			for _, pair := range comp {
				coloredPic.Set(pair.first, pair.second, color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 0xff,
				})
			}
			squares = append(squares, comp)
		}
	}

	printImage("squares.png", coloredPic)
	return squares
}

func listRecognition(pic image.Image) {
	squares := SquaresRecognition(pic)
	if len(squares) != 3 {
		panic(errors.New(fmt.Sprintf("Picture has %d helping squares, should be 3", len(squares))))
	}
}

func main() {
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/harderInitialImage.jpg")
	img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/photo.jpeg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/testForm.png")
	listRecognition(img)
}
