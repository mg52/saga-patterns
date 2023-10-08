package main

import (
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

// Response Control Service controls the microservices' error events
// in order to start rollback process or triggering rollback events to related microservices.
func main() {
	conn, err := amqp.Dial("amqp://admin:123456@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	resMsgs, err := ch.Consume(
		"error_queue", // queue
		"",            // consumer
		true,          // auto ack
		false,         // exclusive
		false,         // no local
		false,         // no wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range resMsgs {
			var tr *pkg.Transaction
			json.Unmarshal(d.Body, &tr)
			if tr.Sender == "serviceA" {
				slog.Info("Got error from Service A, starting rollback process for Service B", slog.Int("transactionID", tr.TransactionID))
				// Rollback simulation
				time.Sleep(time.Second * 2)
				slog.Info("Rollback for Service B is completed successfully", slog.Int("transactionID", tr.TransactionID))

			} else if tr.Sender == "serviceB" {
				slog.Info("Got error from Service B, starting rollback process for Service A", slog.Int("transactionID", tr.TransactionID))
				// Rollback simulation
				time.Sleep(time.Second * 2)
				slog.Info("Rollback for Service A is completed successfully", slog.Int("transactionID", tr.TransactionID))

			}
		}
	}()

	var forever chan struct{}

	slog.Info("Response Control Service is running. To exit press CTRL+C")
	<-forever
}
