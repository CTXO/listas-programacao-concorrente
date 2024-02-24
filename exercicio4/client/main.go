package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/rpc"
	"os"
	"path/filepath"
)

type Args struct {
	Image []byte
}

func main() {
	//absolutePath, err := filepath.Abs("imgs/Apple.png")
	//absolutePath, err := filepath.Abs("imgs/Cake.png")
	absolutePath, err := filepath.Abs("imgs/Painting.png")
	//absolutePath, err := filepath.Abs("imgs/Star.png")
	if err != nil {
		fmt.Println("Error getting absolute path: ", err)
		return
	}

	//Getting the image from the client adress
	img, err := openImage(absolutePath)
	if err != nil {
		fmt.Println("Error getting image: ", err)
		return
	}

	//Converting the image to bytes so it can be sent
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		fmt.Println("Error encoding image: ", err)
		return
	}
	imageBytes := buf.Bytes()

	//Connecting with the server
	client, err := rpc.DialHTTP("tcp", "localhost"+":8080")
	if err != nil {
		fmt.Println(err)
	}

	//Requesting Greyscale Service
	var reply []byte
	args := Args{imageBytes}
	err = client.Call("GreyImage.GreyscaleRPC", args, &reply)
	if err != nil {
		fmt.Println(err)
	}

	//Saving Image received from Server
	newImg, err := bytesToImg(reply)
	if err != nil {
		fmt.Println(err)
	}
	saveImage(newImg)
}

func openImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	img, _, err := image.Decode(f)

	if err != nil {
		return nil, err
	}
	return img, nil
}

func bytesToImg(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(img image.Image) error {
	path, err := filepath.Abs("greyscale.png")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}

	fg, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer fg.Close()
	err = png.Encode(fg, img)
	if err != nil {
		fmt.Println("Error encoding file:", err)
		return err
	}

	fmt.Println("Image saved")
	return nil
}
