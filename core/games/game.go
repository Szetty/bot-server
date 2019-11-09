// Package games contains all the supported games
package games

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Status represents the outcome of a round or game
type Status string

const (
	// WIN - positive outcome
	WIN Status = "win"
	// LOSE - negative outcome
	LOSE Status = "lose"
	// DRAW - neutral outcome
	DRAW Status = "draw"
)

// GameType describes a specific game
type GameType interface {
	Validate(int) bool
	GetDefaultNumberOfPlayers() int
	GetDefaultNumberOfRounds() int
	ValidateMove(interface{}) error
	EvaluateRound(moves []PlayerMove) RoundResult
}

// PlayerMove has the moves associated to a player
type PlayerMove struct {
	ID   uuid.UUID
	Move interface{}
}

// RoundResult represents the result of a round
type RoundResult struct {
	Status        Status
	PlayerResults []PlayerResult
	Winner        uuid.UUID
}

// PlayerResult represents the result of a player in the context of a round
type PlayerResult struct {
	ID     uuid.UUID
	Status Status
}

// NewGame instantiates a concrete game type specified by the name of the game
func NewGame(name string) (GameType, error) {
	switch name {
	case "rps":
		return &rockPaperScissors{}, nil
	}
	return nil, errors.New("game name was not provided or does not exist")
}
