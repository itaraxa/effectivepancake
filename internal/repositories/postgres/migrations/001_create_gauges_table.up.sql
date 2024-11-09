CREATE TABLE IF NOT EXISTS gauges (
    metric_id bigint PRIMARY KEY,
    metric_value double precision NOT NULL,
    metric_timestamp TIMESTAMPTZ NOT NULL);
