package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)


// readIDParam() doesn't use dependencies from our app struct, so it could be a regular function
// rather than a method on application. Set up all your app-specific handlers and helpers so
// that they are methods on applications, it helps maintain consistency and future-proof codes for
// when helpers and handlers change and need access to a dependency
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invaliud id parameter")
	}

	return id, nil
}