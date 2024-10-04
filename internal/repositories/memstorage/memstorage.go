package memstorage

import (
	"fmt"
	"sync"

	"github.com/itaraxa/effectivepancake/internal/errors"
)

// Структура для хранения метрик в памяти
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string][]int64
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

	m.Counter[metricName] = append(m.Counter[metricName], value)

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
		return fmt.Sprintf("%f", m.Gauge[metricaName]), nil

	case "counter":
		if _, ok := m.Counter[metricaName]; !ok {
			return "", errors.ErrMetricaNotFaund
		}
		return fmt.Sprintf("%v", m.Counter[metricaName]), nil

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
		s += fmt.Sprintf("<<%s: %f\n\r", metric, value)
	}
	s += "<Counter:\n\r"
	for metric, values := range m.Counter {
		s += fmt.Sprintf("<<%s: %v\n\r", metric, values)
	}
	return s
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string][]int64),
	}
}
