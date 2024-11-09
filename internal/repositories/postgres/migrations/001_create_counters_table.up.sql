CREATE TABLE IF NOT EXISTS counters (
    metric_id bigint PRIMARY KEY,
    metric_delta bigint NOT NULL,
    metric_timestamp TIMESTAMPTZ NOT NULL);
