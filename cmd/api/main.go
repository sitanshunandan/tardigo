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
	// In the future, this comes from the URL or Auth Token
	userID := "user_001"

	// 1. Fetch real data from the Brain (DB)
	state, err := s.repo.GetLatestCapacity(r.Context(), userID)
	if err != nil {
		// If no data exists yet (or DB error)
		http.Error(w, "Biological signal lost: "+err.Error(), http.StatusNotFound)
		return
	}

	// 2. Construct the Response
	// We add "Recommendation" logic here to show System Design maturity.
	// We separate "Data" (the number) from "Insight" (the string).
	var recommendation string
	if state.TotalCapacity > 0.8 {
		recommendation = "Deep Work / High Focus (Coding, Math)"
	} else if state.TotalCapacity < 0.3 {
		recommendation = "Rest / Recovery (NSDR, Sleep)"
	} else {
		recommendation = "Admin / Low Stakes (Email, Meetings)"
	}

	response := map[string]interface{}{
		"user":           userID,
		"status":         "connected",
		"capacity_score": state.TotalCapacity,
		"components": map[string]float64{
			"freshness": state.ProcessS,
			"circadian": state.ProcessC,
		},
		"recommendation": recommendation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
