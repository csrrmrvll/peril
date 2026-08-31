package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/csrrmrvll/peril/internal/gamelogic"
	"github.com/csrrmrvll/peril/internal/pubsub"
	"github.com/csrrmrvll/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Starting Peril client...")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error during client welcome: %v", err)
	}
	fmt.Printf("Welcome, %s!\n", username)

	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("error declaring and binding queue: %v", err)
	}
	defer ch.Close()

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
