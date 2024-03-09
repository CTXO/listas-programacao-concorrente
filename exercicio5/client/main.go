package main

import (
  "context"
  "log"
  "time"
  "os"
  "image"
  "image/png"
  "fmt"
  "path/filepath"
  "bytes"

  amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
	  log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	absolutePath, err := filepath.Abs("imgs/Painting.png")
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Sending image...")
	
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
	var receivedImage []byte
	for msg := range msgs {
		receivedImage = msg.Body
		break
	}
	
	greyscaleImg, err := bytesToImg(receivedImage)
	failOnError(err, "Failed to transform greyscale bytes to an image object")
	
	err = saveImage(greyscaleImg)
	failOnError(err, "Failed to save greyscale image")
	
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
