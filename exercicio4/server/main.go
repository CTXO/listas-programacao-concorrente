package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"net/http"
	"net/rpc"
	"runtime"
	"sync"
)

type Args struct {
	Image []byte
}

type GreyImage []byte

func (g *GreyImage) GreyscaleRPC(args *Args, reply *[]byte) error {
	img, err := bytesToImg(args.Image)
	if err != nil {
		fmt.Println(err)
	}

	pixels := imgToTensor(img)
	greyScale(&pixels)
	img = tensorToImg(pixels)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	*reply = buf.Bytes()
	return nil
}

func main() {
	greyImage := new(GreyImage)
	rpc.Register(greyImage)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	}
	http.Serve(l, nil)
}

func bytesToImg(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Converting image.YCbCr format to one which we can manipulate the pixels
func imgToTensor(img image.Image) [][]color.Color {
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

func tensorToImg(pixels [][]color.Color) image.Image {
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

	return nImg

}

func greyScale(pixels *[][]color.Color) {
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
