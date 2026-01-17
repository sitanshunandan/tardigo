-- 1. Enable the TimescaleDB extension (should be on by default, but good practice)
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 2. Create the table for raw biological metrics
-- We use TEXT for user_id for simplicity, but UUID is better for prod.
CREATE TABLE IF NOT EXISTS bio_telemetry (
    time                TIMESTAMPTZ NOT NULL,
    user_id             TEXT NOT NULL,
    heart_rate          DOUBLE PRECISION,
    hrv                 DOUBLE PRECISION, -- Heart Rate Variability (ms)
    process_s           DOUBLE PRECISION, -- Sleep Pressure (0.0 - 1.0)
    process_c           DOUBLE PRECISION, -- Circadian Drive (0.0 - 1.0)
    overall_capacity    DOUBLE PRECISION  -- The Final TardiGo Score
);

-- 3. Convert standard table into a Hypertable
-- This tells TimescaleDB to partition this data by time for massive performance.
-- We check if it's already a hypertable to prevent errors on restart.
SELECT create_hypertable('bio_telemetry', 'time', if_not_exists => TRUE);

-- 4. Create an index for fast querying by user
CREATE INDEX IF NOT EXISTS idx_bio_telemetry_user ON bio_telemetry (user_id, time DESC);