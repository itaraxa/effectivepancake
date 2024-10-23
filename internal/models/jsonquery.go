package models

import "encoding/json"

type JSONQuery struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewJSONQuery() *JSONQuery {
	return &JSONQuery{}
}

func (jq JSONQuery) String() string {
	b, err := json.MarshalIndent(jq, "\t", "\t")
	if err != nil {
		return "{marshal error}"
	}
	return string(b)
}

func (jq JSONQuery) GetMetricaType() string {
	return jq.MType
}

func (jq JSONQuery) GetMetricaName() string {
	return jq.ID
}

func (jq JSONQuery) GetMetricaValue() interface{} {
	if jq.MType == `gauge` {
		return *jq.Value
	}
	if jq.MType == `counter` {
		return *jq.Delta
	}
	return nil
}
