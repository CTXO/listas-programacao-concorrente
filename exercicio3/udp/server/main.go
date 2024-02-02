package main

import (
	"fmt"
	"net"
    "image"
)

func main() {
	// Listen for incoming UDP packets on port 8080
	serverAddr, err := net.ResolveUDPAddr("udp", ":8080")
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
	buffer := make([]byte, 1024)

	// Read from UDP connection
	n, clientAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading from UDP:", err)
		return
	}

    img := bytesToImg(conn)
    fmt.Println(img)

	// Echo the message back to the client
	_, err = conn.WriteToUDP(buffer[:n], clientAddr)
	if err != nil {
		fmt.Println("Error writing to UDP:", err)
		return
	}
}

func bytesToImg(conn net.Conn) image.Image {
	img, _, err := image.Decode(conn)
	if err != nil {
		fmt.Println(err)
	}
	return img
}
