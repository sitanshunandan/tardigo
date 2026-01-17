TardiGo

TardiGo is a task scheduler written in Go that prioritizes tasks based on estimated metabolic capacity rather than just available time slots. It implements the Borbély Two-Process Model to simulate biological constraints and uses a greedy optimization algorithm to assign high-effort tasks to windows of peak cognitive performance.
Biological Model

The system calculates a "Cognitive Capacity Score" using three primary inputs:

    Process S (Homeostatic Sleep Pressure): Models the exponential decay of energy relative to hours awake.

    Process C (Circadian Rhythm): Models the sinusoidal oscillation of the Suprachiasmatic Nucleus (wakefulness drive).

    Vagal Tone: Uses simulated Heart Rate Variability (HRV) to apply dynamic penalties during high-stress intervals.

Architecture

The system uses a microservices approach with a clean architecture pattern (Handler -> Service -> Repository).
Code snippet

graph TD
    A[Bio-Simulator] -->|Generates HRV/Sleep Data| B(TimescaleDB)
    B -->|Time-Series Data| C[Cortex API]
    D[User Client] -->|POST Tasks| C
    C -->|Optimized Schedule| D

    Bio-Simulator: A background service that generates biological signals (HRV, Sleep) and writes them to the time-series database.

    TimescaleDB: PostgreSQL extension optimized for storing high-frequency biological time-series data.

    Cortex API: REST API that executes the scheduling algorithm. It retrieves current capacity state and maps tasks weighted by cognitive load.

Tech Stack

    Language: Go 1.25

    Database: TimescaleDB (PostgreSQL 14)

    Infrastructure: Docker & Docker Compose

Installation & Usage
1. Start the System

Clone the repository and start the services using Docker Compose.
Bash

git clone https://github.com/sitanshunandan/tardigo.git
cd tardigo
docker-compose up --build

2. Check Biological Status

Verify that the simulator is generating data and the API is calculating capacity.
Bash

curl http://localhost:8080/capacity/now

3. Generate a Schedule

Send a POST request with a list of tasks. The effort_level should be an integer between 1 (low) and 10 (high).
Bash

curl -X POST http://localhost:8080/schedule/optimize \
  -H "Content-Type: application/json" \
  -d '[
    { "name": "Deep Work (Coding)", "duration_minutes": 60, "effort_level": 9 },
    { "name": "Email Triage", "duration_minutes": 30, "effort_level": 3 }
  ]'

Roadmap

    [ ] Integration with Apple Health or Oura Ring API for real biological data.

    [ ] CLI tool for terminal-based scheduling.

    [ ] Panic threshold logic to auto-clear schedules when HRV drops below defined limits.