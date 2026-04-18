package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/versitygw/versitygw/s3api/middlewares"
)

func TestDetectAuthType_V4(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20230101/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-date, Signature=abc123")

	authType := middlewares.DetectAuthType(req)
	if authType != middlewares.AuthTypeV4 {
		t.Errorf("expected AuthTypeV4, got %v", authType)
	}
}

func TestDetectAuthType_V2(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "AWS AKIAIOSFODNN7EXAMPLE:wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	authType := middlewares.DetectAuthType(req)
	if authType != middlewares.AuthTypeV2 {
		t.Errorf("expected AuthTypeV2, got %v", authType)
	}
}

func TestDetectAuthType_None(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	authType := middlewares.DetectAuthType(req)
	if authType != middlewares.AuthTypeNone {
		t.Errorf("expected AuthTypeNone, got %v", authType)
	}
}

func TestParseV4AuthHeader_Valid(t *testing.T) {
	authHeader := "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20230101/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-date, Signature=abc123def456"

	meta, err := middlewares.ParseV4AuthHeader(authHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.AccessKey != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected access key AKIAIOSFODNN7EXAMPLE, got %s", meta.AccessKey)
	}
	if meta.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", meta.Region)
	}
	if meta.Service != "s3" {
		t.Errorf("expected service s3, got %s", meta.Service)
	}
	if meta.Signature != "abc123def456" {
		t.Errorf("expected signature abc123def456, got %s", meta.Signature)
	}
}

func TestParseV4AuthHeader_MissingCredential(t *testing.T) {
	authHeader := "AWS4-HMAC-SHA256 SignedHeaders=host;x-amz-date, Signature=abc123"

	_, err := middlewares.ParseV4AuthHeader(authHeader)
	if err == nil {
		t.Error("expected error for missing Credential, got nil")
	}
}

func TestParseV4AuthHeader_MalformedCredential(t *testing.T) {
	authHeader := "AWS4-HMAC-SHA256 Credential=BADCREDENTIAL, SignedHeaders=host, Signature=abc123"

	_, err := middlewares.ParseV4AuthHeader(authHeader)
	if err == nil {
		t.Error("expected error for malformed Credential, got nil")
	}
}

func TestGetAuthMetadata_StoredInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20230101/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-date, Signature=abc123")

	var capturedMeta *middlewares.AuthMetadata
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMeta = middlewares.GetAuthMetadata(r.Context())
	})

	middleware := middlewares.AuthParser(nextHandler)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if capturedMeta == nil {
		t.Fatal("expected auth metadata in context, got nil")
	}
	if capturedMeta.AccessKey != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected access key AKIAIOSFODNN7EXAMPLE, got %s", capturedMeta.AccessKey)
	}
}

func TestAuthParser_NoAuthHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var capturedMeta *middlewares.AuthMetadata
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMeta = middlewares.GetAuthMetadata(r.Context())
	})

	middleware := middlewares.AuthParser(nextHandler)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if capturedMeta != nil {
		t.Errorf("expected nil auth metadata for unauthenticated request, got %+v", capturedMeta)
	}
}
