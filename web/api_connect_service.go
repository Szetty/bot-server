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
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/url"
)

// ConnectAPIService is a service that implents the logic for the ConnectAPIServicer
// This service should implement the business logic for every endpoint for the ConnectApi API.
// Include any external packages or services that will be required by this service.
type ConnectAPIService struct {
}

// NewConnectAPIService creates a default api service
func NewConnectAPIService() ConnectAPIServicer {
	return &ConnectAPIService{}
}

// HelloPost -
func (s *ConnectAPIService) HelloPost(helloRequest model.HelloRequest) (model.HelloResponse, error) {
	if helloRequest.EventCallback == "" {
		return model.HelloResponse{}, errors.New("event callback URL cannot be empty")
	}
	callbackURL, err := url.Parse(helloRequest.EventCallback)
	if err != nil {
		return model.HelloResponse{}, err
	}
	connectResponse, err := core.Connect(core.ConnectRequest{
		GameName:      helloRequest.Game.Name,
		Token:         helloRequest.Game.ConnectionToken,
		NoOfPlayers:   helloRequest.Game.NumberOfTotalPlayers,
		PlayerName:    helloRequest.PlayerName,
		EventCallback: callbackURL,
		TotalRounds:   helloRequest.Game.TotalRounds,
	})
	if err != nil {
		return model.HelloResponse{}, err
	}
	return model.HelloResponse{
		GameID: connectResponse.GameID.String(),
		Rounds: connectResponse.Rounds,
		Player: model.HelloResponsePlayer{
			ID:   connectResponse.Player.ID.String(),
			Name: connectResponse.Player.Name,
		},
	}, nil
}

func (s *ConnectAPIService) SwitchToWs(request model.SwitchToWsRequest, conn *websocket.Conn) error {
	gameID, err := uuid.Parse(request.GameID)
	if err != nil {
		return errors.Wrap(err, "could not switch to WS: invalid game id")
	}
	playerID, err := uuid.Parse(request.PlayerID)
	if err != nil {
		return errors.Wrap(err, "could not switch to WS: invalid player id")
	}

	err = core.RegisterWS(gameID, playerID, conn)
	if err != nil {
		return err
	}
	return nil
}
