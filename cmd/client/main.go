package main

import (
	"fmt"
	"log"

	"github.com/csrrmrvll/peril/internal/gamelogic"
	"github.com/csrrmrvll/peril/internal/pubsub"
	"github.com/csrrmrvll/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gs := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			// fmt.Println("Publishing spawned game state")
			err := gs.CommandSpawn(words)
			if err != nil {
				log.Printf("could not execute spawn command: %v", err)
			}
			// err = pubsub.PublishJSON(
			// 	conn.Channel(),
			// 	routing.ExchangePerilDirect,
			// 	routing.PauseKey,
			// 	routing.PlayingState{
			// 		IsPaused: true,
			// 	},
			// )
			// if err != nil {
			// 	log.Printf("could not publish time: %v", err)
			// }
		case "move":
			fmt.Println("Publishing moved game state")
			move, err := gs.CommandMove(words)
			if err != nil {
				log.Printf("could not execute move command: %v", err)
			} else {
				fmt.Printf("Move executed: %v\n", move)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("unknown command: %v\n", words[0])
			continue
		}
	}
}
