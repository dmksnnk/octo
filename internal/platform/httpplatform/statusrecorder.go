package httpplatform

import "net/http"

// StatusRecorder is a wrapper around http.ResponseWriter that
// allows capturing of the HTTP status code written during request handling.
type StatusRecorder struct {
	http.ResponseWriter     // Embedded http.ResponseWriter to delegate functionality.
	statusCode          int // statusCode captures the HTTP status code set by the handler.
}

// NewStatusRecorder initializes a new StatusRecorder with the provided ResponseWriter.
// By default, it assumes an initial status code of http.StatusOK.
func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{w, http.StatusOK}
}

// WriteHeader captures the provided status code and then delegates
// the call to the embedded ResponseWriter's WriteHeader method.
func (rw *StatusRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status returns the captured HTTP status code. If no code was explicitly
// set by the handler, it returns http.StatusOK by default.
func (rw *StatusRecorder) Status() int {
	return rw.statusCode
}
