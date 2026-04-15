package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/versitygw/versitygw/s3err"
)

const (
	// DefaultRequestTimeout is the default timeout for incoming S3 requests.
	DefaultRequestTimeout = 30 * time.Second
)

// RequestTimeout returns a middleware that cancels the request context
// after the specified duration. If duration is zero, DefaultRequestTimeout
// is used.
func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	if timeout <= 0 {
		timeout = DefaultRequestTimeout
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()

			select {
			case <-done:
				// request completed normally
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					reqID := GetRequestID(r.Context())
					w.Header().Set("x-amz-request-id", reqID)
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(s3err.GetAPIError(s3err.ErrRequestTimeout).Description))
				}
			}
		})
	}
}
