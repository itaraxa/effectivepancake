package models

import (
	"fmt"
	"strings"
)

// Структура представляющая полученный запрос
type Query struct {
	raw            string
	MetricType     string
	MetricName     string
	MetricRawValue string
}

// Функция для создания query
func Parse(raw string) (q Query, err error) {
	q.raw = raw
	q.MetricType = strings.Split(q.raw, `/`)[2]
	q.MetricName = strings.Split(q.raw, `/`)[3]
	q.MetricRawValue = strings.Split(q.raw, `/`)[4]

	return q, nil
}

func (q Query) String() string {
	s := ">>>Query<<<\n\r"
	s += fmt.Sprintf("Raw: %s\n\r", q.raw)
	s += fmt.Sprintf("Type: %s\n\r", q.MetricType)
	s += fmt.Sprintf("Name: %s\n\r", q.MetricName)
	s += fmt.Sprintf("Value: %s\n\r", q.MetricRawValue)

	return s
}
