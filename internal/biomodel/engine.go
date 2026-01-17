package biomodel

import (
	"math"
	"time"
)

// BioParams represents the biological constants unique to a specific user.
// These would eventually be loaded from a DB, but for now, they are the model's configuration.
type BioParams struct {
	WakeTime      time.Time // The reference point for Homeostatic pressure (Process S)
	ChronotypeLag float64   // Shift in hours for Circadian Rhythm (Process C). e.g., 0 for normal, +2 for Night Owl.
	FatigueRate   float64   // Sensitivity to adenosine. Lower = faster fatigue. Typical range 14.0 - 18.0.
}

// BioState holds the calculated result of the model at a specific point in time.
type BioState struct {
	ProcessS      float64 // Sleep Pressure (0.0 to 1.0)
	ProcessC      float64 // Circadian Arousal (0.0 to 1.0)
	TotalCapacity float64 // The final available "Brain Battery" (0.0 to 1.0)
}

// CalculateState computes the biological capacity for a specific target time.
// It implements the Borbély Two-Process Model.
func (b *BioParams) CalculateState(targetTime time.Time) BioState {

	// --- Process S: The Homeostat (Sleep Pressure) ---
	// Formula: S(t) = 1 - e^(-t / tau)
	// As 'hoursAwake' increases, the result approaches 1.0 (maximum sleep pressure).
	hoursAwake := targetTime.Sub(b.WakeTime).Hours()
	if hoursAwake < 0 {
		hoursAwake = 0 // Handle edge case if checking time before wake
	}

	// We calculate "Freshness" as the inverse of pressure.
	// 1.0 = Fresh, 0.0 = Exhausted.
	sleepPressure := 1.0 - math.Exp(-hoursAwake/b.FatigueRate)
	processS_Freshness := 1.0 - sleepPressure

	// --- Process C: The Circadian Pacemaker ---
	// Formula: Sinusoidal wave representing the SCN drive.
	// We map the 24h cycle to 2*Pi radians.

	// Calculate hours since midnight for the target day
	timeOfDay := float64(targetTime.Hour()) + float64(targetTime.Minute())/60.0

	// Standard circadian peak is usually around late afternoon.
	// We use ChronotypeLag to shift the wave left or right.
	// The "- 6" shifts the standard sine wave so it starts rising in the morning.
	processC_Raw := math.Sin((2 * math.Pi / 24.0) * (timeOfDay - 6.0 - b.ChronotypeLag))

	// Normalize Process C from [-1, 1] to [0, 1]
	processC_Normalized := (processC_Raw + 1.0) / 2.0

	// --- Integration: The TardiGo Algorithm ---
	// We combine the Circadian drive with the Homeostatic freshness.
	// Heuristic: Capacity is the average of your freshness and your circadian drive.
	totalCapacity := (processS_Freshness + processC_Normalized) / 2.0

	return BioState{
		ProcessS:      processS_Freshness,
		ProcessC:      processC_Normalized,
		TotalCapacity: math.Max(0.0, math.Min(1.0, totalCapacity)), // Clamp between 0-1
	}
}

// Task represents a unit of work to be scheduled.
type Task struct {
	Name     string `json:"name"`
	Duration int    `json:"duration_minutes"` // e.g., 60
	Effort   int    `json:"effort_level"`     // 1-10 (10 = Hardest)
}

// ScheduleItem is a task assigned to a specific time slot.
type ScheduleItem struct {
	StartTime    string  `json:"start_time"`
	TaskName     string  `json:"task_name"`
	PredictedCap float64 `json:"predicted_capacity"`
	FitScore     string  `json:"fit_score"` // "Perfect", "Good", "Bad"
}
