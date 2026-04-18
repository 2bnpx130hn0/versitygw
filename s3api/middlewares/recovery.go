package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// PanicRecovery is a middleware that recovers from panics and returns a 500
// Internal Server Error response, logging the stack trace.
// Note: using "panic recovered" as the log message to make it easy to grep for in logs.
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
						"remote_addr", r.RemoteAddr,
						"panic", rec,
						"stack", string(stack),
					)

					// Use WriteHeader instead of http.Error to avoid including the
					// status text in the body, which can leak implementation details.
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
