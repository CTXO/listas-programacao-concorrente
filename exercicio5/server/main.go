package main

import (
  "log"
  "context"
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
	  
	  
	  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	  defer cancel()
	  for {
		for d := range msgs {
		  log.Printf("Received a message: %s", d.Body)
		  body := "Greyscale image from server"

		  ch.PublishWithContext(ctx, 
				"",     // exchange
				d.ReplyTo, // routing key
				false,  // mandatory
				false,  // immediate
				
				amqp.Publishing {
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)
			failOnError(err, "Failed to publish a message")
		}
	  }
	  
	

}