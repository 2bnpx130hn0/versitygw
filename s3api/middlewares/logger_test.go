package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["method"] != http.MethodGet {
		t.Errorf("expected method %q, got %q", http.MethodGet, logEntry["method"])
	}
	if logEntry["path"] != "/test-path" {
		t.Errorf("expected path %q, got %q", "/test-path", logEntry["path"])
	}
	if int(logEntry["status"].(float64)) != http.StatusOK {
		t.Errorf("expected status %d, got %v", http.StatusOK, logEntry["status"])
	}
}

func TestRequestLogger_IncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	const testID = "test-request-id-123"

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPut, "/bucket/key", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, testID))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if logEntry["request_id"] != testID {
		t.Errorf("expected request_id %q, got %q", testID, logEntry["request_id"])
	}
	if int(logEntry["status"].(float64)) != http.StatusAccepted {
		t.Errorf("expected status %d, got %v", http.StatusAccepted, logEntry["status"])
	}
}

func TestRequestLogger_CapturesNonOKStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if int(logEntry["status"].(float64)) != http.StatusNotFound {
		t.Errorf("expected status %d, got %v", http.StatusNotFound, logEntry["status"])
	}
}
