package models

// import (
// 	"fmt"
// 	"sync"
// )

// type Metrics struct {
// 	Data []Metric
// 	mu   sync.Mutex
// }

// type Metric struct {
// 	MType string
// 	ID    string
// 	Value string
// }

// func (ms *Metrics) String() string {
// 	ms.mu.Lock()
// 	defer ms.mu.Unlock()

// 	out := "==== Metrics ====\n\r"
// 	for _, m := range ms.Data {
// 		out += fmt.Sprintf(">>%s(%s) = %s\n\r", m.ID, m.MType, m.Value)
// 	}

// 	return out
// }

// func (ms *Metrics) AddData(data []Metric) error {
// 	ms.mu.Lock()
// 	defer ms.mu.Unlock()

// 	ms.Data = append(ms.Data, data...)
// 	return nil
// }

// func (ms *Metrics) AddPollCount(pollCount uint64) error {
// 	ms.mu.Lock()
// 	defer ms.mu.Unlock()

// 	ms.Data = append(ms.Data, Metric{
// 		ID:    "PollCount",
// 		MType: "counter",
// 		Value: fmt.Sprintf("%d", pollCount),
// 	})

// 	return nil
// }

// func (ms *Metrics) GetData() []Metric {
// 	ms.mu.Lock()
// 	defer ms.mu.Unlock()

// 	return ms.Data[:]
// }
