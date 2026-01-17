package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sitanshunandan/tardigo/internal/biomodel"
)

// TelemetryRepository handles all database interactions for bio-data.
type TelemetryRepository struct {
	conn *pgx.Conn
}

// NewTelemetryRepository creates a connection to the database.
// It includes a "Retry Loop" because in Docker, the App often starts before the DB is ready.
func NewTelemetryRepository(ctx context.Context, dbURL string) (*TelemetryRepository, error) {
	var conn *pgx.Conn
	var err error

	// Retry logic: Try to connect 5 times with a 2-second delay
	for i := 0; i < 5; i++ {
		conn, err = pgx.Connect(ctx, dbURL)
		if err == nil {
			break
		}
		fmt.Printf("Database not ready yet... retrying (%d/5)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &TelemetryRepository{conn: conn}, nil
}

// Close closes the connection.
func (r *TelemetryRepository) Close(ctx context.Context) {
	r.conn.Close(ctx)
}

// Save inserts a new biological state into the database.
func (r *TelemetryRepository) Save(ctx context.Context, userID string, timestamp time.Time, state biomodel.BioState) error {
	query := `
		INSERT INTO bio_telemetry (time, user_id, process_s, process_c, overall_capacity)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.conn.Exec(ctx, query, timestamp, userID, state.ProcessS, state.ProcessC, state.TotalCapacity)
	return err
}

// GetLatestCapacity fetches the most recent bio-state for a user.
func (r *TelemetryRepository) GetLatestCapacity(ctx context.Context, userID string) (*biomodel.BioState, error) {
	query := `
		SELECT process_s, process_c, overall_capacity 
		FROM bio_telemetry 
		WHERE user_id = $1 
		ORDER BY time DESC 
		LIMIT 1
	`

	var state biomodel.BioState
	// We scan the database columns directly into our Go struct fields
	err := r.conn.QueryRow(ctx, query, userID).Scan(
		&state.ProcessS,
		&state.ProcessC,
		&state.TotalCapacity,
	)

	if err != nil {
		return nil, err
	}

	return &state, nil
}
