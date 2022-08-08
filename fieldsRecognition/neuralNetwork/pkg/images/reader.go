package images

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func GrayScaleImageFromPath(path string) (*image.Gray, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open: %s, %w", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode image %w", err)
	}

	var (
		bounds = img.Bounds()
		gray   = image.NewGray(bounds)
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = img.At(x, y)
			gray.Set(x, y, rgba)
		}
	}

	return gray, nil
}
func Float64sToImage(data []float64) (image.Image, error) {
	b := make([]byte, len(data))
	for i := 0; i < len(b); i++ {
		b[i] = 128 - byte(data[i]*255.0)
	}

	return BytesToImage(b)
}

func BytesToImage(data []byte) (image.Image, error) {
	bounds := image.Rect(0, 0, 28, 28)
	gray := image.NewGray(bounds)

	fmt.Println(data)

	index := 0
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			gray.SetGray(i, j, color.Gray{Y: data[index]})
			fmt.Printf("%d ", data[index])
			index++
		}
	}
	return gray, nil
}
