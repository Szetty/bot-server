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
	"botServer/web/model"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

var Upgrader = websocket.Upgrader{}

// A ConnectAPIController binds http requests to an api service and writes the service results to the http response
type ConnectAPIController struct {
	service ConnectAPIServicer
}

// NewConnectAPIController creates a default api controller
func NewConnectAPIController(s ConnectAPIServicer) Router {
	return &ConnectAPIController{service: s}
}

// Routes returns all of the api route for the ConnectAPIController
func (c *ConnectAPIController) Routes() Routes {
	return Routes{
		{
			"HelloPost",
			strings.ToUpper("Post"),
			"/hello",
			c.HelloPost,
		},
		{
			"SwitchToWs",
			strings.ToUpper("Get"),
			"/ws",
			c.SwitchToWs,
		},
	}
}

// HelloPost -
func (c *ConnectAPIController) HelloPost(w http.ResponseWriter, r *http.Request) {
	helloRequest := &model.HelloRequest{}
	if err := json.NewDecoder(r.Body).Decode(&helloRequest); err != nil {
		errorResponse := &model.Error{Message: err.Error()}
		err = EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}

	result, err := c.service.HelloPost(*helloRequest)
	if err != nil {
		errorResponse := &model.Error{Message: err.Error()}
		err = EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}

	err = EncodeJSONResponse(result, http.StatusOK, w)
	if err != nil {
		handleServerError(w, err)
	}
}

func (c *ConnectAPIController) SwitchToWs(w http.ResponseWriter, r *http.Request) {
	gameId, ok := r.URL.Query()["gameId"]
	if !ok || len(gameId) < 1 || len(gameId[0]) < 1 {
		errorResponse := &model.Error{Message: "Missing gameId in query string"}
		err := EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}
	playerId, ok := r.URL.Query()["playerId"]
	if !ok || len(playerId) < 1 || len(playerId[0]) < 1 {
		errorResponse := &model.Error{Message: "Missing playerId in query string"}
		err := EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}
	switchToWsRequest := model.SwitchToWsRequest{
		GameID:   gameId[0],
		PlayerID: playerId[0],
	}
	websocketConn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		err = errors.Wrap(err, "could not upgrade to websocket")
		errorResponse := &model.Error{Message: err.Error()}
		err = EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}
	err = c.service.SwitchToWs(switchToWsRequest, websocketConn)
	if err != nil {
		err = errors.Wrap(err, "could not upgrade to websocket")
		errorResponse := &model.Error{Message: err.Error()}
		err = EncodeJSONResponse(errorResponse, http.StatusBadRequest, w)
		if err != nil {
			handleServerError(w, err)
		}
		return
	}
}
