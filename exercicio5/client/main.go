package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
	  log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	absolutePath, err := filepath.Abs("imgs/Apple.png")
	logFilename := "apple.log"
	failOnError(err, "Failed to get absolutePath")
	
	img, err := openImage(absolutePath)
	failOnError(err, "Failed to open image")
	
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	failOnError(err, "Error encoding image")

	imageBytes := buf.Bytes()
	

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	
	// TODO: Change coloredQueue and q2 variable names
	coloredQueue, err := ch.QueueDeclare(
		"colored", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	greyscaleQueue, err := ch.QueueDeclare(
		"greyscale", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	  )
	  failOnError(err, "Failed to declare queue2")
	
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	// fmt.Println("Sending image...")
	

	iterations := 50
	var totalElapsed time.Duration

	msgs, err := ch.Consume(
		greyscaleQueue.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for i := 0; i < iterations; i++ {
		start := time.Now()
		err = ch.PublishWithContext(ctx,
		"",     // exchange
		coloredQueue.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        imageBytes,
			ReplyTo: 	 greyscaleQueue.Name,
		})
		failOnError(err, "Failed to publish a message")

		var receivedImage []byte
		msg := <-msgs
		receivedImage = msg.Body
		rttTime := time.Since(start)
		totalElapsed += rttTime	
		
		greyscaleImg, err := bytesToImg(receivedImage)
		failOnError(err, "Failed to transform greyscale bytes to an image object")
		
		err = saveImage(greyscaleImg)
		failOnError(err, "Failed to save greyscale image")

		err = appendTimeToFile(logFilename, rttTime, "")
		failOnError(err, "Failed to append time to file")
	}

	_, err = ch.QueueDelete(
		greyscaleQueue.Name, 
		false,   // ifUnused
		false,   // ifEmpty
		false,   // noWait
	)
	failOnError(err, "Failed to delete queue");

	averageElapsed := totalElapsed / time.Duration(iterations)
	err = appendTimeToFile(logFilename, averageElapsed, "Average ")
	failOnError(err, "Failed to append average time to file")

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
