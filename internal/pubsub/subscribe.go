package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	channel, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		log.Fatalf("could not subscribe to %v: %v", queueName, err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	fmt.Println("Starting to consume messages from queue:", queue.Name)
	deliveries, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("could not start consuming messages: %v", err)
	}

	go func() {
		for d := range deliveries {
			var payload T
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("could not unmarshal message: %v", err)
				continue
			}
			fmt.Printf("Received message from queue %v: %s\n", queue.Name, d.Body)
			handler(payload)
			d.Ack(false)
		}
	}()
	return nil
}
