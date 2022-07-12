package main

import (
	"errors"
	"fieldsRecognition/AI"
	"fieldsRecognition/neuralNetwork/pkg/cnn"
	"fieldsRecognition/neuralNetwork/pkg/cnn/metrics"
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
const DOC_SIZE = 1200
const PERCEPTRON = false

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

type Dsu struct {
	size int
	p    []int
	d    []int
}

func (snm *Dsu) initializeDsu(n int) {
	snm.size = n
	snm.p = make([]int, n)
	snm.d = make([]int, n)
	for i := 0; i < n; i++ {
		snm.p[i] = i
	}
}

func (snm *Dsu) getParent(v int) (int, error) {
	if v >= snm.size || v < 0 {
		return 0, errors.New("index out of range in DSU 'getParent' inquiry")
	}
	if snm.p[v] == v {
		return v, nil
	}
	snm.p[v], _ = snm.getParent(snm.p[v])
	return snm.p[v], nil
}

func (snm *Dsu) merge(v, u int) (bool, error) {
	v, err := snm.getParent(v)
	if err != nil {
		return false, errors.New("index out of range in DSU 'merge' inquiry")
	}
	u, err = snm.getParent(u)
	if err != nil {
		return false, errors.New("index out of range in DSU 'merge' inquiry")
	}
	if u == v {
		return false, nil
	}
	if snm.d[v] > snm.d[u] {
		v, u = u, v
	}
	snm.p[v] = u
	if snm.d[v] == snm.d[u] {
		snm.d[u]++
	}
	return true, nil
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
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	return grayImg
}

func inverseGray(img *image.Gray) {
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			img.SetGray(x, y, color.Gray{Y: uint8(255 - img.GrayAt(x, y).Y)})
		}
	}
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
			result[i] += int(math.Round(m.matrix[i][k] * arr[k]))
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

func applyAffineTransformationOnGray(init *image.Gray, affine Matrix) *image.Gray {
	img := image.NewGray(init.Bounds())
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
	result[0] = IntAbs(int(math.Round(dimension3[0] / dimension3[2])))
	result[1] = IntAbs(int(math.Round(dimension3[1] / dimension3[2])))
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

func resizeImage(init image.Image, a, b int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, a, b))
	for i := init.Bounds().Min.X; i <= init.Bounds().Max.X; i++ {
		for j := init.Bounds().Min.Y; j <= init.Bounds().Max.Y; j++ {
			img.Set(i, j, init.At(i, j))
		}
	}
	return img
}

func photoToStandardDocument(init image.Image) (result *image.RGBA) {
	// making image twice smaller
	var pic *image.RGBA
	{
		dx := max(float64(init.Bounds().Size().X)/DOC_SIZE, float64(init.Bounds().Size().Y)/DOC_SIZE)
		if dx < 1 {
			init = resizeImage(init, DOC_SIZE, DOC_SIZE)
		}
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
	canvas := image.NewRGBA(image.Rect(0, 0, DOC_SIZE, DOC_SIZE))
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
			1, tg, -DOC_SIZE*tg,
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
		shift := int(min(min(centers[0].x-1, centers[0].y-1), math.Sqrt(float64(len(squares[0])))))
		if centers[1].x < zeroPoint.x {
			zeroPoint = centers[1]
			shift = int(min(min(centers[1].x-1, centers[1].y-1), math.Sqrt(float64(len(squares[1])))))
		}
		if centers[2].x < zeroPoint.x {
			zeroPoint = centers[2]
			shift = int(min(min(centers[2].x-1, centers[2].y-1), math.Sqrt(float64(len(squares[2])))))
		}
		transformation := generateAffineMatrixFor2DCords(
			1, 0, zeroPoint.x-float64(shift),
			0, 1, zeroPoint.y-float64(shift),
		)
		canvas = applyAffineTransformation(canvas, transformation)

		dx := 0
		dy := 0
		if zeroPoint == centers[1] {
			centers[0], centers[1] = centers[1], centers[0]
		} else if zeroPoint == centers[2] {
			centers[0], centers[2] = centers[2], centers[0]
		}
		dx = max(int(centers[2].x-centers[0].x), int(centers[1].x-centers[0].x)) + int(math.Sqrt(float64(len(squares[1]))))/2 + shift
		dy = max(int(centers[1].y-centers[0].y), int(centers[2].y-centers[0].y)) + int(math.Sqrt(float64(len(squares[2]))))/2 + shift

		result = image.NewRGBA(image.Rect(0, 0, dx, dy))
		for i := 0; i < dx; i++ {
			for j := 0; j < dy; j++ {
				result.Set(i, j, canvas.At(i, j))
			}
		}
	}
	return result
}

func augmentWhiteFiguresThickness(init *image.Gray) *image.Gray {
	img := image.NewGray(init.Bounds())
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	for x := init.Bounds().Min.X; x <= init.Bounds().Max.X; x++ {
		for y := init.Bounds().Min.Y; y <= init.Bounds().Max.Y; y++ {
			img.SetGray(x, y, init.GrayAt(x, y))
			if init.GrayAt(x, y).Y == 255 {
				for dir := 0; dir < 4; dir++ {
					nx := x + dirX[dir]
					ny := y + dirY[dir]
					if nx >= 0 && nx <= init.Bounds().Size().X && ny >= 0 && ny <= init.Bounds().Size().Y {
						img.SetGray(nx, ny, color.Gray{Y: 255})
					}
				}
			}
		}
	}
	return img
}

func lightUpWhiteABit(init *image.Gray) *image.Gray {
	img := image.NewGray(init.Bounds())
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	for x := init.Bounds().Min.X; x <= init.Bounds().Max.X; x++ {
		for y := init.Bounds().Min.Y; y <= init.Bounds().Max.Y; y++ {
			img.SetGray(x, y, init.GrayAt(x, y))
			if init.GrayAt(x, y).Y > 150 {
				img.SetGray(x, y, color.Gray{Y: 255})
				for dir := 0; dir < 4; dir++ {
					nx := x + dirX[dir]
					ny := y + dirY[dir]
					if nx >= 0 && nx <= init.Bounds().Size().X && ny >= 0 && ny <= init.Bounds().Size().Y {
						img.SetGray(nx, ny, color.Gray{Y: 255})
					}
				}
			}
		}
	}
	return img
}

func getWhiteComponents(img *image.Gray) (components [][]IntPair) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y

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
			if used[i-lx][j-ly] != -1 || img.GrayAt(i, j).Y == 0 {
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
					if nx >= lx && nx <= rx && ny >= ly && ny <= ry && used[nx-lx][ny-ly] == -1 && img.GrayAt(nx, ny).Y > 0 {
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

func saveOnlyMainComponent(img *image.Gray) {
	components := getWhiteComponents(img)
	maxSize := 0
	for _, comp := range components {
		maxSize = max(maxSize, len(comp))
	}
	for _, comp := range components {
		if maxSize != len(comp) {
			for _, pos := range comp {
				img.SetGray(pos.first, pos.second, color.Gray{Y: 0})
			}
		}
	}
}

func centralizeMainComponent(img *image.Gray, extend bool) {
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
	cpy := copyGray(img)
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
	if extend {
		scale = max(float64(boundsY.second-boundsY.first)/20., float64(boundsX.second-boundsX.first)/18.)
	}
	boundsX.first = int(float64(boundsX.first)/scale) - centerX
	boundsX.second = int(float64(boundsX.second)/scale) - centerX
	boundsY.first = int(float64(boundsY.first)/scale) - centerY
	boundsY.second = int(float64(boundsY.second)/scale) - centerY
	for i := -28; i < 28; i++ {
		for j := -28; j < 28; j++ {
			shiftX := 14
			if extend {
				shiftX = 15
			}
			img.SetGray(i+shiftX+(boundsX.second-boundsX.first)/2-boundsX.second, j+24-boundsY.second, cpy.GrayAt(int(math.Round(scale*float64(i+centerX))), int(math.Round(scale*float64(j+centerY)))))
		}
	}
	img.Rect = image.Rect(0, 0, 28, 28)
	printImage("kek0.png", img)
}

func imageFragmentTo28x28(rec image.Rectangle, init *image.Gray) (result [785]int) {
	img := image.NewGray(image.Rect(0, 0, max(rec.Size().X, 28), max(rec.Size().Y, 28)))
	for x := rec.Min.X; x <= rec.Max.X; x++ {
		for y := rec.Min.Y; y <= rec.Max.Y; y++ {
			img.SetGray(x-rec.Min.X, y-rec.Min.Y, init.GrayAt(x, y))
		}
	}
	//printImage("kek1.png", img)
	otsuThreshold(img)
	saveOnlyMainComponent(img)
	//img = augmentWhiteFiguresThickness(img)
	//img = augmentWhiteFiguresThickness(img)

	dx := float64(rec.Size().X) / 28.
	dy := float64(rec.Size().Y) / 28.
	magnifier := Matrix{matrix: [][]float64{
		[]float64{dx, 0, 0},
		[]float64{0, dy, 0},
	}}
	img = applyAffineTransformationOnGray(img, magnifier)
	centralizeMainComponent(img, false)
	img = augmentWhiteFiguresThickness(img)
	gaussianBlur(img, 2)
	img = lightUpWhiteABit(img)
	img = lightUpWhiteABit(img)
	gaussianBlur(img, 2)

	//{
	//	pic := image.NewGray(img.Bounds())
	//	for i := 0; i < 28; i++ {
	//		for j := 0; j < 28; j++ {
	//			pic.SetGray(i, j, img.GrayAt(i, j-1))
	//		}
	//	}
	//	printImage("kek1.png", pic)
	//}

	//printImage("kek1.png", img)
	result[0] = 1
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			result[1+x*28+y] = int(img.GrayAt(x, y).Y)
		}
	}
	//printArrayImage28x28("kek1.png", result)
	return result
}

func imageFragmentTo28x28cnnVersion(rec image.Rectangle, init *image.Gray, extended bool) (result []int) {
	img := image.NewGray(image.Rect(0, 0, max(rec.Size().X, 28), max(rec.Size().Y, 28)))
	for x := rec.Min.X; x <= rec.Max.X; x++ {
		for y := rec.Min.Y; y <= rec.Max.Y; y++ {
			img.SetGray(x-rec.Min.X, y-rec.Min.Y, init.GrayAt(x, y))
		}
	}
	otsuThreshold(img)
	saveOnlyMainComponent(img)
	img = augmentWhiteFiguresThickness(img)
	if extended {
		img = augmentWhiteFiguresThickness(img)
		img = augmentWhiteFiguresThickness(img)
	}
	gaussianBlur(img, 2)

	dx := float64(rec.Size().X) / 28.
	dy := float64(rec.Size().Y) / 28.
	magnifier := Matrix{matrix: [][]float64{
		[]float64{dx, 0, 0},
		[]float64{0, dy, 0},
	}}
	img = applyAffineTransformationOnGray(img, magnifier)
	centralizeMainComponent(img, extended)

	//img = augmentWhiteFiguresThickness(img)
	//gaussianBlur(img, 2)
	//img = lightUpWhiteABit(img)
	//img = lightUpWhiteABit(img)
	//gaussianBlur(img, 2)
	//printImage("kek0.png", img)

	//{
	//	pic := image.NewGray(img.Bounds())
	//	for i := 0; i < 28; i++ {
	//		for j := 0; j < 28; j++ {
	//			pic.SetGray(i, j, img.GrayAt(i, j-1))
	//		}
	//	}
	//	printImage("kek1.png", pic)
	//}

	//printImage("kek1.png", img)
	result = make([]int, 784)
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			result[x*28+y] = int(img.GrayAt(x, y).Y)
		}
	}
	return result
}

func imageFragmentTo28x28secondTry(rec image.Rectangle, init *image.Gray) (result [785]int) {
	pimg := image.NewGray(image.Rect(0, 0, max(rec.Size().X, 28), max(rec.Size().Y, 28)))
	for x := rec.Min.X; x <= rec.Max.X; x++ {
		for y := rec.Min.Y; y <= rec.Max.Y; y++ {
			pimg.SetGray(x-rec.Min.X, y-rec.Min.Y, init.GrayAt(x, y))
		}
	}
	otsuThreshold(pimg)
	saveOnlyMainComponent(pimg)
	//img = augmentWhiteFiguresThickness(img)
	//img = augmentWhiteFiguresThickness(img)

	dx := float64(rec.Size().X) / 28.
	dy := float64(rec.Size().Y) / 28.
	magnifier := Matrix{matrix: [][]float64{
		[]float64{dx, 0, 0},
		[]float64{0, dy, 0},
	}}
	pimg = applyAffineTransformationOnGray(pimg, magnifier)
	img := image.NewGray(image.Rect(0, 0, 28, 28))
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			img.SetGray(i, j, pimg.GrayAt(i, j))
		}
	}

	//img = augmentWhiteFiguresThickness(img)
	centralizeMainComponent(img, true)
	//img = augmentWhiteFiguresThickness(img)
	//gaussianBlur(img, 2)

	img = lightUpWhiteABit(img)
	gaussianBlur(img, 2)
	//printImage("kek0.png", img)

	//{
	//	pic := image.NewGray(img.Bounds())
	//	for i := 0; i < 28; i++ {
	//		for j := 0; j < 28; j++ {
	//			pic.SetGray(i, j, img.GrayAt(i, j-1))
	//		}
	//	}
	//	printImage("kek1.png", pic)
	//}

	//printImage("kek1.png", img)
	result[0] = 1
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			result[1+x*28+y] = int(img.GrayAt(x, y).Y)
		}
	}
	return result
}

func printArrayImage28x28(filepath string, x [785]int) {
	img := image.NewGray(image.Rect(0, 0, 28, 28))
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			img.SetGray(i, j, color.Gray{Y: uint8(x[1+i*28+j])})
		}
	}
	printImage(filepath, img)
}

func gaussianBlur(img *image.Gray, kernel int) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y
	pref := make([][]int, rx-lx+1)
	pref[0] = make([]int, ry-ly+1)
	pref[0][0] = int(img.GrayAt(0, 0).Y)
	for j := ly + 1; j <= ry; j++ {
		pref[0][j] = pref[0][j-1] + int(img.GrayAt(0, j).Y)
	}
	for i := lx + 1; i <= rx; i++ {
		pref[i] = make([]int, ry-ly+1)
		pref[i][0] = pref[i-1][0] + int(img.GrayAt(i, 0).Y)
		for j := ly + 1; j <= ry; j++ {
			pref[i][j] = pref[i-1][j] + pref[i][j-1] - pref[i-1][j-1] + int(img.GrayAt(i, j).Y)
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
			img.SetGray(i, j, color.Gray{Y: uint8(float64(sum) / float64(cnt))})
		}
	}
}

func diminishBlackGradient(img *image.Gray) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y
	old := image.NewGray(img.Bounds())
	var q Queue[IntPair]
	used := make([][]bool, rx-lx+1)
	for i := 0; i < rx-lx+1; i++ {
		used[i] = make([]bool, ry-ly+1)
	}
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			if img.GrayAt(i, j).Y <= 50 {
				q.Push(IntPair{i, j})
				used[i-lx][j-ly] = true
			}
			old.SetGray(i, j, img.GrayAt(i, j))
		}
	}
	dirsX := []int{-1, 0, 0, 1}
	dirsY := []int{0, -1, 1, 0}
	for q.Size() > 1 {
		pos, _ := q.Front()
		q.Pop()
		img.Set(pos.first, pos.second, color.Gray{Y: 0})
		for dir := 0; dir < 4; dir++ {
			nx := pos.first + dirsX[dir]
			ny := pos.second + dirsY[dir]
			if nx >= lx && nx <= rx && ny >= ly && ny <= ry && !used[nx-lx][ny-ly] && IntAbs(int(old.GrayAt(pos.first, pos.second).Y)-int(old.GrayAt(nx, ny).Y)) < 10 {
				q.Push(IntPair{nx, ny})
				used[nx-lx][ny-ly] = true
			}
		}
	}
}

func noiseDecrease(img *image.Gray) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y
	dirsX := []int{-1, 0, 0, 1, 1, 1, -1, -1}
	dirsY := []int{0, -1, 1, 0, 1, -1, 1, -1}
	for i := lx; i <= lx+(rx-lx+1)/8; i++ {
		for j := ly; j <= ry; j++ {
			if img.GrayAt(i, j).Y == 255 {
				cnt := 0
				for depth := 1; depth < 2; depth++ {
					for dir := 0; dir < 4; dir++ {
						nx := i + depth*dirsX[dir]
						ny := j + depth*dirsY[dir]
						if nx >= lx && nx <= rx && ny >= ly && ny <= ry && img.GrayAt(nx, ny).Y == 255 {
							cnt++
						}
					}
				}
				if cnt < 2 {
					img.SetGray(i, j, color.Gray{Y: 0})
				}
			}
		}
	}
}

func otsuThreshold(img *image.Gray) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y
	sz := (rx - lx + 1) * (ry - ly + 1)
	colors := make([]int, 256)
	pref := make([]int, 256)
	mean := make([]int, 256)
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			colors[img.GrayAt(i, j).Y]++
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
	bestInd := uint8(0)
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
			bestInd = uint8(i)
		}
	}
	for i := lx; i <= rx; i++ {
		for j := ly; j <= ry; j++ {
			if img.GrayAt(i, j).Y <= bestInd {
				img.SetGray(i, j, color.Gray{Y: 0})
			} else {
				img.SetGray(i, j, color.Gray{Y: 255})
			}
		}
	}
}

func fieldsRecognizer(img *image.Gray, kernelX, kernelY, minimumFieldSize int) (components [][]IntPair) {
	lx := img.Bounds().Min.X
	ly := img.Bounds().Min.Y
	rx := img.Bounds().Max.X
	ry := img.Bounds().Max.Y

	// examining white components
	var q Queue[IntPair]
	dirX := []int{-1, 0, 0, 1}
	dirY := []int{0, -1, 1, 0}
	used := make([][]int, rx-lx+1)
	for i := 0; i < rx-lx+1; i++ {
		used[i] = make([]int, ry-ly+1)
		for j := 0; j < ry-ly+1; j++ {
			used[i][j] = -1
		}
	}
	for j := ly; j <= ry; j++ {
		for i := lx; i <= rx; i++ {
			if used[i-lx][j-ly] != -1 || img.GrayAt(i, j).Y != 255 {
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
				for dir := 0; dir < 4; dir++ {
					nx := pos.first + dirX[dir]
					ny := pos.second + dirY[dir]
					if nx >= lx && nx <= rx && ny >= ly && ny <= ry && used[nx-lx][ny-ly] == -1 && img.GrayAt(nx, ny).Y == 255 {
						used[nx-lx][ny-ly] = comp
						q.Push(IntPair{nx, ny})
						components[comp] = append(components[comp], IntPair{nx, ny})
					}
				}
			}
		}
	}

	// merging near lying black components
	var snm Dsu
	snm.initializeDsu(len(components))
	for j := ly; j <= ry; j++ {
		accumulator := -1
		counter := 0
		for i := lx; i <= rx; i++ {
			for k := 0; k < kernelY && j-k >= ly; k++ {
				if i-kernelX >= 0 && used[i-kernelX][j-k] != -1 {
					counter--
				}
			}
			if counter == 0 {
				accumulator = -1
			}
			for k := 0; k < kernelY && j-k >= ly; k++ {
				if used[i][j-k] != -1 {
					currentComponent, _ := snm.getParent(used[i][j-k])
					if accumulator == -1 {
						accumulator = currentComponent
					}
					snm.merge(accumulator, currentComponent)
					counter++
				}
			}
			accumulator, _ = snm.getParent(accumulator)
		}
	}

	// calculating final components
	var resultingComps [][]IntPair
	reindexation := make([]int, len(components))
	for i := 0; i < len(components); i++ {
		reindexation[i] = -1
	}
	for i, _ := range components {
		ind, _ := snm.getParent(i)
		if reindexation[ind] == -1 {
			reindexation[ind] = len(resultingComps)
			resultingComps = append(resultingComps, []IntPair{})
		}
		resultingComps[reindexation[ind]] = append(resultingComps[reindexation[ind]], components[i]...)
	}

	// saving components with covering area not less than minimumFieldSize
	components = make([][]IntPair, 0)
	for _, comp := range resultingComps {
		minX := img.Bounds().Max.X
		maxX := 0
		minY := img.Bounds().Max.Y
		maxY := 0
		for _, v := range comp {
			minX = min(minX, v.first)
			maxX = max(maxX, v.first)
			minY = min(minY, v.second)
			maxY = max(maxY, v.second)
		}
		if (maxX-minX)*(maxY-minY) >= minimumFieldSize {
			components = append(components, comp)
		}
	}
	return components
}

var perceptrons = AI.InitializePerceptronMesh()
var NN *cnn.Network

func formValuesProcessing(init image.Image) (results []string) {
	if !PERCEPTRON {
		NN = cnn.New([]int{28, 28}, 0.005, &metrics.CrossEntropyLoss{})
		NN.LoadConvolutionLayer([]int{3, 3}, 8).
			AddMaxPoolingLayer(2, []int{2, 2}).
			LoadFullyConnectedLayer(10). // 0-9
			AddSoftmaxLayer()
	}
	// recognizing fields
	img := imageToGrayScale(init)
	inverseGray(img)
	diminishBlackGradient(img)
	otsuThreshold(img)
	noiseDecrease(img)
	printImage("kek.png", img)

	imgSize := (img.Bounds().Max.X - img.Bounds().Min.X + 1) * (img.Bounds().Max.Y - img.Bounds().Min.Y + 1)
	fields := fieldsRecognizer(img, int(float64(img.Bounds().Max.X-img.Bounds().Min.X+1)/85.), int(float64(img.Bounds().Max.X-img.Bounds().Min.Y+1)/100.), int(float64(imgSize)*0.01))

	// backup of a noisy image
	img = imageToGrayScale(init)
	inverseGray(img)

	for fieldID, field := range fields {
		currentValue := ""
		if len(field) == 0 {
			continue
		}
		minX := img.Bounds().Max.X
		maxX := 0
		minY := img.Bounds().Max.Y
		maxY := 0
		for _, v := range field {
			minX = min(minX, v.first)
			maxX = max(maxX, v.first)
			minY = min(minY, v.second)
			maxY = max(maxY, v.second)
		}
		blocks := 8
		borders := int(float64(imgSize) * 0.000003)
		if fieldID < 2 {
			blocks = 4
			borders = int(float64(imgSize) * 0.00001)
		}
		dx := int(math.Round(float64(maxX-minX) / float64(blocks)))
		for block := 0; block < blocks; block++ {
			start := minX + dx*block
			finish := minX + dx*(block+1)

			digit := 10
			if PERCEPTRON {
				digit = AI.GetPrediction(imageFragmentTo28x28(image.Rect(start+int(2*float64(borders)), minY+borders, finish-1, maxY-borders), img), perceptrons)
				if digit == 10 {
					digit = AI.GetPrediction(imageFragmentTo28x28secondTry(image.Rect(start+int(2*float64(borders)), minY+borders, finish-1, maxY-borders), img), perceptrons)
				}
			} else {
				if fieldID < 2 {
					digit = NN.GetDigitPredictionFromImageArray([]int{28, 28}, imageFragmentTo28x28cnnVersion(image.Rect(start+int(2*float64(borders)), minY+borders, finish-1, maxY-borders), img, true))
				} else {
					digit = NN.GetDigitPredictionFromImageArray([]int{28, 28}, imageFragmentTo28x28cnnVersion(image.Rect(start+int(2*float64(borders)), minY+borders, finish-1, maxY-borders), img, false))
				}
			}
			fmt.Print(digit, " ")
			digit = rand.Intn(50) + 200
			for i := start; i < finish; i++ {
				for j := minY; j <= maxY; j++ {
					img.SetGray(i, j, color.Gray{Y: uint8(digit)})
				}
			}
			currentValue += string(rune(digit + '0'))
			if block == 0 && fieldID == 5 {
				return
			}
		}
		results = append(results, currentValue)
		fmt.Println()
	}

	printImage("fields.png", img)
	return results
}

func BringTestResultsFromPhoto(filepath string) []string {
	img, _ := getImageFromFile(filepath)
	img = photoToStandardDocument(img)
	return formValuesProcessing(img)
}

func BringTestResultsFromPDFs(filepath string) (result [][]string) {
	imgs, _ := getImagesFromPdf(filepath)
	for _, img := range imgs {
		img = photoToStandardDocument(img)
		result = append(result, formValuesProcessing(img))
	}
	return result
}

func main() {
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/insane.jpeg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/harderInitialImage.jpg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/photo.jpeg")
	//img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/testForm.png")
	imgs, _ := getImagesFromPdf("/Users/arseniyx92/go/src/fieldsRecognition/Scan.pdf")
	img := imgs[0]
	img = photoToStandardDocument(img)
	printImage("final.png", img)
	formValuesProcessing(img)

	//init, _ := getImageFromFile("pic5.png") // TODO 5, 8, 9
	//img := imageToGrayScale(init)
	//perceptrons := AI.InitializePerceptronMesh()
	//fmt.Println(AI.GetPrediction(imageFragmentTo28x28secondTry(image.Rect(0, 0, 28, 28), img), perceptrons))

	//init, _ := getImageFromFile("kek0.png")
	//img := imageToGrayScale(init)
	//perceptron := AI.GetPerceptronFromFile("perceptron2.txt")
	//lol := imageFragmentTo28x28(image.Rect(0, 0, 28, 28), img)
	//obj := AI.GenerateObject(lol[:], true)
	//fmt.Println(perceptron.GetSuggestion(obj))
}
