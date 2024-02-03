package main

import (
	"fmt"
	"net"
    "io"
	"os"
    "image"
    "image/color"
    "image/png"
    "bytes"
	"sync"
	"runtime"
    "path/filepath"
)
func main() {
	// Listen for incoming UDP packets on port 8080
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Server listening on", serverAddr)

	// Handle incoming UDP messages
	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
    bufferSize := 65000
    buffer := make([]byte, bufferSize) // Use a larger buffer for handling potential larger chunks

	// Receive the image data in chunks
	var imageData []byte
    var clientAddrFinal *net.UDPAddr
    i := 0 
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from udp: ", err)
            break
        }
		fmt.Printf("Received %d bytes from %s\n", n, clientAddr)
		imageData = append(imageData, buffer[:n]...)
        if n < bufferSize {
            clientAddrFinal = clientAddr;
			fmt.Println("Image completely received!")
            break
        }

		i += 1
		msg := fmt.Sprintf("Server response: Received chunk: %d", i)
		_, err = conn.WriteToUDP([]byte(msg), clientAddr)
		if err != nil{
			fmt.Println(err)
		}

	}

	// Process the received image data (e.g., save it to a file)
	img := bytesToImg(imageData)
    pixels := imgToTensor(img)
	greyScale(&pixels)
	img = tensorToImg(pixels)

	err := sendImage(conn, img, clientAddrFinal)
	if err != nil {
		fmt.Println("Error on send image: ", err)
	}

    // file, err := os.Create("greyscale.png")
    // if err != nil {
	// 	fmt.Println("Error creating file:", err)
	// 	return
	// }
	// defer file.Close()
	// err = png.Encode(file, img)
    // if err !=nil {
    //     fmt.Println("Error encoding png: ", err)
    // }

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

func bytesToImg(imgBytes []byte) image.Image {
 	img, _, err := image.Decode(bytes.NewReader(imgBytes))
 	if err != nil {
 		fmt.Println("bytes to img error: ", err)
 	}
 	return img
}

func sendImage(conn *net.UDPConn, img image.Image, addr *net.UDPAddr) error {
	imageBytes, err := imageToBytes(img)
	if err != nil {
		return err
	}
    err = saveImage(imageBytes, "temp.png")
	if err != nil {
        fmt.Println("Error creating temp file: ", err)
        return err
	}
    
    path, err := filepath.Abs("./temp.png")
    if err != nil {
        fmt.Println("Error getting filepath: ", err)
        return err
    }

    tempFile, err := os.Open(path)
    if err != nil {
        fmt.Println("Error opening file: ", err)
    }
    defer tempFile.Close()
    

    _, err = tempFile.Write(imageBytes)
    if err != nil {}


	buffer := make([]byte, 65000) // Use a larger buffer for handling potential larger chunks
	for {
		// Read a chunk of the image file
		n, err := tempFile.Read(buffer)
		if err != nil {
			if err == io.EOF {
                fmt.Println("EOF")
				break // End of file, exit the loop
			}
			return err
		}

		// Send the chunk to the server
		fmt.Println("Sending ", n, "bytes to client")
		_, err = conn.WriteToUDP(buffer[:n], addr)
		if err != nil {
			return err
		}
		if n < 65000 {
			fmt.Println("Image completely sent!")
            break
        }
		reply := make([]byte, 1024) 
		_, _, err = conn.ReadFromUDP(reply);
        if err != nil {
            fmt.Println("Error reading server response: ", err)
        }
        // fmt.Println(string(reply[:nReply]))
		
	}

	return nil
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

func imageToBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	// Converta para base64 ou use buf.Bytes() diretamente dependendo dos requisitos
	return buf.Bytes(), nil
}

