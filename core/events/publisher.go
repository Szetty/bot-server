// Package events is responsible for notifying clients
package events

import (
	"botServer/web/model"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

// Subscriber represents an entity that will be notified with events
type Subscriber struct {
	Callback      *url.URL
	WebsocketConn *websocket.Conn
}

// StartGame is an intermediate structure for the StartGame event
type StartGame struct {
	GameID      uuid.UUID
	NextRound   int
	Players     []string
	Subscribers []Subscriber
}

// RoundFinished is an intermediate structure for the RoundFinished event
type RoundFinished struct {
	GameID        uuid.UUID
	CurrentRound  int
	NextRound     int
	PlayerResults []PlayerResult
	Winner        string
	Moves         map[string]interface{}
}

// GameFinished is an intermediate structure for the GameFinished event
type GameFinished struct {
	GameID        uuid.UUID
	PlayerResults []PlayerResult
	Winner        string
}

// PlayerResult holds the data specific to a player
// in the context of a RoundFinished or GameFinished event
type PlayerResult struct {
	Status     string
	Score      string
	Subscriber Subscriber
}

// PublishStartGame publishes the StartGame event
func PublishStartGame(startGame StartGame) {
	for _, subscriber := range startGame.Subscribers {
		publish(subscriber, model.Event{
			Type: "startGame",
			Body: model.StartGame{
				GameID:    startGame.GameID.String(),
				NextRound: startGame.NextRound,
				Players:   startGame.Players,
			},
		})
	}
}

// PublishRoundFinished publishes the RoundFinished event
func PublishRoundFinished(roundFinished RoundFinished) {
	moves := make(map[string]model.Move, len(roundFinished.Moves))
	for player, move := range roundFinished.Moves {
		moves[player] = model.Move{Value: move.(string)}
	}
	for _, playerResult := range roundFinished.PlayerResults {
		publish(playerResult.Subscriber, model.Event{
			Type: "roundFinished",
			Body: model.RoundFinished{
				GameID:       roundFinished.GameID.String(),
				CurrentRound: roundFinished.CurrentRound,
				NextRound:    roundFinished.NextRound,
				Score:        playerResult.Score,
				RoundResult: model.Result{
					Winner: roundFinished.Winner,
					Status: playerResult.Status,
					Moves:  moves,
				},
			},
		})
	}
}

// PublishGameFinished publishes the GameFinished event
func PublishGameFinished(gameFinished GameFinished) {
	for _, playerResult := range gameFinished.PlayerResults {
		publish(playerResult.Subscriber, model.Event{
			Type: "gameFinished",
			Body: model.GameFinished{
				GameID: gameFinished.GameID.String(),
				Score:  playerResult.Score,
				GameResult: model.Result{
					Status: playerResult.Status,
					Winner: gameFinished.Winner,
				},
			},
		})
	}
}

// PublishError publishes the Error event
func PublishError(subscribers []Subscriber, message string) {
	for _, subscriber := range subscribers {
		publish(subscriber, model.Event{
			Type: "error",
			Body: model.Error{
				Message: message,
			},
		})
	}
}

func publish(subscriber Subscriber, event model.Event) {
	go func() {
		time.Sleep(time.Second)
		if subscriber.WebsocketConn != nil {
			publishUsingWebsocket(subscriber.WebsocketConn, event)
			return
		}
		publishUsingHTTP(subscriber.Callback.String(), event)
	}()
}

func publishUsingHTTP(callback string, event interface{}) {
	body, err := json.Marshal(event)
	if err != nil {
		msg := fmt.Sprintf("publishing: could not encode %+v", event)
		logger.Error(errors.Wrap(err, msg))
	}
	resp, err := http.Post(callback, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Error(errors.Wrap(err, "publishing through HTTP failed"))
		return
	}
	if resp.StatusCode != 204 {
		msg := fmt.Sprintf("expecting status code 204 (No content) but got %d", resp.StatusCode)
		logger.Warning(errors.New("publishing: " + msg))
	}
}

func publishUsingWebsocket(conn *websocket.Conn, event interface{}) {
	err := conn.WriteJSON(event)
	if err != nil {
		logger.Error(errors.Wrap(err, "publishing through websocket failed"))
		return
	}
}
