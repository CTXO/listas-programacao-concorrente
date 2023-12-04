package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	//absolutePath, err := filepath.Abs("./greyscale//imgs/Painting.png")
	//absolutePath, err := filepath.Abs("./greyscale//imgs/Apple.png")
	//absolutePath, err := filepath.Abs("./greyscale/imgs/Cake.png")
	absolutePath, err := filepath.Abs("./greyscale//imgs/Star.png")
	if err != nil {
		fmt.Println("Error getting absolute path: ", err)
		return
	}

	img, err := openImage(absolutePath)
	if err != nil {
		fmt.Println("Error getting image: ", err)
		return
	}

	pixels := img2tensor(img)

	greyScale(&pixels)
	//greyScaleConcV1(&pixels)
	//greyScaleConcV2(&pixels)

	tensor2img(pixels)
}

func openImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(f)

	if err != nil {
		return nil, err
	}
	return img, nil
}

// Converting image.YCbCr format to one which we can manipulate the pixels
func img2tensor(img image.Image) [][]color.Color {
	size := img.Bounds().Size()
	var pixels [][]color.Color

	for i := 0; i < size.X; i++ {
		var y []color.Color
		for j := 0; j < size.Y; j++ {
			y = append(y, img.At(i, j))
		}
		pixels = append(pixels, y)
	}
	return pixels
}

func tensor2img(pixels [][]color.Color) {
	rect := image.Rect(0, 0, len(pixels), len(pixels[0]))
	nImg := image.NewRGBA(rect)

	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[0]); y++ {
			q := pixels[x]
			if q == nil {
				continue
			}
			p := pixels[x][y]
			if p == nil {
				continue
			}
			original, ok := color.RGBAModel.Convert(p).(color.RGBA)
			if ok {
				nImg.Set(x, y, original)
			}
		}
	}

	path, err := filepath.Abs("greyscale.png")
	if err != nil {
		fmt.Println("Error getting absolute path in img2tensor", err)
		return
	}

	fg, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer fg.Close()

	// Salvar a imagem no arquivo
	err = png.Encode(fg, nImg)
}

func greyScaleConcV1(pixels *[][]color.Color) {
	ppixels := *pixels
	xLen := len(ppixels)
	yLen := len(ppixels[0])

	wg := sync.WaitGroup{}
	for x := 0; x < xLen; x++ {
		for y := 0; y < yLen; y++ {
			wg.Add(1)
			go func(x, y int) {
				pixel := ppixels[x][y]
				originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
				if !ok {
					fmt.Println("type conversion went wrong")
				}
				grey := uint8(float64(originalColor.R)*0.21 + float64(originalColor.G)*0.72 + float64(originalColor.B)*0.07)
				col := color.RGBA{
					grey,
					grey,
					grey,
					originalColor.A,
				}
				ppixels[x][y] = col
				wg.Done()
			}(x, y)

		}
	}
	wg.Wait()
	*pixels = ppixels
}

func greyScaleConcV2(pixels *[][]color.Color) {
	ppixels := *pixels
	xLen := len(ppixels)
	yLen := len(ppixels[0])

	numThreads := runtime.NumCPU()

	var wg sync.WaitGroup
	wg.Add(numThreads)

	processSection := func(startX, endX int) {
		defer wg.Done()
		for x := startX; x < endX; x++ {
			for y := 0; y < yLen; y++ {
				pixel := ppixels[x][y]
				originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
				if !ok {
					fmt.Println("type conversion went wrong")
				}
				grey := uint8(float64(originalColor.R)*0.21 + float64(originalColor.G)*0.72 + float64(originalColor.B)*0.07)
				col := color.RGBA{
					grey,
					grey,
					grey,
					originalColor.A,
				}
				ppixels[x][y] = col
			}
		}
	}

	for i := 0; i < numThreads; i++ {
		startX := (xLen * i) / numThreads
		endX := (xLen * (i + 1)) / numThreads
		go processSection(startX, endX)
	}

	wg.Wait()
}

func greyScale(pixels *[][]color.Color) {
	ppixels := *pixels
	xLen := len(ppixels)
	yLen := len(ppixels[0])

	for x := 0; x < xLen; x++ {
		for y := 0; y < yLen; y++ {
			pixel := ppixels[x][y]
			originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
			if !ok {
				fmt.Println("type conversion went wrong")
			}
			grey := uint8(float64(originalColor.R)*0.21 + float64(originalColor.G)*0.72 + float64(originalColor.B)*0.07)
			col := color.RGBA{
				grey,
				grey,
				grey,
				originalColor.A,
			}
			ppixels[x][y] = col
		}
	}
	*pixels = ppixels
}
