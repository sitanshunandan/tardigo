package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sitanshunandan/tardigo/internal/storage"
)

// Server struct to hold dependencies
type Server struct {
	repo *storage.TelemetryRepository
}

func main() {
	// 1. Setup Database Connection (Same as Simulator)
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), // In Docker, this will be "db"
		os.Getenv("DB_NAME"),
	)

	ctx := context.Background()
	repo, err := storage.NewTelemetryRepository(ctx, dbURL)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer repo.Close(ctx)

	srv := &Server{repo: repo}

	// 2. Setup Routes
	http.HandleFunc("/capacity/now", srv.HandleGetCurrentCapacity)

	// 3. Start Server
	port := ":8080"
	fmt.Printf("--- TardiGo Cortex Online. Listening on %s ---\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// HandleGetCurrentCapacity returns the latest bio-data for a user
func (s *Server) HandleGetCurrentCapacity(w http.ResponseWriter, r *http.Request) {
	// In a real app, we'd get this from a JWT token. For now, hardcode.
	userID := "user_001"

	// We need to implement a "GetLatest" method in our repo next!
	// For now, we return a placeholder to verify the server works.
	response := map[string]string{
		"status":  "online",
		"user":    userID,
		"message": "Brain-Computer Interface Ready",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
