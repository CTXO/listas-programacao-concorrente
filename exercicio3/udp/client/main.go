package main

import (
	"fmt"
	"image"
	"net"
	"os"
	"path/filepath"
    // "io"
)

func main() {
	// Resolve server address
	 absolutePath, err := filepath.Abs("imgs/Apple.png")
	// absolutePath, err := filepath.Abs("imgs/Cake.png")
	// absolutePath, err := filepath.Abs("imgs/Painting.png")
	// absolutePath, err := filepath.Abs("imgs/Star.png")
	if err != nil {
		fmt.Println("Error getting absolute path: ", err)
		return
	}

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

    err = sendImage(conn, absolutePath)
    if err != nil {
        fmt.Println("Error sending image:", err)
        return
    }

    err = receiveImage(conn)
    if err != nil {
        fmt.Println("Error receiving image: ", err)
        return
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

func sendImage(conn *net.UDPConn, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

    bufferSize := 65000
	buffer := make([]byte, bufferSize) // Use a larger buffer for handling potential larger chunks
	reply := make([]byte, 1024) 
	for {
		// Read a chunk of the image file
		n, err := file.Read(buffer)
		if err != nil {
            fmt.Println("Error reading from udp: ", err)
            break
    }
        fmt.Println("Sending ", n, " bytes to server")
		// Send the chunk to the server
		_, err = conn.Write(buffer[:n])
		if err != nil {
			return err
		}

		if n < bufferSize {
            fmt.Println("Image completely sent!")
            break
        }

        _, _, err = conn.ReadFromUDP(reply);
        if err != nil {
            fmt.Println("Error reading server response: ", err)
        }
        // fmt.Println(string(reply[:nReply]))

	}

	return nil
}

func receiveImage(conn *net.UDPConn) error {
    bufferSize := 65000
    buffer := make([]byte, bufferSize) // Use a larger buffer for handling potential larger chunks

	// Receive the image data in chunks
	var imageData []byte
    i := 0
	for {
		n, _, err := conn.ReadFromUDP(buffer)
        fmt.Println("Received ", n, " bytes from server")
        if err != nil {
            fmt.Println("Error reading from udp: ", err)
            break
        }
		imageData = append(imageData, buffer[:n]...)
        if n < bufferSize {
            fmt.Println("Image completely received!")
            break
        }

        i += 1
		msg := fmt.Sprintf("Client response: Received chunk: %d", i)
		_, err = conn.Write([]byte(msg))
		if err != nil{
			fmt.Println(err)
		}
	}
	err := saveImage(imageData, "greyscale.png")
	if err != nil{
		fmt.Println(err)
        return err
	}
    return nil
}

func saveImage(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}