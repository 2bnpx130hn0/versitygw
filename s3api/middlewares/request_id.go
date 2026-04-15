package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// GenerateRequestID creates a cryptographically random 16-byte hex string
// suitable for use as an S3-compatible request ID.
func GenerateRequestID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RequestID is an HTTP middleware that attaches a unique request ID to each
// incoming request and propagates it via the X-Amz-Request-Id response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Amz-Request-Id")
		if reqID == "" {
			var err error
			reqID, err = GenerateRequestID()
			if err != nil {
				// Fall back to a static placeholder so the request is not blocked.
				reqID = "00000000000000000000000000000000"
			}
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set("X-Amz-Request-Id", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID stored in the context, or an empty
// string if none is present.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
