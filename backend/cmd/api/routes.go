package main

import (
	"expvar"
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

	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Position CORs middleware before rate limiter because any CORs that exceed the rate limit
	// should not have the Access-Control-Allow-Origin header set
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
