package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"relay/internal/job"
	"relay/internal/store"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler := newHandler(store.NewMemoryStore())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	const wantBody = `{"status":"ok"}`
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
}

func TestCreateJobEndpoint(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	handler := newHandler(memoryStore)

	const requestBody = `{
		"type": "send_email",
		"payload": {"recipient": "user@example.com"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var created job.Job
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if created.ID == "" {
		t.Error("created job ID is empty")
	}
	if created.Type != "send_email" {
		t.Errorf("created job type = %q, want %q", created.Type, "send_email")
	}
	if created.Status != job.StatusQueued {
		t.Errorf("created job status = %q, want %q", created.Status, job.StatusQueued)
	}
	if created.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("updated_at is zero")
	}
	if !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Errorf("created_at = %v, updated_at = %v; want equal timestamps", created.CreatedAt, created.UpdatedAt)
	}

	var payload struct {
		Recipient string `json:"recipient"`
	}
	if err := json.Unmarshal(created.Payload, &payload); err != nil {
		t.Fatalf("decode created payload: %v", err)
	}
	if payload.Recipient != "user@example.com" {
		t.Errorf("payload recipient = %q, want %q", payload.Recipient, "user@example.com")
	}

	stored, err := memoryStore.Get(created.ID)
	if err != nil {
		t.Fatalf("get created job from store: %v", err)
	}
	assertJobsEquivalent(t, stored, created)
}

func TestCreateJobEndpointRejectsMalformedJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "truncated object", body: `{"type":"send_email"`},
		{name: "invalid payload", body: `{"type":"send_email","payload":not-json}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandler(store.NewMemoryStore())
			req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestUnsupportedRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler := newHandler(store.NewMemoryStore())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func assertJobsEquivalent(t *testing.T, got, want job.Job) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("stored ID = %q, response ID = %q", got.ID, want.ID)
	}
	if got.Type != want.Type {
		t.Errorf("stored type = %q, response type = %q", got.Type, want.Type)
	}
	assertJSONEquivalent(t, got.Payload, want.Payload)
	if got.Status != want.Status {
		t.Errorf("stored status = %q, response status = %q", got.Status, want.Status)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("stored created_at = %v, response created_at = %v", got.CreatedAt, want.CreatedAt)
	}
	if !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Errorf("stored updated_at = %v, response updated_at = %v", got.UpdatedAt, want.UpdatedAt)
	}
}

func assertJSONEquivalent(t *testing.T, got, want json.RawMessage) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode stored payload %q: %v", got, err)
	}

	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("decode response payload %q: %v", want, err)
	}

	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Errorf("stored payload = %s, response payload = %s", got, want)
	}
}
