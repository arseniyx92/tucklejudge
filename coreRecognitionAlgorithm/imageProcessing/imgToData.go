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
	"math/rand"
	"os"
	"sort"
)

const pixel = 1
const MINIMUM_COMPONENT_SIZE = 50

type IntPair struct {
	first, second int
}

type Point struct {
	x, y float64
}

type Vector struct {
	dx, dy, len float64
}

func (a Point) dist(b Point) float64 {
	dx := a.x - b.x
	dy := a.y - b.y
	return math.Sqrt(dx*dx + dy*dy)
}

func (a Point) vect(b Point) Vector {
	v := Vector{
		dx: b.x - a.x,
		dy: b.y - a.y,
	}
	v.len = math.Sqrt(v.dx*v.dx + v.dy*v.dy)
	return v
}

func (a Vector) dotProduct(b Vector) float64 {
	return a.dx*b.dx + a.dy*b.dy
}

func (a Vector) cosBetween(b Vector) float64 {
	return a.dotProduct(b) / (a.len * b.len)
}

func (a Vector) crossProduct(b Vector) float64 {
	return a.dx*b.dy - a.dy*b.dx
}

func (a Vector) sinBetween(b Vector) float64 {
	return a.crossProduct(b) / (a.len * b.len)
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
			if img.GrayAt(i, j).Y < max(uint8(mean/2), 100) {
				img.SetGray(i, j, color.Gray{Y: 0}) //uint8(max(int(img.GrayAt(i, j).Y)-mean/2, 0))})
			} else {
				img.SetGray(i, j, color.Gray{Y: 255}) //uint8(min(int(img.GrayAt(i, j).Y)+mean, 255))})
			}
		}
	}
}

func makeBW(img *image.Gray) {
	K := 100
	arr := make([]int, K)
	for iteration := 0; iteration < K; iteration++ {
		x := rand.Intn(img.Bounds().Max.X)
		y := rand.Intn(img.Bounds().Max.Y)
		arr[iteration] = int(img.GrayAt(x, y).Y)
	}
	sort.Ints(arr)
	colorSeparator := uint8(arr[10])
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			if img.GrayAt(i, j).Y >= colorSeparator {
				img.SetGray(i, j, color.Gray{Y: 255})
			} else {
				img.SetGray(i, j, color.Gray{Y: 0})
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

func squaresRecognition(initColored image.Image) (squares [][]IntPair) {
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
		if len(comp) > int(float64(imageSIZE)*0.01) {
			continue
		}
		for _, pair := range comp {
			img.Set(pair.first, pair.second, color.Gray{Y: 255})
		}
	}

	components = divideOnComponents(img, 2*pixel, 10.*pixel)
	//printImage("naked.png", img)

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
		if math.Abs(S-float64(len(comp))) < float64(max(S, float64(len(comp))))*0.3 && len(comp) < int(float64(imageSIZE)*0.01) {
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

	sort.Slice(squares, func(i, j int) bool {
		return len(squares[i]) < len(squares[j])
	})

	if len(squares) >= 3 {
		squares = squares[len(squares)-3:]
	}
	//printImage("squares.png", coloredPic)
	return squares
}

type Matrix struct {
	matrix [][]float64
}

func (m *Matrix) divideByNumber(x float64) {
	for _, vec := range m.matrix {
		for _, val := range vec {
			val /= x
		}
	}
}

func generateAffineMatrixFor2DCords(a1, a2, a3, b1, b2, b3 float64) (affine Matrix) {
	affine.matrix = make([][]float64, 2)
	affine.matrix[0] = make([]float64, 3)
	affine.matrix[0][0] = a1
	affine.matrix[0][1] = a2
	affine.matrix[0][2] = a3
	affine.matrix[1] = make([]float64, 3)
	affine.matrix[1][0] = b1
	affine.matrix[1][1] = b2
	affine.matrix[1][2] = b3
	return affine
}

func (m *Matrix) D2multToInt(arr []float64) (result []int) {
	result = make([]int, 2)
	for i := 0; i < len(m.matrix); i++ {
		for k := 0; k < len(arr); k++ {
			result[i] += int(m.matrix[i][k] * arr[k])
		}
	}
	return result
}

func Flip(init image.Image) *image.RGBA {
	img := image.NewRGBA(init.Bounds())
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			img.Set(init.Bounds().Size().X-i, init.Bounds().Size().Y-j, init.At(i, j))
		}
	}
	//printImage("affine.png", img)
	return img
}

func applyAffineTransformation(init image.Image, affine Matrix) *image.RGBA {
	img := image.NewRGBA(init.Bounds())
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			newCords := affine.D2multToInt([]float64{float64(i), float64(j), 1.})
			img.Set(i, j, init.At(newCords[0], newCords[1]))
		}
	}
	//printImage("affine.png", img)
	return img
}

func getInverseMatrix(A Matrix) Matrix {
	det := A.matrix[0][0]*A.matrix[1][1]*A.matrix[2][2] + A.matrix[0][1]*A.matrix[1][2]*A.matrix[2][0] + A.matrix[0][2]*A.matrix[2][1]*A.matrix[1][0] - A.matrix[0][2]*A.matrix[1][1]*A.matrix[2][0] - A.matrix[0][1]*A.matrix[1][0]*A.matrix[2][2] - A.matrix[0][0]*A.matrix[2][1]*A.matrix[1][2]
	Adjoint := Matrix{matrix: [][]float64{
		[]float64{A.matrix[1][1]*A.matrix[2][2] - A.matrix[1][2]*A.matrix[2][1], A.matrix[0][2]*A.matrix[2][1] - A.matrix[0][1]*A.matrix[2][2], A.matrix[0][1]*A.matrix[1][2] - A.matrix[0][2]*A.matrix[1][1]},
		[]float64{A.matrix[1][2]*A.matrix[2][0] - A.matrix[1][0]*A.matrix[2][2], A.matrix[0][0]*A.matrix[2][2] - A.matrix[0][2]*A.matrix[2][0], A.matrix[0][2]*A.matrix[1][0] - A.matrix[0][0]*A.matrix[1][2]},
		[]float64{A.matrix[1][0]*A.matrix[2][1] - A.matrix[1][1]*A.matrix[2][0], A.matrix[0][1]*A.matrix[2][0] - A.matrix[0][0]*A.matrix[2][1], A.matrix[0][0]*A.matrix[1][1] - A.matrix[0][1]*A.matrix[1][0]},
	}}
	Adjoint.divideByNumber(det)
	return Adjoint
}

func (m *Matrix) D3multToInt(arr []float64) (result []int) {
	dimension3 := make([]float64, 3)
	for i := 0; i < len(m.matrix); i++ {
		for k := 0; k < len(arr); k++ {
			dimension3[i] += m.matrix[i][k] * arr[k]
		}
	}
	result = make([]int, 2)
	result[0] = IntAbs(int(dimension3[0] / dimension3[2]))
	result[1] = IntAbs(int(dimension3[1] / dimension3[2]))
	return result
}

func perspectiveTransformation(init image.Image, x, y float64) *image.RGBA {
	m := Matrix{matrix: [][]float64{
		[]float64{1, 0, 0},
		[]float64{0, 1, 0},
		[]float64{0, -0.003 * y, 1},
	}}
	m = getInverseMatrix(m)
	img := image.NewRGBA(init.Bounds())
	for i := img.Bounds().Min.X; i <= img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.Y; j <= img.Bounds().Max.Y; j++ {
			newCords := m.D3multToInt([]float64{float64(i), float64(j), 1.})
			img.Set(i, j, init.At(newCords[0], newCords[1]))
		}
	}
	//printImage("PerspectiveNotAffine.png", img)
	return img
}

func photoToStandardDocument(init image.Image) (result *image.RGBA) {
	// making image twice smaller
	var pic *image.RGBA
	{
		dx := max(float64(init.Bounds().Size().X)/800., float64(init.Bounds().Size().Y)/800.)
		transformation := generateAffineMatrixFor2DCords(
			dx, 0, 0,
			0, dx, 0,
		)
		smallerImage := applyAffineTransformation(init, transformation)
		pic = image.NewRGBA(image.Rect(0, 0, int(float64(init.Bounds().Size().X)/dx), int(float64(init.Bounds().Size().Y)/dx)))
		for i := pic.Bounds().Min.X; i <= pic.Bounds().Max.X; i++ {
			for j := pic.Bounds().Min.Y; j <= pic.Bounds().Max.Y; j++ {
				pic.Set(i, j, smallerImage.At(i, j))
			}
		}
	}

	// looking for squares
	squares := squaresRecognition(pic)
	if len(squares) != 3 {
		panic(errors.New(fmt.Sprintf("Picture has %d helping squares, should be 3", len(squares))))
	}

	// finding squares centers
	centers := make([]Point, 3)
	for i, comp := range squares {
		meanX := 0.
		meanY := 0.
		for _, point := range comp {
			meanX += float64(point.first)
			meanY += float64(point.second)
		}
		meanX /= float64(len(comp))
		meanY /= float64(len(comp))
		centers[i] = Point{meanX, meanY}
	}

	// repositioning legs and hypotenuse
	a := centers[0].vect(centers[1])
	b := centers[0].vect(centers[2])
	c := centers[1].vect(centers[2])
	if centers[0].dist(centers[1]) > centers[0].dist(centers[2]) && centers[0].dist(centers[1]) > centers[1].dist(centers[2]) {
		a = centers[2].vect(centers[0])
		b = centers[2].vect(centers[1])
		c = centers[0].vect(centers[1])
	} else if centers[0].dist(centers[2]) > centers[1].dist(centers[2]) && centers[0].dist(centers[2]) > centers[0].dist(centers[1]) {
		a = centers[1].vect(centers[0])
		b = centers[1].vect(centers[2])
		c = centers[0].vect(centers[2])
	}

	// sorting sides a.len() < b.len() < c.len()
	if a.len > b.len && a.len < c.len {
		a, b = b, a
	} else if a.len < b.len && b.len > c.len {
		if c.len > a.len {
			b, c = c, b
		} else {
			a, b, c = c, a, b
		}
	} else if a.len > b.len && a.len > c.len {
		if c.len > b.len {
			a, b, c = b, c, a
		} else {
			a, b, c = c, b, a
		}
	}

	// making STANDARD larger canvas for image TODO: rectify coz it's redundant
	canvas := image.NewRGBA(image.Rect(0, 0, 800, 800))
	for i := pic.Bounds().Min.X; i <= pic.Bounds().Max.X; i++ {
		for j := pic.Bounds().Min.Y; j <= pic.Bounds().Max.Y; j++ {
			canvas.Set(i, j, pic.At(i, j))
		}
	}

	// rotating image to make OX parallel to one of the sides
	{
		OX := Vector{
			dx:  1,
			dy:  0,
			len: 1,
		}
		cos := math.Abs(a.cosBetween(OX))
		sin := math.Sqrt(1 - cos*cos)
		transformation := generateAffineMatrixFor2DCords(
			cos, sin, 0,
			-sin, cos, 0,
		)
		canvas = applyAffineTransformation(canvas, transformation)
	}

	// rotating image to make a and b perpendicular TODO: redundant if we have perspective transformation
	{
		sin := a.cosBetween(b)        // sin(90-A) = cos(A)
		cos := math.Sqrt(1 - sin*sin) // cos(90-A)
		tg := sin / cos
		transformation := generateAffineMatrixFor2DCords(
			1, tg, -800*tg,
			0, 1, 0,
		)
		canvas = applyAffineTransformation(canvas, transformation)
	}

	// perspective transformation
	{
		sin := a.cosBetween(b) // sin(90-A) = cos(A)
		//cos := math.Sqrt(1 - sin*sin) // cos(90-A)
		//tx := cos
		ty := sin
		canvas = perspectiveTransformation(canvas, 0, ty)
	}

	// if picture is flipped it should be reversed
	{
		if b.dy < 0 {
			canvas = Flip(canvas)
		}
	}

	// adjust picture so (0,0) pixel is the middle of the left above square
	{
		// finding squares again
		squares = squaresRecognition(canvas)
		if len(squares) != 3 {
			panic(errors.New(fmt.Sprintf("Picture has %d helping squares, should be 3", len(squares))))
		}
		centers := make([]Point, 3)
		for i, comp := range squares {
			meanX := 0.
			meanY := 0.
			for _, point := range comp {
				meanX += float64(point.first)
				meanY += float64(point.second)
			}
			meanX /= float64(len(comp))
			meanY /= float64(len(comp))
			centers[i] = Point{meanX, meanY}
		}
		zeroPoint := centers[0]
		if centers[1].x < zeroPoint.x {
			zeroPoint = centers[1]
		}
		if centers[2].x < zeroPoint.x {
			zeroPoint = centers[2]
		}
		transformation := generateAffineMatrixFor2DCords(
			1, 0, zeroPoint.x,
			0, 1, zeroPoint.y,
		)
		canvas = applyAffineTransformation(canvas, transformation)

		dx := 0
		dy := 0
		if zeroPoint == centers[1] {
			centers[0], centers[1] = centers[1], centers[0]
		} else if zeroPoint == centers[2] {
			centers[0], centers[2] = centers[2], centers[0]
		}
		dx = max(int(centers[2].x-centers[0].x), int(centers[1].x-centers[0].x)) + int(math.Sqrt(float64(len(squares[1]))))/2
		dy = max(int(centers[1].y-centers[0].y), int(centers[2].y-centers[0].y)) + int(math.Sqrt(float64(len(squares[2]))))/2

		result = image.NewRGBA(image.Rect(0, 0, dx, dy))
		for i := 0; i < dx; i++ {
			for j := 0; j < dy; j++ {
				result.Set(i, j, canvas.At(i, j))
			}
		}
	}
	return result
}

func formValuesProcessing(init image.Image) {
	img := imageToGrayScale(init)
	makeBW(img)
	printImage("kek.png", img)

	for i := img.Bounds().Min.X; i < img.Bounds().Max.X; i++ {
		for j := img.Bounds().Min.X; j < img.Bounds().Max.Y; j++ {

		}
	}
}

func main() {
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/insane.jpeg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/harderInitialImage.jpg")
	img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/photo.jpeg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/testForm.png")
	img = photoToStandardDocument(img)
	printImage("final.png", img)
	formValuesProcessing(img)
}