package main

import (
	"fmt"
	"image"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
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

	logFilename := "cake.log"
	iterations := 1
	var totalElapsed time.Duration
	for i :=0; i < iterations; i++ {
		start := time.Now()
		conn, err := net.DialUDP("udp", nil, serverAddr)
		defer conn.Close()
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}

		
		err = sendImage(conn, absolutePath)
		if err != nil {
			fmt.Println("Error sending image:", err)
			return
		}

		imageData, err := receiveImage(conn)
		if err != nil {
			fmt.Println("Error receiving image: ", err)
			return
		}
		rttTime := time.Since(start)
		totalElapsed += rttTime
		
		
		err = saveImage(imageData, "greyscale.png")
		if err != nil{
			fmt.Println(err)
		}
		

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

func sendImage(conn *net.UDPConn, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

    bufferSize := 65000
	buffer := make([]byte, bufferSize) 
	reply := make([]byte, 1024) 
	for {
		n, err := file.Read(buffer)
		if err != nil {
            fmt.Println("Error reading from udp: ", err)
            break
		}
        fmt.Println("Sending ", n, " bytes to server")
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
	}

	return nil
}

func receiveImage(conn *net.UDPConn) ([]byte, error) {
    bufferSize := 65000
    buffer := make([]byte, bufferSize) 

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
    return imageData, nil
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