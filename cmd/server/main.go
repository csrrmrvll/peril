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
	gamelogic.PrintServerHelp()

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}
	defer ch.Close()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		if words[0] == "pause" {
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
		} else if words[0] == "resume" {
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatalf("could not publish message: %v", err)
			}
		} else if words[0] == "quit" {
			log.Println("Quitting server...")
			break
		} else {
			log.Println("Unknown command")
		}
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
