package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	// TODO: REPLACE THIS with your actual module name from go.mod
	"github.com/sitanshunandan/tardigo/internal/biomodel"
)

func main() {
	// 1. Initialize the MCP Server
	s := server.NewMCPServer(
		"TardiGo-Bio-scheduler",
		"1.0.0",
		server.WithLogging(),
	)

	// 2. Define the Tool Schema
	// We tell the AI exactly what "OptimizeSchedule" expects.
	scheduleTool := mcp.NewTool("plan_biological_schedule",
		mcp.WithDescription("Generates an optimal daily schedule based on biological energy (Circadian Rhythms). Use this to plan tasks when the user is most alert."),
		mcp.WithStringArgument("wake_time", "The time the user woke up today (RFC3339 format, e.g. 2024-02-17T07:00:00Z)."),
		// We define the complex 'tasks' array manually to ensure strict typing for the AI
	)

	// Manually adding the array schema for tasks since it's a complex object
	scheduleTool.InputSchema.Properties["tasks"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":             map[string]interface{}{"type": "string"},
				"duration_minutes": map[string]interface{}{"type": "integer"},
				"effort_level":     map[string]interface{}{"type": "integer", "description": "1-10 scale (10=Highest Effort)"},
			},
			"required": []string{"name", "duration_minutes", "effort_level"},
		},
	}

	// 3. Register the Handler
	s.AddTool(scheduleTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// A. Parse Arguments
		var args struct {
			WakeTime string          `json:"wake_time"`
			Tasks    []biomodel.Task `json:"tasks"`
		}
		if err := json.Unmarshal(request.Params.Arguments, &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
		}

		// B. Parse Time (Default to Now if invalid)
		wakeTime, err := time.Parse(time.RFC3339, args.WakeTime)
		if err != nil {
			return mcp.NewToolResultError("Invalid wake_time format. Use RFC3339."), nil
		}

		// C. Setup Biological Parameters (The "World Model")
		// In a real system, you might fetch these from your DB (internal/storage)
		bioParams := biomodel.BioParams{
			WakeTime:      wakeTime,
			ChronotypeLag: 0.0,  // Assuming normal chronotype for now
			FatigueRate:   16.0, // Standard fatigue rate
		}

		// D. Run the Heavy Lifting (Your Scheduler Logic)
		// Note: We start scheduling from "Now" or "WakeTime", depending on your logic.
		// Let's assume we start scheduling from the moment the user woke up + 1 hour.
		startSim := wakeTime.Add(1 * time.Hour)

		schedule := biomodel.OptimizeSchedule(args.Tasks, startSim, bioParams)

		// E. Return the Result as JSON
		responseBytes, _ := json.Marshal(schedule)
		return mcp.NewToolResultText(string(responseBytes)), nil
	})

	// 4. Start the Server (Stdio Mode)
	// This allows Claude Desktop or Python Scripts to talk to it via Standard Input/Output
	fmt.Println("Starting TardiGo MCP Server on Stdio...")
	if err := s.ServeStdio(); err != nil {
		panic(err)
	}
}
