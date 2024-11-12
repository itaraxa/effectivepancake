CREATE TABLE IF NOT EXISTS counters (
    metric_id TEXT NOT NULL,
    metric_delta bigint NOT NULL,
    metric_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);
