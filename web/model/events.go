package model

// Event represents data that is notified to the clients
type Event struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

// StartGame is the event which signals clients that the game can start
type StartGame struct {
	GameID    string   `json:"gameId,omitempty"`
	Players   []string `json:"players,omitempty"`
	NextRound int      `json:"nextRound,omitempty"`
}

// RoundFinished is the event which tells clients that the round is finished,
// and sends them the results
type RoundFinished struct {
	GameID       string `json:"gameId"`
	CurrentRound int    `json:"currentRound"`
	RoundResult  Result `json:"roundResult"`
	NextRound    int    `json:"nextRound"`
	// Score after the current round
	Score string `json:"score"`
}

// GameFinished is the event which tells clients that the game is finished,
// and sends them the results
type GameFinished struct {
	GameID     string `json:"gameId"`
	Score      string `json:"score"`
	GameResult Result `json:"gameResult"`
}

// Result holds the data that is the result of a round or a game
type Result struct {
	Status string          `json:"status"`
	Winner string          `json:"winner,omitempty"`
	Moves  map[string]Move `json:"moves,omitempty"`
}
