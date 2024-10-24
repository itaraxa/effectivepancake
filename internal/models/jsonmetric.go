package models

import (
	"encoding/json"
	"fmt"
	"sync"
)

// type JSONMetrics struct {
// 	ID    string   `json:"id"`              // имя метрики
// 	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
// 	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
// 	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
// }

// func (jm *JSONMetrics) Marshal() ([]byte, error) {
// 	return nil, nil
// }

// func Unmarshal(data []byte, jm *JSONMetrics) error {
// 	return nil
// }

type JSONMetric struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type JSONMetrics struct {
	Data []JSONMetric
	mu   sync.Mutex
}

func (jms *JSONMetrics) AddData(data []JSONMetric) error {
	jms.mu.Lock()
	defer jms.mu.Unlock()

	jms.Data = append(jms.Data, data...)
	return nil
}

func (jms *JSONMetrics) AddPollCount(pollCount int64) error {
	jms.mu.Lock()
	defer jms.mu.Unlock()

	jms.Data = append(jms.Data, JSONMetric{
		ID:    "PollCount",
		MType: "counter",
		Delta: pollCount,
	})

	return nil
}

func (jms *JSONMetrics) GetData() []JSONMetric {
	jms.mu.Lock()
	defer jms.mu.Unlock()

	return jms.Data[:]
}

func (jms *JSONMetrics) String() string {
	jms.mu.Lock()
	defer jms.mu.Unlock()

	out := ""
	for _, jm := range jms.Data {
		t, _ := json.MarshalIndent(jm, ``, `    `)
		out += fmt.Sprintln(string(t))
	}
	return out
}
