package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	gauge   = `gauge`
	counter = `counter`
)

// type metricStorager interface {
// 	metricGetter
// 	metricUpdater
// 	metricPrinter
// 	PingContext(context.Context) error
// }

// type metricUpdater interface {
// 	UpdateGauge(context.Context, string, float64) error
// 	AddCounter(context.Context, string, int64) error
// }

// type metricGetter interface {
// 	GetMetrica(ctx context.Context, metricaType string, metricaName string) (interface{}, error)
// 	GetAllMetrics(ctx context.Context) (interface{}, error)
// }

// type metricPrinter interface {
// 	String() string
// 	HTML() string
// }

/*
PostgresRepository is the struct for wrapping PostgreSQL storage
*/
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = prepareTablesContext(ctx, db)
	if err != nil {
		return &PostgresRepository{}, fmt.Errorf("cannot create tables in database storage: %w", err)
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
UpdateGauge add gauge into db storage. Adding a timestamp to determine the most recent value

Args:

	ctx context.Context
	metricName string: unique identifier for the metric
	value float64: gauge value

Returns:

	error
*/
func (pr *PostgresRepository) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	_, err := pr.db.ExecContext(ctx, "INSERT INTO gauges (metric_id, metric_value) VALUES ($1, $2);", metricName, value)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (pr *PostgresRepository) UpdateBatchGauge(ctx context.Context, metrics map[string]*float64) error {
	// init transacion
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

	for metricName, value := range metrics {
		_, err := pr.db.ExecContext(ctx, "INSERT INTO gauges (metric_id, metric_value) VALUES ($1, $2);", metricName, *value)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	return nil
}

/*
Update add counter into db storage. Adding a timestamp to determine the most recent value

Args:

	ctx context.Context
	metricName string: unique identifier for the metric
	value int64: counter value

Returns:

	error
*/
func (pr *PostgresRepository) AddCounter(ctx context.Context, metricName string, delta int64) error {
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
	err = tx.QueryRowContext(ctx, "SELECT metric_delta FROM counters WHERE metric_id = $1 ORDER BY metric_timestamp DESC LIMIT 1;", metricName).Scan(&currentDelta)
	if err != nil {
		// delta-value not in db
		if err == sql.ErrNoRows {
			_, err = tx.ExecContext(ctx, "INSERT INTO counters (metric_id, metric_delta, metric_timestamp) VALUES ($1, $2, $3);", metricName, delta, time.Now())
			if err != nil {
				return fmt.Errorf("cannot insert new record: %w", err)
			}
		} else {
			return fmt.Errorf("cannot check counter delta in DB: %w", err)
		}
	}
	// delta-value in db
	newDelta := currentDelta + delta
	_, err = tx.ExecContext(ctx, "INSERT INTO counters (metric_id, metric_delta, metric_timestamp) VALUES ($1, $2, $3);", metricName, newDelta, time.Now())
	if err != nil {
		return fmt.Errorf("cannot update record: %w", err)
	}

	return nil
}

func (pr *PostgresRepository) AddBatchCounter(ctx context.Context, metrics map[string]*int64) error {
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

	for metricName, delta := range metrics {
		// check delta-value in db
		var currentDelta int64
		err = tx.QueryRowContext(ctx, "SELECT metric_delta FROM counters WHERE metric_id = $1 ORDER BY metric_timestamp DESC LIMIT 1;", metricName).Scan(&currentDelta)
		if err != nil {
			// delta-value not in db
			if err == sql.ErrNoRows {
				_, err = tx.ExecContext(ctx, "INSERT INTO counters (metric_id, metric_delta, metric_timestamp) VALUES ($1, $2, $3);", metricName, *delta, time.Now())
				if err != nil {
					return fmt.Errorf("cannot insert new record: %w", err)
				}
			} else {
				return fmt.Errorf("cannot check counter delta in DB: %w", err)
			}
		}
		// delta-value in db
		newDelta := currentDelta + *delta
		_, err = tx.ExecContext(ctx, "INSERT INTO counters (metric_id, metric_delta, metric_timestamp) VALUES ($1, $2, $3);", metricName, newDelta, time.Now())
		if err != nil {
			return fmt.Errorf("cannot update record: %w", err)
		}

	}
	return nil
}

/*
GetMetrica return value of requested metrica

Args:

	ctx context.Context
	metricaType string: type of requested metrica
	metricaName string: name of requested metrica

Returns:

	interface{}: value of requested metrica, float64 for gauge or int64 for counter
	error: nil or error, if value cannot be getted
*/
func (pr *PostgresRepository) GetMetrica(ctx context.Context, metricaType string, metricaName string) (interface{}, error) {
	switch metricaType {
	case gauge:
		SQL := `SELECT metric_value FROM gauges WHERE metric_id = $1 ORDER BY metric_timestamp DESC LIMIT 1;`
		row := pr.db.QueryRowContext(ctx, SQL, metricaName)
		var gauge sql.NullFloat64
		err := row.Scan(&gauge)
		if err != nil {
			return nil, fmt.Errorf("cannot get gauge value from db: %w", err)
		}
		if !gauge.Valid {
			return nil, fmt.Errorf("empty gauge value in db")
		}
		return gauge.Float64, nil
	case counter:
		SQL := `SELECT metric_delta FROM counters WHERE metric_id = $1 ORDER BY metric_timestamp DESC LIMIT 1;`
		row := pr.db.QueryRowContext(ctx, SQL, metricaName)
		var delta sql.NullInt64
		err := row.Scan(&delta)
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

/*
GetAllMetrics returns last values of gauges and last delta of counters

Args:

	ctx context.Context

Returns:

	interface{}
	error
*/
func (pr *PostgresRepository) GetAllMetrics(ctx context.Context) (interface{}, error) {
	var name sql.NullString
	var value sql.NullFloat64
	var delta sql.NullInt64
	Gauges := make(map[string]float64)
	Counters := make(map[string]int64)

	// Getting gauges
	gaugesSQLString := `SELECT DISTINCT ON (metric_id) metric_id, metric_value FROM gauges ORDER BY metric_id, metric_timestamp DESC;`
	gaugesRows, err := pr.db.QueryContext(ctx, gaugesSQLString)
	if err != nil {
		return nil, err
	}
	defer gaugesRows.Close()

	for gaugesRows.Next() {
		err = gaugesRows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		if name.Valid && value.Valid {
			Gauges[name.String] = value.Float64
		}
	}

	err = gaugesRows.Err()
	if err != nil {
		return nil, err
	}

	// Getting counters
	countersSQLString := `SELECT DISTINCT ON (metric_id) metric_id, metric_delta FROM counters ORDER BY metric_id, metric_timestamp DESC;`
	countersRows, err := pr.db.QueryContext(ctx, countersSQLString)
	if err != nil {
		return nil, err
	}
	defer countersRows.Close()

	for countersRows.Next() {
		err = countersRows.Scan(&name, &delta)
		if err != nil {
			return nil, err
		}
		if name.Valid && delta.Valid {
			Counters[name.String] = delta.Int64
		}
	}

	err = countersRows.Err()
	if err != nil {
		return nil, err
	}

	// return all metric
	return struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{Gauges, Counters}, nil
}

func (pr *PostgresRepository) String(ctx context.Context) string {
	s := ""
	metrics, err := pr.GetAllMetrics(ctx)
	if err != nil {
		fmt.Printf("error getting metrics from postgres storage: %v", err.Error())
		return ""
	}
	gauges := metrics.(struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}).Gauges
	counters := metrics.(struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}).Counters

	s += "Gauges:"
	for metricName, metricValue := range gauges {
		s += fmt.Sprintf("%s: %g\n\r", metricName, metricValue)
	}
	s += "Counters:"
	for metricName, metricDelta := range counters {
		s += fmt.Sprintf("%s: %d\n\r", metricName, metricDelta)
	}
	return ""
}

/*
HTML returns html-view of postgres metric storage

Args:

	ctx context.Context

Returns:

	stirng
*/
func (pr *PostgresRepository) HTML(ctx context.Context) string {
	h := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics Table</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f9;
            margin: 40px;
        }
        table {
            width: 70%;
            margin: 0 auto;
            border-collapse: collapse;
            background-color: #fff;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #ab7416;
            color: white;
        }
        tr:hover {
            background-color: #f1f1f1;
        }
    </style>
</head>
<body>

    <h2 style="text-align:center;">Metrics Table</h2>

    <table>
        <thead>
            <tr>
                <th>Metric Name</th>
                <th>Metric Value</th>
            </tr>
        </thead>
        <tbody>`

	metrics, err := pr.GetAllMetrics(ctx)
	if err != nil {
		fmt.Printf("error getting metrics from postgres storage: %v", err.Error())
		return ""
	}
	gauges := metrics.(struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}).Gauges
	counters := metrics.(struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}).Counters

	for metricaName, metricaValue := range gauges {
		h += fmt.Sprintf("<tr><td>%s</td><td>%g</td></tr>", metricaName, metricaValue)
	}
	for metricaName, metricaDelta := range counters {
		h += fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", metricaName, metricaDelta)
	}

	h += `        </tbody>
    </table>

</body>
</html>
`
	return h
}
