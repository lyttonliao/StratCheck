package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

// readIDParam() doesn't use dependencies from our app struct, so it could be a regular function
// rather than a method on application. Set up all your app-specific handlers and helpers so
// that they are methods on applications, it helps maintain consistency and future-proof codes for
// when helpers and handlers change and need access to a dependency
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// time.Time encoded as RFC3339-format JSON string instead of JSON object
// []byte encoded as base64-encoded JSON string instead of JSON array
// channels, functions, complex number types cannot be encoded
// pointer values will encode as the value pointed to
// interface values will encode as value contained in the interface
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// limits the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, call DisallowUnknownFields before decoding
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// errors.As() checks if err has type *json.SyntaxError
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		// Decode() may return io.ErrUnexpectedEOF error for syntax errors
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		// unmarshalTypeErrors occur when JSON value is the wrong type for target destination
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		// io.Eor error will be returned by Decode() if request body is empty
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		// extract field name from the error and interpolate it into custom error message
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// json.InvalidUnmarshalError occurs if passing a non-nil pointer to Decode()
		// learn about catching errors and panicking
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to empty anonymous struct as destination
	// This checks if request body only contained a single JSON value
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
