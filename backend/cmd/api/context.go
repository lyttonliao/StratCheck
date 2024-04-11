package main

import (
	"context"
	"net/http"

	"github.com/lyttonliao/StratCheck/internal/data"
)

type contextKey string

// Convert the string "user" to a contextKey type and assign it to a constant
const userContextKey = contextKey("user")

// Returns a new copy of the request with the provided User struct added to the context
// 'userContextKey' constant is the key
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Retrieves the User struct from the request context, only use this when we logically
// expect there to be a User struct value in the context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
