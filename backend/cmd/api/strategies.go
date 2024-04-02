package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(strategy.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
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

func (app *application) listStrategiesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string
		Fields   []string
		Criteria []string
		data.Filters
	}

	v := validator.New()
	// r.URL.Query() returns url.Values map containing the query string data
	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Fields = app.readCSV(qs, "fields", []string{})
	input.Criteria = app.readCSV(qs, "criteria", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "fields", "criteria", "-id", "-fields", "-criteria"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
