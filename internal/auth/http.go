package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/dmksnnk/octo/internal"
)

// HeaderAPIKey header name for the API key.
const HeaderAPIKey = "X-API-Key"

// DB for getting users.
type DB interface {
	UserByAPIKey(ctx context.Context, apiKey string) (internal.User, error)
}

// ErrNotFound is returned when a user is not found in the database.
var ErrNotFound = errors.New("not found")

// CheckUser is an HTTP middleware that finds the user by API key and adds it to the request context.
// If the API key is missing or invalid, it returns an error response.
func CheckUser(db DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(HeaderAPIKey)
			if apiKey == "" {
				http.Error(w, "missing API key", http.StatusUnauthorized)
				return
			}

			user, err := db.UserByAPIKey(r.Context(), apiKey)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					http.Error(w, "invalid API key", http.StatusUnauthorized)
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ctx := ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
