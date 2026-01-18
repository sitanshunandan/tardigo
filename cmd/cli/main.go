package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
)

// Config: Where does the CLI look for the brain?
const API_URL = "http://127.0.0.1:8080"

// Task structure matches the API expectation
type Task struct {
	Name     string `json:"name"`
	Duration int    `json:"duration_minutes"`
	Effort   int    `json:"effort_level"`
}

// Response structures for parsing JSON
type ScheduleItem struct {
	StartTime    string  `json:"start_time"`
	TaskName     string  `json:"task_name"`
	PredictedCap float64 `json:"predicted_capacity"`
	FitScore     string  `json:"fit_score"`
}

type ScheduleResponse struct {
	Algorithm string         `json:"algorithm"`
	Schedule  []ScheduleItem `json:"schedule"`
}

type CapacityResponse struct {
	Status         string             `json:"status"`
	CapacityScore  float64            `json:"capacity_score"`
	Recommendation string             `json:"recommendation"`
	Components     map[string]float64 `json:"components"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "plan":
		handlePlan(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  tardigo status                  # Get current brain capacity")
	fmt.Println("  tardigo plan <name> <min> <1-10> # optimize a single task")
	fmt.Println("Example:")
	fmt.Println("  tardigo plan \"Learn Rust\" 60 9")
}

func handleStatus() {
	resp, err := http.Get(API_URL + "/capacity/now")
	if err != nil {
		fmt.Printf("Error connecting to Cortex: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var data CapacityResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Println("\n--- 🧠 Current Bio-State ---")
	fmt.Printf("Capacity:       %.2f (%.0f%%)\n", data.CapacityScore, data.CapacityScore*100)
	fmt.Printf("Freshness (S):  %.2f\n", data.Components["freshness"])
	fmt.Printf("Circadian (C):  %.2f\n", data.Components["circadian"])
	fmt.Printf("Advice:         %s\n", data.Recommendation)
	fmt.Println("---------------------------")
}

func handlePlan(args []string) {
	if len(args) < 3 {
		fmt.Println("Error: Missing arguments for plan.")
		printUsage()
		return
	}

	name := args[0]
	duration, _ := strconv.Atoi(args[1])
	effort, _ := strconv.Atoi(args[2])

	// Construct payload (List of 1 task for now)
	tasks := []Task{
		{Name: name, Duration: duration, Effort: effort},
	}

	jsonData, _ := json.Marshal(tasks)
	resp, err := http.Post(API_URL+"/schedule/optimize", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error calling scheduler: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var plan ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		fmt.Printf("Error parsing schedule: %v\n", err)
		return
	}

	fmt.Printf("\n--- 📅 Optimized Schedule (%s) ---\n", plan.Algorithm)

	// Use TabWriter for clean columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "START\tTASK\tCAPACITY\tFIT\t")
	fmt.Fprintln(w, "-----\t----\t--------\t---\t")

	for _, item := range plan.Schedule {
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%s\t\n",
			item.StartTime,
			item.TaskName,
			item.PredictedCap,
			item.FitScore,
		)
	}
	w.Flush()
	fmt.Println()
}
