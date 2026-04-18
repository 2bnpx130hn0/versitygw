package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger returns a middleware that logs incoming HTTP requests
// along with their method, path, status code, duration, and request ID.
// Note: uses Warn level for 4xx responses and Error level for 5xx responses
// to make it easier to distinguish client errors from server errors in logs.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			requestID := GetRequestID(r.Context())
			duration := time.Since(start)

			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", duration),
				slog.String("request_id", requestID),
				slog.String("remote_addr", r.RemoteAddr),
				// include query string to help with debugging presigned URL issues
				slog.String("query", r.URL.RawQuery),
			}

			switch {
			case rw.statusCode >= 500:
				logger.ErrorContext(r.Context(), "request completed", attrs...)
			case rw.statusCode >= 400:
				logger.WarnContext(r.Context(), "request completed", attrs...)
			default:
				logger.InfoContext(r.Context(), "request completed", attrs...)
			}
		})
	}
}
