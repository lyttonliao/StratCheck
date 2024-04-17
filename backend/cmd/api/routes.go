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
	router.HandlerFunc(http.MethodPost, "/v1/strategies", app.requirePermission("strategies:write", app.createStrategyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/strategies", app.requirePermission("strategies:read", app.listStrategiesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/strategies/:id", app.requirePermission("strategies:read", app.showStrategyHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/strategies/:id", app.requirePermission("strategies:write", app.updateStrategyHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/strategies/:id", app.requirePermission("strategies:write", app.deleteStrategyHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Position CORs middleware before rate limiter because any CORs that exceed the rate limit
	// should not have the Access-Control-Allow-Origin header set
	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
