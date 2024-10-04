package memstorage

import (
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

func (m *MemStorage) UpdateGauge(metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Gauge[metricName] = value

	return nil
}

func (m *MemStorage) AddCounter(metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Counter[metricName] += value

	return nil
}

func (m *MemStorage) GetMetrica(metricaType string, metricaName string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch metricaType {
	case "gauge":
		if _, ok := m.Gauge[metricaName]; !ok {
			return "", errors.ErrMetricaNotFaund
		}
		return fmt.Sprintf("%g", m.Gauge[metricaName]), nil

	case "counter":
		if _, ok := m.Counter[metricaName]; !ok {
			return "", errors.ErrMetricaNotFaund
		}
		return fmt.Sprintf("%d", m.Counter[metricaName]), nil

	default:
		return "", errors.ErrMetricaNotFaund
	}
}

func (m *MemStorage) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := "==== MemStorage ====\n\r"
	s += "<Gauge:\n\r"
	for metric, value := range m.Gauge {
		s += fmt.Sprintf("<<%s: %g\n\r", metric, value)
	}
	s += "<Counter:\n\r"
	for metric, values := range m.Counter {
		s += fmt.Sprintf("<<%s: %v\n\r", metric, values)
	}
	return s
}

func (m *MemStorage) HTML() string {
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

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}
