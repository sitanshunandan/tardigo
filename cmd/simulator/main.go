package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sitanshunandan/tardigo/internal/biomodel"
	"github.com/sitanshunandan/tardigo/internal/storage"
)

func main() {
	fmt.Println("--- TardiGo: Starting Bio-Ingestion Pipeline ---")

	// 1. Configuration
	// We read DB credentials from Environment Variables (set in docker-compose)
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	// 2. Connect to Database (with Retry Logic)
	ctx := context.Background()
	repo, err := storage.NewTelemetryRepository(ctx, dbURL)
	if err != nil {
		fmt.Printf("CRITICAL: Could not connect to DB: %v\n", err)
		os.Exit(1) // Crash the container if DB fails
	}
	defer repo.Close(ctx)
	fmt.Println("SUCCESS: Connected to Time-Series Database")

	// 3. Setup Bio-Model
	now := time.Now()
	// Simulating a user who woke up at 7 AM
	wakeTime := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())

	params := biomodel.BioParams{
		WakeTime:      wakeTime,
		ChronotypeLag: 0.0,
		FatigueRate:   16.0,
	}

	userID := "user_001" // hardcoded for simulation

	// 4. The Loop: Generate & Ingest Data
	fmt.Println(">>> Ingesting 24 hours of biometric data...")

	for i := 0; i < 24; i++ {
		// Simulate time moving forward hour by hour
		simTime := wakeTime.Add(time.Duration(i) * time.Hour)

		// A. Calculate Logic
		state := params.CalculateState(simTime)

		// B. Persistence Logic (The new part!)
		err := repo.Save(ctx, userID, simTime, state)
		if err != nil {
			fmt.Printf("ERROR writing data: %v\n", err)
		} else {
			// Visual feedback in logs
			fmt.Printf("[SAVED] %s | Capacity: %.2f\n", simTime.Format("15:04"), state.TotalCapacity)
		}
	}

	fmt.Println("--- Ingestion Complete. Exiting... ---")
	// Keep container alive so we can check logs
}
