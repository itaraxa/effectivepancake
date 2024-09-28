package models

import (
	"fmt"
)

// Структура представляющая полученный запрос
type Query struct {
	raw            string
	MetricType     string
	MetricName     string
	MetricRawValue string
}

func (q Query) String() string {
	s := "==== Query ====\n\r"
	s += fmt.Sprintf(">Raw: %s\n\r", q.raw)
	s += fmt.Sprintf(">>Type: %s\n\r", q.MetricType)
	s += fmt.Sprintf(">>Name: %s\n\r", q.MetricName)
	s += fmt.Sprintf(">>Value: %s\n\r", q.MetricRawValue)

	return s
}

func (q Query) GetMetricaType() string {
	return q.MetricType
}

func (q Query) GetMetricName() string {
	return q.MetricName
}

func (q Query) GetMetricaRawValue() string {
	return q.MetricRawValue
}
