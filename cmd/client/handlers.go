package main

import (
	"fmt"

	"github.com/csrrmrvll/peril/internal/gamelogic"
	"github.com/csrrmrvll/peril/internal/pubsub"
	"github.com/csrrmrvll/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWarRecognition(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(recog gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(recog)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			fmt.Printf("Not involved in war: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			fmt.Printf("No units available for war: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Printf("Opponent won the war: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			fmt.Printf("You won the war: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			fmt.Printf("War ended in a draw: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.Ack
		default:
			fmt.Printf("Error: unknown war outcome: attacker=%s, defender=%s\n", recog.Attacker.Username, recog.Defender.Username)
			return pubsub.NackDiscard
		}
	}
}

func handlerMove(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutcomeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			return pubsub.NackRequeue
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}
