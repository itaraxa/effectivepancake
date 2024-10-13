package models

import (
	"fmt"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/errors"
)

// Структура представляющая полученный запрос
type Query struct {
	raw            string
	MetricType     string
	MetricName     string
	MetricRawValue string
}

func NewQuery() *Query {
	return &Query{}
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
func (q *Query) String() string {
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

func (q *Query) SetMetricaType(queryString string) error {
	q.MetricType = strings.Split(queryString, `/`)[1]
	if q.MetricType != "gauge" && q.MetricType != "counter" {
		return errors.ErrBadType
	}
	return nil
}

/*
Get metrica name

Returns:

	string: Name of metrica
*/
func (q Query) GetMetricName() string {
	return q.MetricName
}

func (q *Query) SetMetricaName(queryString string) error {
	if name := strings.Split(queryString, `/`)[2]; name != `` {
		q.MetricName = name
	} else {
		return errors.ErrEmptyMetricaName
	}
	return nil
}

/*
Get metrica raw value (string representation of value)

Returns:

	string: string representation of value
*/
func (q Query) GetMetricaRawValue() string {
	return q.MetricRawValue
}

func (q *Query) SetMetricaRawValue(queryString string) error {
	q.MetricRawValue = strings.Split(queryString, `/`)[3]
	return nil
}
