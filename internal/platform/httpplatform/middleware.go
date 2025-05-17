package httpplatform

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dmksnnk/octo/internal/auth"
)

// Middleware wraps an http.Handler to provide additional functionality.
type Middleware = func(next http.Handler) http.Handler

// LogRequests is a middleware that logs HTTP requests.
func LogRequests(logger *slog.Logger) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sr, ok := w.(*StatusRecorder) // ensure that we have a status recorder
			if !ok {
				sr = NewStatusRecorder(w)
			}

			handler.ServeHTTP(sr, r)

			attrs := make([]slog.Attr, 0, 4)
			user, ok := auth.ContextUser(r.Context())
			if ok {
				attrs = append(attrs, slog.Any("user", user))
			}

			attrs = append(attrs,
				slog.Group("request",
					slog.String("method", r.Method),
					slog.String("proto", r.Proto),
					slog.String("host", r.Host),
					slog.String("remote", r.RemoteAddr),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
				),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.Int("status_code", sr.Status()),
			)

			logger.LogAttrs(r.Context(), slog.LevelInfo, "request handled", attrs...)
		})
	}
}

// Wrap handler with middlewares.
func Wrap(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		mw := mws[i]
		if mw != nil {
			handler = mw(handler)
		}
	}

	return handler
}
