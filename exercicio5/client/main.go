package main

import (
  "context"
  "log"
  "time"

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
	
	// TODO: Change q and q2 variable names
	q, err := ch.QueueDeclare(
		"colored", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	
	failOnError(err, "Failed to declare a queue")

	q2, err := ch.QueueDeclare(
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
	
	body := "Image bytes here"
	err = ch.PublishWithContext(ctx,
	"",     // exchange
	q.Name, // routing key
	false,  // mandatory
	false,  // immediate
	amqp.Publishing {
		ContentType: "text/plain",
		Body:        []byte(body),
		ReplyTo: 	 q2.Name,
	})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent image bytes: %s\n", body)
	

	msgs, err := ch.Consume(
		q2.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	for d := range msgs {
		log.Printf("Received greyscale image back: %s", d.Body)
		break
	}
	
	


}