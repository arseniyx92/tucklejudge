package imageProcessing

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"image"
	_ "image/png"
	"os"
)

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

func imageToGrayScale(img image.Image) image.Image {
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	return grayImg
}

func listRecognition(img image.Image) {
	img = imageToGrayScale(img)
	fmt.Println(img.Bounds())
	fmt.Println(img.At(100, 100))
	//for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
	//	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
	//		fmt.Print(img.At(x, y))
	//	}
	//	fmt.Println()
	//}
}

func main() {
	img, _ := getImageFromFile("/Users/arseniyx92/go/src/fieldsRecognition/testForm.png")
	listRecognition(img)
}
