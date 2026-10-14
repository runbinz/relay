package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"relay/internal/job"
	"relay/internal/store"
	"time"
)

type createJobRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func newID() (string, error) {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func main() {
	memoryStore := store.NewMemoryStore()
	// create a new server
	server := &http.Server{
		Addr:    ":8080",
		Handler: newHandler(memoryStore),
	}

	log.Println("server listening on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// construct and return the router
// this isolates the router when you call it for testing purposes
func newHandler(s *store.MemoryStore) http.Handler {
	mux := http.NewServeMux()

	// performs behavior for matched request and writes response
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Printf("writing response: %v", err)
		}
	})

	mux.HandleFunc("POST /jobs", func(w http.ResponseWriter, r *http.Request) {
		var request createJobRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		id, err := newID()
		if err != nil {
			http.Error(w, "could not generate job ID", http.StatusInternalServerError)
			return
		}

		now := time.Now()

		createdJob := job.Job{
			ID:        id,
			Type:      request.Type,
			Payload:   request.Payload,
			Status:    job.StatusQueued,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.Create(createdJob); err != nil {
			http.Error(w, "could not store job", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(createdJob); err != nil {
			log.Printf("encoding created job: %v", err)
		}
	})

	return mux
}
