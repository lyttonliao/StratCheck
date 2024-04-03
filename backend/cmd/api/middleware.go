package main

import (
	"fmt"
	"net/http"
)

// This middleware will only recover panics that happen in the same goroutine that executed the
// recoverPanic() middleware. Must recover any panics from within goroutines
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create deferred function which will always run in event of a pnic as Go unwinds the stack
		defer func() {
			// Check if there has been a panic, set a "Connection: close" header on the response
			// which triggers Go's HTTP server to automatically close the current connection
			// recover() returns a value with type interface{}, use fmt.Errorf() to normalize it into an error
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
