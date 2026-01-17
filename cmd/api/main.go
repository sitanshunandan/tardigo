package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sitanshunandan/tardigo/internal/biomodel"
	"github.com/sitanshunandan/tardigo/internal/storage"
)

// Server struct to hold dependencies
type Server struct {
	repo *storage.TelemetryRepository
}

func main() {
	// 1. Setup Database Connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbHost,
		os.Getenv("DB_NAME"),
	)

	// Local fallback
	if os.Getenv("DB_USER") == "" {
		dbURL = "postgres://postgres:password@localhost:5432/tardigo"
	}

	ctx := context.Background()
	repo, err := storage.NewTelemetryRepository(ctx, dbURL)
	if err != nil {
		log.Printf("WARNING: Could not connect to DB (%v). Running in offline mode.\n", err)
	} else {
		defer repo.Close(ctx)
	}

	srv := &Server{repo: repo}

	// 2. Setup Routes
	// GET: Status Check
	http.HandleFunc("/capacity/now", srv.HandleGetCurrentCapacity)
	// POST: The Intelligence Engine (NEW)
	http.HandleFunc("/schedule/optimize", srv.HandleOptimizeSchedule)

	// 3. Start Server
	port := ":8080"
	fmt.Printf("--- TardiGo Cortex Online. Listening on %s ---\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// HandleGetCurrentCapacity (GET) - Existing Logic
func (s *Server) HandleGetCurrentCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := "user_001"
	state, err := s.repo.GetLatestCapacity(r.Context(), userID)
	if err != nil {
		http.Error(w, "Biological signal lost: "+err.Error(), http.StatusNotFound)
		return
	}

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

// HandleOptimizeSchedule (POST) - NEW Logic
func (s *Server) HandleOptimizeSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// A. Parse the Incoming Tasks
	var tasks []biomodel.Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// B. Setup User Bio-Params
	// In a real app, we would fetch these "Settings" from the DB.
	// For now, we assume the user woke up at 7 AM today.
	now := time.Now()
	wakeTime := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())

	params := biomodel.BioParams{
		WakeTime:      wakeTime,
		ChronotypeLag: 0.0,
		FatigueRate:   16.0,
	}

	// C. Run the Algorithm
	// We schedule starting from the current hour
	schedule := biomodel.OptimizeSchedule(tasks, now, params)

	// D. Return the Plan
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"algorithm": "TardiGo-Greedy-v1",
		"schedule":  schedule,
	})
}
