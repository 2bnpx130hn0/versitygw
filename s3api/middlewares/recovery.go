package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// PanicRecovery is a middleware that recovers from panics and returns a 500
// Internal Server Error response, logging the stack trace.
func PanicRecovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					reqID := GetRequestID(r.Context())

					logger.Error("panic recovered",
						"request_id", reqID,
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
						"stack", string(stack),
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
