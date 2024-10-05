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

/*
Get string representation instance of Query

Returns:

	string: string representation of the Query instance

Output example:

	==== Query ====
	>Raw: update/gauge/test1/3.14
	>>Type: gauge
	>>Name: test1
	>>Value: 3.14
*/
func (q Query) String() string {
	s := "==== Query ====\n\r"
	s += fmt.Sprintf(">Raw: %s\n\r", q.raw)
	s += fmt.Sprintf(">>Type: %s\n\r", q.MetricType)
	s += fmt.Sprintf(">>Name: %s\n\r", q.MetricName)
	s += fmt.Sprintf(">>Value: %s\n\r", q.MetricRawValue)

	return s
}

/*
Get metrica type

Returns:

	string: Type of metrica
*/
func (q Query) GetMetricaType() string {
	return q.MetricType
}

/*
Get metrica name

Returns:

	string: Name of metrica
*/
func (q Query) GetMetricName() string {
	return q.MetricName
}

/*
Get metrica raw value (string representation of value)

Returns:

	string: string representation of value
*/
func (q Query) GetMetricaRawValue() string {
	return q.MetricRawValue
}
