package web

import (
	"encoding/json"
	"github.com/google/logger"
	"net/http"
	"strconv"
)

// EncodeJSONResponse uses the json encoder to write an interface to the http response with an optional status code
func EncodeJSONResponse(i interface{}, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(i)
}

// parseIntParameter parses a string parameter to an int64
func parseIntParameter(param string) (int64, error) {
	return strconv.ParseInt(param, 10, 64)
}

func handleServerError(w http.ResponseWriter, err error) {
	logger.Error(err)
	w.WriteHeader(500)
}
