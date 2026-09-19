package main

import (
	"log"
	"net/http"
)

func main() {
	// create a new server
	server := &http.Server{
		Addr:    ":8080",
		Handler: newHandler(),
	}

	log.Println("server listening on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// construct and return the router
// this isolates the router when you call it for testing purposes
func newHandler() http.Handler {
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

	return mux
}
