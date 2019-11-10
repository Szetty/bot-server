package core

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
)

// ConnectRequest is the input for the Connect operation in the core
type ConnectRequest struct {
	GameName      string
	Token         string
	NoOfPlayers   int
	PlayerName    string
	EventCallback *url.URL
	TotalRounds   int
}

// ConnectResponse is the output for the Connect operation in the core
type ConnectResponse struct {
	GameID uuid.UUID
	Player Player
	Rounds int
}

// PlayRequest is the input for the Play operation in the core
type PlayRequest struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
	Round    int
	Move     interface{}
}

// PlayResponse is the output for the Play operation in the core
type PlayResponse struct {
	Round       int
	PlayersMove []string
}

// Player holds data of a player in the context of a Connect operation
type Player struct {
	ID            uuid.UUID
	Name          string
	EventCallback *url.URL
	WebsocketConn *websocket.Conn
	currentMove   interface{}
	score         int
}

func getOrCreatePlayer(playerName string, eventCallback *url.URL) *Player {
	if playerName == "" {
		playerName = "Player" + strconv.FormatInt(playerNameNr, 10)
		playerNameNr++
	}
	return &Player{
		ID:            uuid.New(),
		Name:          playerName,
		EventCallback: eventCallback,
		score:         0,
		currentMove:   nil,
	}
}
