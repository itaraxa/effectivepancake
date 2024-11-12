CREATE TABLE IF NOT EXISTS gauges (
    metric_id TEXT NOT NULL,
    metric_value double precision NOT NULL,
    metric_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);
