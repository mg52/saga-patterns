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

// Service A gets event from client and it completes its rollout and sends to Service B queue.
// Also it gets event from Service B if Service B completed its process successfully.
func main() {
	conn, err := amqp.Dial("amqp://admin:123456@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	resMsgs, err := ch.Consume(
		"service_a_queue", // queue
		"",                // consumer
		true,              // auto ack
		false,             // exclusive
		false,             // no local
		false,             // no wait
		nil,               // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range resMsgs {
			var tr *pkg.Transaction
			json.Unmarshal(d.Body, &tr)
			if tr.Sender == "client" {
				slog.Info("Service A started its process by request from client", slog.Int("transactionID", tr.TransactionID))
				// Service A business simulation
				time.Sleep(time.Second * 2)
				slog.Info("Service A rollout is completed successfully", slog.Int("transactionID", tr.TransactionID))

				tr.ServiceAStatus = true
				tr.Sender = "serviceA"
				slog.Info("Service A sends the event to Service B", slog.Int("transactionID", tr.TransactionID))

				b, _ := json.Marshal(tr)
				err = ch.PublishWithContext(context.TODO(),
					"saga_transaction_topic", // exchange
					"serviceB",               // routing key
					false,                    // mandatory
					false,                    // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        b,
					})
				failOnError(err, "Failed to publish a message")
			} else if tr.Sender == "serviceB" {
				if tr.ServiceBStatus {
					slog.Info("All transactions completed successfully", slog.Int("transactionID", tr.TransactionID))
				}
			}
		}
	}()

	var forever chan struct{}

	slog.Info("Service A is running. To exit press CTRL+C")
	<-forever
}
