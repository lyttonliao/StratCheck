package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lyttonliao/StratCheck/internal/data"
	"github.com/lyttonliao/StratCheck/internal/validator"
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
		app.badRequestResponse(w, r, err)
		return
	}

	strategy := &data.Strategy{
		Name:     input.Name,
		Fields:   input.Fields,
		Criteria: input.Criteria,
	}

	v := validator.New()
	if data.ValidateStrategy(v, strategy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Strategies.Insert(strategy)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("v1/strategies/%d", strategy.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"strategy": strategy}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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

	strategy, err := app.models.Strategies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"strategy": strategy}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateStrategyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	strategy, err := app.models.Strategies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name     *string  `json:"name"`
		Fields   []string `json:"fields"`
		Criteria []string `json:"criteria"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		strategy.Name = *input.Name
	}
	// don't need to dereference a slice
	if input.Fields != nil {
		strategy.Fields = input.Fields
	}
	if input.Criteria != nil {
		strategy.Criteria = input.Criteria
	}

	v := validator.New()
	if data.ValidateStrategy(v, strategy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Strategies.Update(strategy)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"strategy": strategy}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteStrategyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Strategies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "strategy successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
