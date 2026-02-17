package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	// TODO: Verify this matches your go.mod module name exactly
	"github.com/YOUR_USERNAME/tardigo/internal/biomodel"
)

func main() {
	// 1. Initialize the MCP Server
	s := server.NewMCPServer(
		"TardiGo-Bio-scheduler",
		"1.0.0",
		server.WithLogging(),
	)

	// 2. Define the Tool
	// We use the basic constructor and then manually enhance the schema for the complex 'tasks' array
	scheduleTool := mcp.NewTool("plan_biological_schedule",
		mcp.WithDescription("Generates an optimal daily schedule based on biological energy. Use this to plan tasks."),
		mcp.WithString("wake_time",
			mcp.Required(),
			mcp.Description("The time the user woke up today (RFC3339 format, e.g. 2026-02-17T07:00:00Z)."),
		),
	)

	// Manually inject the complex array schema for 'tasks'
	// The library's helpers are great for simple fields, but for a []Struct, this is cleaner.
	scheduleTool.InputSchema.Properties["tasks"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":             map[string]interface{}{"type": "string"},
				"duration_minutes": map[string]interface{}{"type": "integer"},
				"effort_level":     map[string]interface{}{"type": "integer", "description": "1-10 scale"},
			},
			"required": []string{"name", "duration_minutes", "effort_level"},
		},
	}
	// Add 'tasks' to the required list
	scheduleTool.InputSchema.Required = append(scheduleTool.InputSchema.Required, "tasks")

	// 3. Register the Handler
	s.AddTool(scheduleTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// A. Parse Arguments safely
		// request.Params.Arguments is a map[string]interface{}.
		// We marshal it back to JSON to unmarshal it into our strong Go struct.
		jsonArgs, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse arguments: %v", err)), nil
		}

		var args struct {
			WakeTime string          `json:"wake_time"`
			Tasks    []biomodel.Task `json:"tasks"`
		}

		if err := json.Unmarshal(jsonArgs, &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments structure: %v", err)), nil
		}

		// B. Parse Time
		wakeTime, err := time.Parse(time.RFC3339, args.WakeTime)
		if err != nil {
			return mcp.NewToolResultError("Invalid wake_time format. Use RFC3339 (e.g., 2026-02-17T07:00:00Z)."), nil
		}

		// C. Setup World Model
		bioParams := biomodel.BioParams{
			WakeTime:      wakeTime,
			ChronotypeLag: 0.0,
			FatigueRate:   16.0,
		}

		// D. Run Scheduler
		// Schedule starting 1 hour after wake time
		startSim := wakeTime.Add(1 * time.Hour)
		schedule := biomodel.OptimizeSchedule(args.Tasks, startSim, bioParams)

		// E. Return Result
		responseBytes, _ := json.MarshalIndent(schedule, "", "  ")
		return mcp.NewToolResultText(string(responseBytes)), nil
	})

	// 4. Start the Server (Stdio Mode)
	// Corrected API call: server.ServeStdio(s)
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
