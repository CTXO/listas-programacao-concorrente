package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	absolutePath, err := filepath.Abs("imgs/Apple.png")
	// absolutePath, err := filepath.Abs("imgs/Cake.png")
	//absolutePath, err := filepath.Abs("imgs/Painting.png")
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

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	imageBytes := buf.Bytes()
	
	logFilename := "cake.log"
	iterations := 1
	var totalElapsed time.Duration
	for i := 0; i < iterations; i++{
		start := time.Now()
		conn, err := net.Dial("tcp", "localhost:8080")
		defer conn.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Sending image to server")
		_, err = conn.Write(imageBytes)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Waiting for image from server...")
		imgGrey := bytesToImg(conn)
		rttTime := time.Since(start)
		totalElapsed += rttTime
		fmt.Println("Received image from server")

		
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
		err = png.Encode(fg, imgGrey)
		
		fmt.Println("Image saved")
		err = appendTimeToFile(logFilename, rttTime, "")
		if err != nil {
			fmt.Println("Error appending time to file: ", err)
		}

	}
	averageElapsed := totalElapsed / time.Duration(iterations)
	err = appendTimeToFile(logFilename, averageElapsed, "Average ")
	if err != nil {
		fmt.Println("Error appending time to file: ", err)
	}
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

func appendTimeToFile(filename string, elapsed time.Duration, prefix string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	elapsedStr := fmt.Sprintf("%s Execution time: %s\n", prefix, elapsed)


	if _, err := file.WriteString(elapsedStr); err != nil {
		return err
	}

	return nil
}