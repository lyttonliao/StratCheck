package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) createStrategyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new strategy")
}

func (app *application) showStrategyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "show the details of the strategy %d\n", id)
}