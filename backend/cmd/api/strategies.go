package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *application) forwardRequestHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	url := fmt.Sprintf("http://localhost:8000%s", r.URL)
	fmt.Println("Forwarding request to: ", url)

	proxyReq, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	proxyReq.Close = true

	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set("Host", r.Host)
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", "Bearer "+cookie.Value)

	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer proxyRes.Body.Close()

	var payload interface{}
	err = json.NewDecoder(proxyRes.Body).Decode(&payload)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, proxyRes.StatusCode, envelope{"payload": payload}, proxyReq.Header)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// func (app *application) createStrategyHandler(w http.ResponseWriter, r *http.Request) {
// 	var input struct {
// 		Name     string   `json:"name"`
// 		Fields   []string `json:"fields"`
// 		Criteria []string `json:"criteria"`
// 		Public   bool     `json:"public"`
// 	}

// 	err := app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	user := app.contextGetUser(r)

// 	strategy := &data.Strategy{
// 		Name:     input.Name,
// 		Fields:   input.Fields,
// 		Criteria: input.Criteria,
// 		Public:   input.Public,
// 		UserID:   user.ID,
// 	}

// 	v := validator.New()
// 	if data.ValidateStrategy(v, strategy); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	err = app.models.Strategies.Insert(user.ID, strategy)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	headers := make(http.Header)
// 	headers.Set("Location", fmt.Sprintf("v1/strategies/%d", strategy.ID))

// 	err = app.writeJSON(w, http.StatusCreated, envelope{"strategy": strategy}, headers)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}

// 	fmt.Fprintf(w, "%+v\n", input)
// }

// func (app *application) showStrategyHandler(w http.ResponseWriter, r *http.Request) {
// 	strategyID, err := app.readIDParam(r)
// 	if err != nil {
// 		app.notFoundResponse(w, r)
// 		return
// 	}

// 	user := app.contextGetUser(r)

// 	strategy, err := app.models.Strategies.Get(user.ID, strategyID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"strategy": strategy}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) updateStrategyHandler(w http.ResponseWriter, r *http.Request) {
// 	strategyID, err := app.readIDParam(r)
// 	if err != nil {
// 		app.notFoundResponse(w, r)
// 		return
// 	}

// 	user := app.contextGetUser(r)

// 	strategy, err := app.models.Strategies.Get(user.ID, strategyID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	if r.Header.Get("X-Expected-Version") != "" {
// 		if strconv.FormatInt(int64(strategy.Version), 32) != r.Header.Get("X-Expected-Version") {
// 			app.editConflictResponse(w, r)
// 			return
// 		}
// 	}

// 	var input struct {
// 		Name     *string  `json:"name"`
// 		Fields   []string `json:"fields"`
// 		Criteria []string `json:"criteria"`
// 		Public   bool     `json:"public"`
// 	}

// 	err = app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	if input.Name != nil {
// 		strategy.Name = *input.Name
// 	}
// 	// don't need to dereference a slice
// 	if input.Fields != nil {
// 		strategy.Fields = input.Fields
// 	}
// 	if input.Criteria != nil {
// 		strategy.Criteria = input.Criteria
// 	}

// 	v := validator.New()
// 	if data.ValidateStrategy(v, strategy); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	err = app.models.Strategies.Update(user.ID, strategy)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrEditConflict):
// 			app.editConflictResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"strategy": strategy}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) deleteStrategyHandler(w http.ResponseWriter, r *http.Request) {
// 	strategyID, err := app.readIDParam(r)
// 	if err != nil {
// 		app.notFoundResponse(w, r)
// 		return
// 	}

// 	user := app.contextGetUser(r)

// 	err = app.models.Strategies.Delete(user.ID, strategyID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"message": "strategy successfully deleted"}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) listStrategiesHandler(w http.ResponseWriter, r *http.Request) {
// 	var input struct {
// 		Name     string
// 		Fields   []string
// 		Criteria []string
// 		data.Filters
// 	}

// 	user := app.contextGetUser(r)

// 	v := validator.New()
// 	// r.URL.Query() returns url.Values map containing the query string data
// 	qs := r.URL.Query()

// 	input.Name = app.readString(qs, "name", "")
// 	input.Fields = app.readCSV(qs, "fields", []string{})
// 	input.Criteria = app.readCSV(qs, "criteria", []string{})
// 	input.Filters.Page = app.readInt(qs, "page", 1, v)
// 	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
// 	input.Filters.Sort = app.readString(qs, "sort", "id")
// 	input.Filters.SortSafelist = []string{"id", "name", "fields", "-id", "-name", "-fields"}

// 	if data.ValidateFilters(v, input.Filters); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	strategies, metadata, err := app.models.Strategies.GetAll(user.ID, input.Name, input.Fields, input.Filters)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"strategies": strategies, "metadata": metadata}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}

// 	fmt.Fprintf(w, "%+v\n", input)
// }
