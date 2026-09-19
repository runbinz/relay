package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	// build incoming request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// create ResponseWriter that records what get written to it
	rec := httptest.NewRecorder()

	// call handler and pass in req and rec
	handler := newHandler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// verify content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", contentType)
	}

	// verify expected body
	expectedBody := `{"status":"ok"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}

}

func TestUnsupportedRoute(t *testing.T) {
	// build incoming request to unsupported route
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	// create ResponseWriter that records what get written to it
	rec := httptest.NewRecorder()

	// call handler and pass in req and rec
	handler := newHandler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
