CREATE TABLE IF NOT EXISTS metrics (
    metric_id bigint PRIMARY KEY,
    metric_name varchar(250) NOT NULL,
    metric_type varchar(32) NOT NULL);
