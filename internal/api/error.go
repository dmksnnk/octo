package api

import (
	"net/http"
)

// Error defines model for Error.
type Error struct {
	// Code is a status code.
	Code int `json:"code"`

	// Message is a developer-facing error message.
	Message string `json:"message"`

	// Details is the error details.
	Details []string `json:"details"`
}

func writeError(w http.ResponseWriter, message string, code int, details ...string) {
	apiError := Error{
		Code:    code,
		Message: message,
		Details: details,
	}

	_ = writeJSON(w, code, apiError)
}
