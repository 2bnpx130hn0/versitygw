package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestTimeout_DefaultTimeout(t *testing.T) {
	mw := RequestTimeout(0)
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestRequestTimeout_CompletesBeforeDeadline(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestTimeout(500 * time.Millisecond)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequestTimeout_ExceedsDeadline(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sleep longer than the timeout to reliably trigger the deadline;
		// 2s sleep vs 50ms timeout gives a generous margin to avoid flakiness
		// on slow CI runners
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestTimeout(50 * time.Millisecond)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
}

func TestRequestTimeout_SetsRequestIDOnTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sleep longer than the timeout to reliably trigger the deadline;
		// 2s sleep vs 50ms timeout gives a generous margin to avoid flakiness
		// on slow CI runners
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestTimeout(50 * time.Millisecond)
	ts := httptest.NewServer(RequestID(mw(handler)))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}

	if resp.Header.Get("x-amz-request-id") == "" {
		t.Error("expected x-amz-request-id header to be set on timeout")
	}
}
