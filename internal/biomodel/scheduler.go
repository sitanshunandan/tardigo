package biomodel

import (
	"sort"
	"time"
)

// Slot represents a 30-minute window of availability
type Slot struct {
	Time     time.Time
	Capacity float64
	IsBooked bool
}

// OptimizeSchedule takes tasks and future capacity, and returns a calendar.
func OptimizeSchedule(tasks []Task, startHour time.Time, bioParams BioParams) []ScheduleItem {
	// 1. Generate Slots for the next 12 hours (30 min chunks)
	var slots []Slot
	for i := 0; i < 24; i++ { // 12 hours * 2 slots/hr
		t := startHour.Add(time.Duration(i*30) * time.Minute)
		state := bioParams.CalculateState(t)
		slots = append(slots, Slot{
			Time:     t,
			Capacity: state.TotalCapacity,
			IsBooked: false,
		})
	}

	// 2. Sort Tasks: Hardest tasks first! (Heuristic: First Fit Descending)
	// We want to book the "Deep Work" before the "Emails".
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Effort > tasks[j].Effort
	})

	var schedule []ScheduleItem

	// 3. The Allocation Loop
	for _, task := range tasks {
		slotsNeeded := task.Duration / 30
		bestStartIdx := -1
		bestScore := -1.0

		// Find the sequence of slots that maximizes (Capacity - Effort Match)
		// We want High Capacity for High Effort.
		for i := 0; i <= len(slots)-slotsNeeded; i++ {

			// Check if slots are free
			available := true
			avgCap := 0.0
			for j := 0; j < slotsNeeded; j++ {
				if slots[i+j].IsBooked {
					available = false
					break
				}
				avgCap += slots[i+j].Capacity
			}
			avgCap /= float64(slotsNeeded)

			if available {
				// Simple scoring: Higher capacity is better for hard tasks
				if avgCap > bestScore {
					bestScore = avgCap
					bestStartIdx = i
				}
			}
		}

		// 4. Book the slots
		if bestStartIdx != -1 {
			for j := 0; j < slotsNeeded; j++ {
				slots[bestStartIdx+j].IsBooked = true
			}

			// Add to schedule
			schedule = append(schedule, ScheduleItem{
				StartTime:    slots[bestStartIdx].Time.Format("15:04"),
				TaskName:     task.Name,
				PredictedCap: bestScore,
				FitScore:     judgeFit(task.Effort, bestScore),
			})
		} else {
			// Handle un-bookable task (e.g., day is full)
			schedule = append(schedule, ScheduleItem{
				StartTime: "UNSCHEDULED",
				TaskName:  task.Name,
				FitScore:  "No Time/Energy",
			})
		}
	}

	// Sort schedule by time for readability
	sort.Slice(schedule, func(i, j int) bool {
		return schedule[i].StartTime < schedule[j].StartTime
	})

	return schedule
}

func judgeFit(effort int, capacity float64) string {
	// Normalize effort 1-10 to 0.1-1.0
	normalizedEffort := float64(effort) / 10.0
	diff := capacity - normalizedEffort

	if diff >= -0.1 {
		return "Perfect" // Capacity meets or exceeds effort
	} else if diff > -0.3 {
		return "Challenging" // Capacity is slightly lower than needed
	}
	return "Burnout Risk" // High effort in low capacity slot
}
