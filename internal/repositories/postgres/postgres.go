package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type MetricStorager interface {
	MetricGetter
	MetricUpdater
	MetricPrinter
}

type MetricUpdater interface {
	UpdateGauge(ctx context.Context, metricName string, value float64) error
	AddCounter(ctx context.Context, metricName string, value int64) error
}

type MetricGetter interface {
	GetMetrica(ctx context.Context, metricaType string, metricaName string) (interface{}, error)
	GetAllMetrics(ctx context.Context) (interface{}, error)
}

type MetricPrinter interface {
	String() string
	HTML() string
}

type PostgresRepository struct {
	db *sql.DB
}

/*
NewPostgresRepository creates instance of PostgresRepository

Args:

	databaseURL: string for connection to databse, example: "postgres://username:password@localhost:5432/database_name"

Returns:

	dbStorager
	error
*/
func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return &PostgresRepository{}, err
	}
	return &PostgresRepository{db: db}, nil
}

/*
PingContext check connection to db

Args:

	ctx context.Context

Returns:

	error: nil or an error that occurred while processing the ping db
*/
func (pr *PostgresRepository) PingContext(ctx context.Context) error {
	if err := pr.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

/*
Close closes the open database connection

Returns:

	error: nil or an error that occurred while closing connection
*/
func (pr *PostgresRepository) Close() error {
	if err := pr.db.Close(); err != nil {
		return err
	}
	return nil
}

/*
PrepareTablesContext create tables for metrice storage if not exist

Args:

	ctx context.Context

Returns:

	error: nil or an error that occurred while processing the request
*/
func (pr *PostgresRepository) PrepareTablesContext(ctx context.Context) error {
	_, err := pr.db.ExecContext(ctx, "CREATE TABLE IF NOT EXIST metrics ('metric_id' bigint, 'metric_name' text, 'metric_type', text)")
	if err != nil {
		return err
	}

	_, err = pr.db.ExecContext(ctx, "CREATE TABLE IF NOT EXIST gauges ('metric_id' bigint, 'value' double precision, 'timestamp' timestamptz)")
	if err != nil {
		return err
	}

	_, err = pr.db.ExecContext(ctx, "CREATE TABLE IF NOT EXIST counters ('metric_id' bigint, 'delta' bigint, 'timestamp' timestamptz)")
	if err != nil {
		return err
	}
	return nil
}

/*
UpdateGauge add gauge into db storage. Adding a timestamp to determine the most recent value

Args:

	ctx context.Context
	metricName string: unique identifier for the metric
	value float64: gauge value

Returns:

	error
*/
func (pr *PostgresRepository) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	_, err := pr.db.ExecContext(ctx, "INSERT INTO gauges(id, value, timestamp) VALUES (?, ?, ?)", metricName, value, time.Now())
	if err != nil {
		return err
	}
	return nil
}

/*
UpdateCounter add counter into db storage. Adding a timestamp to determine the most recent value

Args:

	ctx context.Context
	metricName string: unique identifier for the metric
	value int64: counter value

Returns:

	error
*/
func (pr *PostgresRepository) AddCounter(ctx context.Context, metricName string, delta int64) (err error) {
	// init transaction
	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction: %w", err)
	}
	// rollback on error and commit if all ok
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// check delta-value in db
	var currentDelta int64
	err = tx.QueryRowContext(ctx, "SELECT delta FROM counters WHERE id = $1", metricName).Scan(&currentDelta)
	if err != nil {
		// delta-value not in db
		if err == sql.ErrNoRows {
			_, err = tx.ExecContext(ctx, "INSERT INTO counters(id, delta, timestamp) VALUES ($1, $2, $3)", metricName, delta, time.Now())
			if err != nil {
				return fmt.Errorf("cannot insert new record: %w", err)
			}
			// delta-value in db
		} else {
			newDelta := currentDelta + delta
			_, err = tx.ExecContext(ctx, "INSERT INTO counters(id, delta, timestamp) VALUES ($1, $2, $3)", metricName, newDelta, time.Now())
			if err != nil {
				return fmt.Errorf("cannot update record: %w", err)
			}
		}
	}

	return nil
}

func (pr *PostgresRepository) GetMetrica(ctx context.Context, metricaType string, metricaName string) (interface{}, error) {
	switch metricaType {
	case `gauge`:
		row := pr.db.QueryRowContext(ctx, "SELCET value FROM gauges WHERE id = $1 AND timestamp = (SELECT MAX(timestamp FROM gauges));", metricaName)
		var gauge sql.NullFloat64
		err := row.Scan(&gauge)
		if err != nil {
			return nil, fmt.Errorf("cannot get gauge value from db: %w", err)
		}
		if !gauge.Valid {
			return nil, fmt.Errorf("empty gauge value in db")
		}
		return gauge.Float64, nil
	case `counter`:
		row := pr.db.QueryRowContext(ctx, "SELCET value FROM counters WHERE id = $1 AND timestamp = (SELECT MAX(timestamp FROM counters));", metricaName)
		var delta sql.NullInt64
		err := row.Scan(delta)
		if err != nil {
			return nil, fmt.Errorf("cannot get counter value from db: %w", err)
		}
		if !delta.Valid {
			return nil, fmt.Errorf("empty counter value in db")
		}
		return delta.Int64, nil
	default:
		return nil, fmt.Errorf("unknown metrica type: %s", metricaType)
	}
}

// In progress
func (pr *PostgresRepository) GetAllMetrics(ctx context.Context) interface{} {
	return nil
}

func (pr *PostgresRepository) String(ctx context.Context) string {
	return ""
}

func (pr *PostgresRepository) HTML(ctx context.Context) string {
	return ""
}
