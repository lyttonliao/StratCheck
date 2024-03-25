package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Ensures all our routes are defined in one single place
// Allows access to the router in any test code by initializing an application and calling the routes()
// method on it
func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance
	router := httprouter.New()

	// Register relevant methods, URL patterns, and handler functions for endpoints using HandleFunc() method
	// 
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/strategies", app.createStrategyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/strategies/:id", app.showStrategyHandler)

	return router
}