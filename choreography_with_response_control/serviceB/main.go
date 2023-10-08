package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"github.com/mg52/saga-patterns/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Service B gets event from Service A and it gets error during rollout and sends it to error queue.
func main() {
	conn, err := amqp.Dial("amqp://admin:123456@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	msgs, err := ch.Consume(
		"service_b_queue", // queue
		"",                // consumer
		true,              // auto ack
		false,             // exclusive
		false,             // no local
		false,             // no wait
		nil,               // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var tr *pkg.Transaction
			json.Unmarshal(d.Body, &tr)
			if tr.Sender == "serviceA" && tr.ServiceAStatus {
				slog.Info("Service B started its process by request from service A ", slog.Int("transactionID", tr.TransactionID))
				time.Sleep(time.Second * 2)
				slog.Info("Service B rollout is completed unsuccessfully", slog.Int("transactionID", tr.TransactionID))

				tr.ServiceBStatus = false
				tr.Sender = "serviceB"
				slog.Info("Service B sends the event to Error Queue", slog.Int("transactionID", tr.TransactionID))

				b, _ := json.Marshal(tr)
				err = ch.PublishWithContext(context.TODO(),
					"saga_transaction_topic", // exchange
					"error",                  // routing key
					false,                    // mandatory
					false,                    // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        b,
					})
				failOnError(err, "Failed to publish a message")
			}
		}
	}()

	slog.Info("Service B is running. To exit press CTRL+C")
	<-forever
}
