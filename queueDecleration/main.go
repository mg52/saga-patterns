package main

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://admin:123456@localhost:5672/")
	if err != nil {
		slog.Error("rabbitmq connection error", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("rabbitmq connection error", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"saga_transaction_topic", // name
		"topic",                  // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		slog.Error("rabbitmq ExchangeDeclare error", err)
		return
	}

	qServiceA, err := ch.QueueDeclare(
		"service_a_queue", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		true,              // no-wait
		nil,               // arguments
	)
	if err != nil {
		slog.Error("rabbitmq QueueDeclare error", err)
		return
	}

	qServiceB, err := ch.QueueDeclare(
		"service_b_queue", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		true,              // no-wait
		nil,               // arguments
	)
	if err != nil {
		slog.Error("rabbitmq QueueDeclare error", err)
		return
	}

	qError, err := ch.QueueDeclare(
		"error_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		true,          // no-wait
		nil,           // arguments
	)
	if err != nil {
		slog.Error("rabbitmq QueueDeclare error", err)
		return
	}

	err = ch.QueueBind(
		qServiceA.Name,           // queue name
		"serviceA",               // routing key
		"saga_transaction_topic", // exchange
		false,
		nil,
	)
	if err != nil {
		slog.Error("rabbitmq QueueBind error", err)
		return
	}

	err = ch.QueueBind(
		qServiceB.Name,           // queue name
		"serviceB",               // routing key
		"saga_transaction_topic", // exchange
		false,
		nil,
	)
	if err != nil {
		slog.Error("rabbitmq QueueBind error", err)
		return
	}

	err = ch.QueueBind(
		qError.Name,              // queue name
		"error",                  // routing key
		"saga_transaction_topic", // exchange
		false,
		nil,
	)
	if err != nil {
		slog.Error("rabbitmq QueueBind error", err)
		return
	}

	slog.Info("RabbitMQ declarations done.")
}
