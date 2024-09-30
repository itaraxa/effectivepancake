package models

import (
	"fmt"
	"sync"
)

type Metrics struct {
	Data []Metrica
	mu   sync.Mutex
}

type Metrica struct {
	Type  string
	Name  string
	Value string
}

func (ms *Metrics) String() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	out := "==== Metrics ====\n\r"
	for _, m := range ms.Data {
		out += fmt.Sprintf(">>%s(%s) = %s\n\r", m.Name, m.Type, m.Value)
	}

	return out
}

func (ms *Metrics) AddData(data []Metrica) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Data = append(ms.Data, data...)
	return nil
}

func (ms *Metrics) AddPollCount(pollCount uint64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Data = append(ms.Data, Metrica{
		Name:  "PollCount",
		Type:  "counter",
		Value: fmt.Sprintf("%d", pollCount),
	})

	return nil
}
