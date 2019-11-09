package games

import (
	"github.com/pkg/errors"
	"strings"
)

const (
	defaultNumberOfPlayers = 2
	defaultNumberOfRounds  = 3
)

var (
	winsAgainst = map[string]string{"rock": "paper", "paper": "scissors", "scissors": "rock"}
)

type rockPaperScissors struct{}

// Validate verifies if the given number of players is valid
func (rps *rockPaperScissors) Validate(noOfPlayers int) bool {
	return noOfPlayers == defaultNumberOfPlayers
}

// GetDefaultNumberOfPlayers returns the default number of players
func (rps *rockPaperScissors) GetDefaultNumberOfPlayers() int {
	return defaultNumberOfPlayers
}

// GetDefaultNumberOfRounds returns the default number of rounds
func (rps *rockPaperScissors) GetDefaultNumberOfRounds() int {
	return defaultNumberOfRounds
}

// ValidateMove checks if the given move is valid
func (rps *rockPaperScissors) ValidateMove(move interface{}) error {
	validMovesString := validMovesString()
	s, ok := move.(string)
	if !ok {
		return errors.New("Move needs to be a string, one of the values: " + validMovesString)
	}
	if _, ok := winsAgainst[s]; !ok {
		return errors.New("Move needs to be one of the values: " + validMovesString)
	}
	return nil
}

// EvaluateRound processes the given moves, and computes the result of the round
func (rps *rockPaperScissors) EvaluateRound(moves []PlayerMove) RoundResult {
	var playerResults []PlayerResult
	firstMove := moves[0].Move.(string)
	secondMove := moves[1].Move.(string)
	if winsAgainst[firstMove] == secondMove {
		playerResults = append(playerResults, PlayerResult{ID: moves[0].ID, Status: LOSE})
		playerResults = append(playerResults, PlayerResult{ID: moves[1].ID, Status: WIN})
		return RoundResult{Status: WIN, PlayerResults: playerResults, Winner: moves[1].ID}
	}
	if winsAgainst[secondMove] == firstMove {
		playerResults = append(playerResults, PlayerResult{ID: moves[0].ID, Status: WIN})
		playerResults = append(playerResults, PlayerResult{ID: moves[1].ID, Status: LOSE})
		return RoundResult{Status: WIN, PlayerResults: playerResults, Winner: moves[0].ID}
	}
	playerResults = append(playerResults, PlayerResult{ID: moves[0].ID, Status: DRAW})
	playerResults = append(playerResults, PlayerResult{ID: moves[1].ID, Status: DRAW})
	return RoundResult{Status: DRAW, PlayerResults: playerResults}
}

func validMovesString() string {
	s := ""
	for k := range winsAgainst {
		s += k + ","
	}
	return strings.TrimRight(s, ",")
}
