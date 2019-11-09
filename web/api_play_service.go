/*
 * Bot Server API
 *
 * This is a bot API to let bots battle
 *
 * API version: 1.0.0
 * Contact: szederjesiarnold@gmail.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package web

import (
	"botServer/core"
	"botServer/web/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// PlayAPIService is a service that implents the logic for the PlayAPIServicer
// This service should implement the business logic for every endpoint for the PlayApi API.
// Include any external packages or services that will be required by this service.
type PlayAPIService struct {
}

// NewPlayAPIService creates a default api service
func NewPlayAPIService() PlayAPIServicer {
	return &PlayAPIService{}
}

// PlayPost -
func (s *PlayAPIService) PlayPost(playRequest model.PlayRequest) (model.PlayResponse, error) {
	gameID, err := uuid.Parse(playRequest.GameID)
	if err != nil {
		return model.PlayResponse{}, errors.Wrap(err, "could not process your move: invalid game id")
	}
	playerID, err := uuid.Parse(playRequest.PlayerID)
	if err != nil {
		return model.PlayResponse{}, errors.Wrap(err, "could not process your move: invalid player id")
	}
	if playRequest.Move.Value == "" {
		return model.PlayResponse{}, errors.New("could not process your move: move cannot be empty")
	}
	playResponse, err := core.Play(core.PlayRequest{
		GameID:   gameID,
		PlayerID: playerID,
		Round:    playRequest.Round,
		Move:     playRequest.Move.Value,
	})
	if err != nil {
		return model.PlayResponse{}, err
	}
	return model.PlayResponse{
		PlayersYetToMakeMove: playResponse.PlayersMove,
		Round:                playResponse.Round,
	}, nil
}