package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/strategies", app.createStrategyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/strategies", app.listStrategiesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/strategies/:id", app.showStrategyHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/strategies/:id", app.updateStrategyHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/strategies/:id", app.deleteStrategyHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}
