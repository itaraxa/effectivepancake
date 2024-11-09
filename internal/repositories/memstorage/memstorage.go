package memstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/itaraxa/effectivepancake/internal/errors"
)

// Структура для хранения метрик в памяти
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	mu      sync.Mutex
}

/*
Update gauge value of metrica into MemStorage. If not exists - value will be created, else - updated

Args:

	metricName string: metrica name
	value float64: metrica value

Returns:

	error: nil or error of adding counter to the MemStorgage
*/
func (m *MemStorage) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Gauge[metricName] = value

	return nil
}

/*
Add counter value of metrica to MemStorage. If exists - value will be added

Args:

	metricName string: metrica name
	value int64: metrica value

Returns:

	error: nil or error of adding counter to the MemStorgage
*/
func (m *MemStorage) AddCounter(ctx context.Context, metricName string, delta int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Counter[metricName] += delta

	return nil
}

/*
GetMetrica Get metrica value grom MemStorage by metrica name

Args:

	metricaType string: type of requested metrica. Should be "gauge" or "counter"
	metricaName string: name of requested metrica

Returns:

	string: value of metrica in the string representation
	error: nil or error if metrica was not found in the MemStorage
*/
func (m *MemStorage) GetMetrica(ctx context.Context, metricaType string, metricaName string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch metricaType {
	case "gauge":
		if _, ok := m.Gauge[metricaName]; !ok {
			return nil, errors.ErrMetricaNotFaund
		}
		return m.Gauge[metricaName], nil

	case "counter":
		if _, ok := m.Counter[metricaName]; !ok {
			return nil, errors.ErrMetricaNotFaund
		}
		return m.Counter[metricaName], nil

	default:
		// case with uncorrect type of metrica
		return nil, errors.ErrMetricaNotFaund
	}
}

/*
GetAllMetrica return copy of data in memory storage
*/
func (m *MemStorage) GetAllMetrics(ctx context.Context) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{m.Gauge, m.Counter}
}

/*
Get string representation of the current state the MemStorage.

Returns:

	string: string representation of the current state the MemStorage

Output example:

	==== MemStorage ====
	       Gauge:
	Gauge1: 3.14
	Gauge2: 14.88
	Gauge3: 0.0
	      Counter:
	Counter1: 1
	Counter2: 42
*/
func (m *MemStorage) String(ctx context.Context) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := "==== MemStorage ====\n\r"
	s += "       Gauge:\n\r"
	for metric, value := range m.Gauge {
		s += fmt.Sprintf("%s: %g\n\r", metric, value)
	}
	s += "      Counter:\n\r"
	for metric, values := range m.Counter {
		s += fmt.Sprintf("%s: %d\n\r", metric, values)
	}
	return s
}

/*
Get html representation of the current state the MemStorage

Returns:

	string: html representation of the current state the MemStorage
*/
func (m *MemStorage) HTML(ctx context.Context) string {
	m.mu.Lock()
	defer m.mu.Unlock()

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
            background-color: #4CAF50;
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

	for metrica, value := range m.Gauge {
		h += fmt.Sprintf("<tr><td>%s</td><td>%g</td></tr>", metrica, value)
	}
	for metrica, value := range m.Counter {
		h += fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", metrica, value)
	}

	h += `        </tbody>
    </table>

</body>
</html>
`
	return h
}

/*
Create and return an instance of the memstorage object

Returns:

	*MemStorage: new instance of the MemStorage
*/
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *MemStorage) PingContext(ctx context.Context) error {
	return nil
}

func (m *MemStorage) Close() error {
	m.mu.Unlock()
	defer m.mu.Lock()
	clear(m.Gauge)
	clear(m.Counter)
	return nil
}
