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
	// Resolve server address
	//absolutePath, err := filepath.Abs("imgs/Apple.png")
	absolutePath, err := filepath.Abs("imgs/Cake.png")
	// absolutePath, err := filepath.Abs("imgs/Painting.png")
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

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send a message to the server
	_, err = conn.Write(imageBytes)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	// Receive the response from the server
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error receiving response:", err)
		return
	}

	fmt.Println("Server response:", string(buffer[:n]))
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