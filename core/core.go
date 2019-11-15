// Package core contains the main operations for the game engine
package core

import (
	"botServer/core/events"
	"botServer/core/games"
	"fmt"
	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

var (
	tokenToGameID = make(map[string]uuid.UUID)
	gameIDToGame  = make(map[uuid.UUID]*game)
	playerNameNr  int64
)

type game struct {
	id              uuid.UUID
	gameType        games.GameType
	numberOfPlayers int
	players         map[uuid.UUID]*Player
	currentRound    int
	totalRounds     int
}

// Connect tries to connect a new user to a game specified by the token
func Connect(req ConnectRequest) (ConnectResponse, error) {
	g, err := getOrCreateGame(req.Token, req.GameName, req.NoOfPlayers, req.TotalRounds)
	if err != nil {
		return ConnectResponse{}, errors.Wrap(err, "could not connect to game")
	}
	logger.Infof("Game with id %s was created", g.id.String())
	p := getOrCreatePlayer(req.PlayerName, req.EventCallback)
	if len(g.players) >= g.numberOfPlayers {
		err := errors.New("all players are already connected")
		return ConnectResponse{}, errors.Wrap(err, "could not connect to game")
	}
	g.players[p.ID] = p
	logger.Infof("Player with id %s joined the game with id %s", p.ID.String(), g.id.String())
	if len(g.players) == g.numberOfPlayers {
		go tryStartGame(g)
	}
	return ConnectResponse{GameID: g.id, Player: *p, Rounds: g.totalRounds}, nil
}

func RegisterWS(gameId, playerId uuid.UUID, conn *websocket.Conn) error {
	g, ok := gameIDToGame[gameId]
	if !ok {
		err := errors.New("game id is not correct")
		return errors.Wrap(err, "could not register ws")
	}
	p, ok := g.players[playerId]
	if !ok {
		err := errors.New("player id is not correct")
		return errors.Wrap(err, "could not register ws")
	}
	p.WebsocketConn = conn
	return nil
}

// Play processes a given player's move in a given round in a specific game
func Play(req PlayRequest) (PlayResponse, error) {
	g, ok := gameIDToGame[req.GameID]
	if !ok {
		err := errors.New("game id is not correct")
		return PlayResponse{}, errors.Wrap(err, "could not make move")
	}
	if g.currentRound == 0 {
		err := errors.New("game has not started yet")
		return PlayResponse{}, errors.Wrap(err, "could not make move")
	}
	if req.Round != g.currentRound {
		err := errors.Errorf("%d is not the current round (%d)", req.Round, g.currentRound)
		return PlayResponse{}, errors.Wrap(err, "could not make move")
	}
	err := g.gameType.ValidateMove(req.Move)
	if err != nil {
		return PlayResponse{}, errors.Wrap(err, "could not make move")
	}
	p, ok := g.players[req.PlayerID]
	if !ok {
		err := errors.New("player id is not correct")
		return PlayResponse{}, errors.Wrap(err, "could not make move")
	}
	logger.Infof("Play for game %s round %d and player %s", req.GameID.String(), req.Round, g.players[req.PlayerID].Name)
	p.currentMove = req.Move
	playersToMove := playersToMakeMove(g.players)
	if len(playersToMove) == 0 {
		go finishRound(g)
	}
	return PlayResponse{
		PlayersMove: playersToMove,
		Round:       g.currentRound,
	}, nil
}

func getOrCreateGame(token, gameName string, noOfPlayers, totalRounds int) (*game, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}
	gameID, ok := tokenToGameID[token]
	if !ok {
		gameType, err := games.NewGame(gameName)
		if err != nil {
			return nil, errors.Wrap(err, "could not create new game")
		}
		numberOfPlayers, err := getNumberOfPlayers(gameType, noOfPlayers)
		if err != nil {
			return nil, errors.Wrap(err, "could not create new game")
		}

		if totalRounds <= 0 {
			totalRounds = gameType.GetDefaultNumberOfRounds()
		}
		gameID := uuid.New()
		tokenToGameID[token] = gameID
		gameIDToGame[gameID] = &game{
			id:              gameID,
			gameType:        gameType,
			numberOfPlayers: numberOfPlayers,
			players:         make(map[uuid.UUID]*Player),
			currentRound:    0,
			totalRounds:     totalRounds,
		}
		return gameIDToGame[gameID], nil
	}
	return gameIDToGame[gameID], nil
}

func getNumberOfPlayers(gameType games.GameType, noOfPlayers int) (int, error) {
	if noOfPlayers != 0 {
		if !gameType.Validate(noOfPlayers) {
			return 0, errors.New("number of players is invalid for this game")
		}
		return noOfPlayers, nil
	}
	return gameType.GetDefaultNumberOfPlayers(), nil
}

func playersToMakeMove(players map[uuid.UUID]*Player) []string {
	var playersToMakeMove = make([]string, 0)
	for _, p := range players {
		if p.currentMove == nil {
			playersToMakeMove = append(playersToMakeMove, p.Name)
		}
	}
	return playersToMakeMove
}

func finishRound(g *game) {
	var moves = make([]games.PlayerMove, 0)
	for id, p := range g.players {
		moves = append(moves, games.PlayerMove{ID: id, Move: p.currentMove})
	}
	result := g.gameType.EvaluateRound(moves)
	if result.Status == games.DRAW {
		logger.Infof("Result for game %s and round %d is DRAW", g.id, g.currentRound)
		oldRound := g.currentRound
		for _, player := range g.players {
			player.currentMove = nil
		}
		notifyRoundFinished(g, oldRound, result, moves)
		return
	}
	oldRound := g.currentRound
	g.currentRound++
	for _, playerResult := range result.PlayerResults {
		g.players[playerResult.ID].currentMove = nil
		if playerResult.Status == games.WIN {
			g.players[playerResult.ID].score++
		}
	}

	if isGameOver(g) {
		logger.Infof("Game %s is over, Winner is %s, Score is %s", g.id, g.players[result.Winner].Name, scoreAsString(g.players))
		removeGame(g)
		notifyGameFinished(g, result)
	} else {
		logger.Infof("Winner for game %s and round %d is %s, Score is %s", g.id, oldRound, g.players[result.Winner].Name, scoreAsString(g.players))
		notifyRoundFinished(g, oldRound, result, moves)
	}
}

func tryStartGame(g *game) {
	reachablePlayers, unreachablePlayers := splitReachableAndUnreachablePlayers(g.players)
	for i := 10; i > 0; i-- {
		if len(unreachablePlayers) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
		reachablePlayers, unreachablePlayers = splitReachableAndUnreachablePlayers(g.players)
	}
	if len(unreachablePlayers) > 0 {
		logger.Warningf("Game will not start, unreachable players are: %s", strings.Join(unreachablePlayers, ", "))
		message := fmt.Sprintf("Game will not start and you will need to reconnect, unreachable players are: %s", strings.Join(unreachablePlayers, ", "))
		notifyError(reachablePlayers, message)
		removeGame(g)
		return
	}
	g.currentRound = 1
	notifyStartGame(g.players, g.id, g.currentRound)
}

func notifyStartGame(players map[uuid.UUID]*Player, gameID uuid.UUID, nextRound int) {
	var subscribers []events.Subscriber
	var playerNames []string
	for _, player := range players {
		subscribers = append(subscribers, events.Subscriber{
			Callback:      player.EventCallback,
			WebsocketConn: player.WebsocketConn,
		})
		playerNames = append(playerNames, player.Name)
	}
	events.PublishStartGame(events.StartGame{
		GameID:      gameID,
		Players:     playerNames,
		Subscribers: subscribers,
		NextRound:   nextRound,
	})
}

func notifyRoundFinished(g *game, oldRound int, result games.RoundResult, moves []games.PlayerMove) {
	var winner string
	if result.Status != games.DRAW {
		winner = g.players[result.Winner].Name
	}
	movesMap := make(map[string]interface{}, len(moves))
	for _, move := range moves {
		movesMap[g.players[move.ID].Name] = move.Move
	}
	events.PublishRoundFinished(events.RoundFinished{
		GameID:        g.id,
		CurrentRound:  oldRound,
		NextRound:     g.currentRound,
		PlayerResults: computePlayerResults(g.players, result),
		Winner:        winner,
		Moves:         movesMap,
	})
}

func notifyGameFinished(g *game, result games.RoundResult) {
	events.PublishGameFinished(events.GameFinished{
		GameID:        g.id,
		PlayerResults: computePlayerResults(g.players, result),
		Winner:        g.players[result.Winner].Name,
	})
}

func notifyError(players map[uuid.UUID]*Player, message string) {
	var subscribers []events.Subscriber
	for _, player := range players {
		subscribers = append(subscribers, events.Subscriber{
			Callback:      player.EventCallback,
			WebsocketConn: player.WebsocketConn,
		})
	}
	events.PublishError(subscribers, message)
}

func computePlayerResults(players map[uuid.UUID]*Player, result games.RoundResult) []events.PlayerResult {
	var playerResults []events.PlayerResult
	for _, playerResult := range result.PlayerResults {
		scores := []string{strconv.Itoa(players[playerResult.ID].score)}
		for id, p := range players {
			if id != playerResult.ID {
				scores = append(scores, strconv.Itoa(p.score))
			}
		}
		playerResults = append(playerResults, events.PlayerResult{
			Status: string(playerResult.Status),
			Score:  strings.Join(scores, "-"),
			Subscriber: events.Subscriber{
				Callback:      players[playerResult.ID].EventCallback,
				WebsocketConn: players[playerResult.ID].WebsocketConn,
			},
		})
	}
	return playerResults
}

func scoreAsString(players map[uuid.UUID]*Player) string {
	var scores []string
	for _, p := range players {
		scores = append(scores, strconv.Itoa(p.score))
	}
	return strings.Join(scores, "-")
}

func isGameOver(g *game) bool {
	if g.currentRound > g.totalRounds {
		return true
	}
	for _, p := range g.players {
		if p.score > (g.totalRounds-1)/2 {
			return true
		}
	}
	return false
}

func splitReachableAndUnreachablePlayers(players map[uuid.UUID]*Player) (map[uuid.UUID]*Player, []string) {
	reachablePlayers := make(map[uuid.UUID]*Player, 0)
	var unreachablePlayers []string
	for id, player := range players {
		if (player.EventCallback == nil || !player.EventCallback.IsAbs()) && (player.WebsocketConn == nil) {
			unreachablePlayers = append(unreachablePlayers, player.Name)
		} else {
			reachablePlayers[id] = player
		}
	}
	return reachablePlayers, unreachablePlayers
}

func removeGame(g *game) {
	t := ""
	for token, gameID := range tokenToGameID {
		if gameID == g.id {
			t = token
			break
		}
	}
	delete(tokenToGameID, t)
	delete(gameIDToGame, g.id)
}
