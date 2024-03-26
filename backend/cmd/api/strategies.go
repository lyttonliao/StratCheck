package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lyttonliao/StratCheck/internal/data"
)

func (app *application) createStrategyHandler(w http.ResponseWriter, r *http.Request) {
	// declare an anonymous struct to be in HTTP request body
	// this struct will be our *target decode destination*
	var input struct {
		Name     string   `json:"name"`
		Fields   []string `json:"fields"`
		Criteria []string `json:"criteria"`
	}

	// must pass non-nil pointer as target decode destination
	// if destination is a struct, fields must be exported
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showAllStrategyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "show all strategies")
}

func (app *application) showStrategyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	strategy := data.Strategy{
		ID:        id,
		CreatedAt: time.Now(),
		Name:      "Small Cap Gappers",
		Fields:    []string{"Current Day Open Price", "Previous Day Close Price", "volume", "market capitalization", "float"},
		Criteria:  []string{"Current Day Open Price - Previous Day Close Price"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"strategy": strategy}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
