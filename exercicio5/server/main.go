package main

import (
  	"log"
  	"context"
  	"time"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"sync"

  	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
  if err != nil {
    log.Panicf("%s: %s", msg, err)
  }
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
	  "colored", // name
	  false,   // durable
	  false,   // delete when unused
	  false,   // exclusive
	  false,   // no-wait
	  nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	  )
	  failOnError(err, "Failed to register a consumer")
	  
	  
	  fmt.Println("Listening...")
	  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	  defer cancel()
	  for {
		for d := range msgs {
		 
		  
			img, err := bytesToImg(d.Body)
			failOnError(err, "Fail to convert colored bytes to img")
			
			
		  	pixels := imgToTensor(img)
		  	greyScale(&pixels)
		  	img = tensorToImg(pixels)

		  	buf := new(bytes.Buffer)
		  	err = png.Encode(buf, img)
			
			imageBytes := buf.Bytes()

		  	ch.PublishWithContext(ctx, 
				"",     // exchange
				d.ReplyTo, // routing key
				false,  // mandatory
				false,  // immediate
				
				amqp.Publishing {
					ContentType: "text/plain",
					Body:        imageBytes,
				},
			)
			failOnError(err, "Failed to publish a message")
		}
	  }
	  
	

}


func bytesToImg(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
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
