package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

type Args struct {
	Image []byte
}

func main() {
	//absolutePath, err := filepath.Abs("imgs/Apple.png")
	// absolutePath, err := filepath.Abs("imgs/Cake.png")
	absolutePath, err := filepath.Abs("imgs/Painting.png")
	// absolutePath, err := filepath.Abs("imgs/Star.png")
	logFilename := "painting.log"
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

	iterations := 1
	var totalElapsed time.Duration
	for i := 0; i < iterations; i++ {
		client, err := rpc.DialHTTP("tcp", "localhost"+":8080")
		start := time.Now()
		//Connecting with the server
		if err != nil {
			fmt.Println(err)
		}

		//Requesting Greyscale Service
		var reply []byte
		args := Args{imageBytes}
		
		fmt.Println("Sending image to server...")
		err = client.Call("GreyImage.GreyscaleRPC", args, &reply)
		fmt.Println("Received greyscale image")
		if err != nil {
			fmt.Println(err)
		}
		rttTime := time.Since(start)
		totalElapsed += rttTime

		//Saving Image received from Server
		newImg, err := bytesToImg(reply)
		if err != nil {
			fmt.Println(err)
		}
		saveImage(newImg)
		fmt.Println("Saved greyscale image")

		//Saving the elapsed time of the iteration
		err = appendTimeToFile(logFilename, rttTime, "")
		if err != nil {
			fmt.Println("Error appending time to file: ", err)
		}
		client.Close()
	}

	//Getting the average elapsed time
	averageElapsed := totalElapsed / time.Duration(iterations)
	err = appendTimeToFile(logFilename, averageElapsed, "Average ")
	if err != nil {
		fmt.Println("Error appending time to file: ", err)
	}
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

	return nil
}

func appendTimeToFile(filename string, elapsed time.Duration, prefix string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	elapsedStr := fmt.Sprintf("%s%d\n", prefix, elapsed.Microseconds())

	if _, err := file.WriteString(elapsedStr); err != nil {
		return err
	}

	return nil
}
