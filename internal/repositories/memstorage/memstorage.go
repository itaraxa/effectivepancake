package memstorage

import (
	"fmt"
	"sync"
)

// Структура для хранения метрик в памяти
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string][]int64
	mu      sync.Mutex
}

func (m *MemStorage) UpdateGauge(metric_name string, value float64) error {
	return nil
}

func (m *MemStorage) AddCounter(metric_name string, value int64) error {
	return nil
}

func (m *MemStorage) GetGauge(metric_name string) (float64, error) {
	return 0.0, nil
}

func (m *MemStorage) GetCounter(metric_name string) ([]int64, error) {
	return []int64{}, nil
}

func (m *MemStorage) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := "==== MemStorage ====\n\r"
	s += "<Gauge:\n\r"
	for metric, value := range m.Gauge {
		s += fmt.Sprintf("<<%s: %f\n\r", metric, value)
	}
	s += "<Counter:\n\r"
	for metric, values := range m.Counter {
		s += fmt.Sprintf("<<%s: %v\n\r", metric, values)
	}
	return s
}
