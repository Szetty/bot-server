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
	"github.com/gorilla/mux"
	"net/http"
)

// A Route defines the parameters for an api endpoint
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes are a collection of defined api endpoints
type Routes []Route

// Router defines the required methods for retrieving api routes
type Router interface {
	Routes() Routes
}

// NewRouter creates a new router for any number of api routers
func NewRouter(routers ...Router) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, api := range routers {
		for _, route := range api.Routes() {
			var handler http.Handler
			handler = route.HandlerFunc
			handler = Logger(handler, route.Name)

			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}

	return router
}
