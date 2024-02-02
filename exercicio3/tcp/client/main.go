package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net"
	"os"
	"path/filepath"
)

func main() {
	//absolutePath, err := filepath.Abs("imgs/Apple.png")
	//absolutePath, err := filepath.Abs("imgs/Cake.png")
	absolutePath, err := filepath.Abs("imgs/Painting.png")
	//absolutePath, err := filepath.Abs("imgs/Star.png")
	if err != nil {
		fmt.Println("Error getting absolute path: ", err)
		return
	}

	img, err := openImage(absolutePath)
	if err != nil {
		fmt.Println("Error getting image: ", err)
		return
	}
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	imageBytes := buf.Bytes()

	// Send some data to the server
	_, err = conn.Write(imageBytes)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Display what the server responded

	imgGrey := bytesToImg(conn)
	path, err := filepath.Abs("greyscale.png")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	fg, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer fg.Close()

	//Save img to a File
	err = png.Encode(fg, imgGrey)
	// Close the connection
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

func bytesToImg(conn net.Conn) image.Image {
	img, _, err := image.Decode(conn)
	if err != nil {
		fmt.Println(err)
	}
	return img
}
