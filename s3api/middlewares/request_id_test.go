package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestID_Length(t *testing.T) {
	id, err := GenerateRequestID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 16 bytes encoded as hex = 32 characters
	if len(id) != 32 {
		t.Errorf("expected request ID length 32, got %d", len(id))
	}
}

func TestGenerateRequestID_Uniqueness(t *testing.T) {
	ids := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id, err := GenerateRequestID()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if _, exists := ids[id]; exists {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestRequestIDMiddleware_SetsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Amz-Request-Id") == "" {
		t.Error("expected X-Amz-Request-Id header to be set")
	}
}

func TestRequestIDMiddleware_PreservesClientHeader(t *testing.T) {
	const clientID = "client-provided-id-1234"
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Amz-Request-Id", clientID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Amz-Request-Id"); got != clientID {
		t.Errorf("expected %q, got %q", clientID, got)
	}
}

func TestRequestIDMiddleware_ContextPropagation(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected request ID to be present in context")
	}
	if capturedID != rec.Header().Get("X-Amz-Request-Id") {
		t.Errorf("context ID %q does not match response header %q",
			capturedID, rec.Header().Get("X-Amz-Request-Id"))
	}
}
