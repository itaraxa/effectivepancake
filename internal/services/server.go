package services

import (
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

/*
Creating new instance of models.Query from request.URL string

Args:

	raw string: request.URL in string format. Example: "/update/gauge/test1/3.14"

Returns:

	q models.Query: copy of instance models.Query
	err error: nil or error occurred while parsing the raw string
*/
func ParseQueryString(raw string) (q models.Query, err error) {
	queryString := raw[1:]
	if len(strings.Split(queryString, `/`)) != 4 {
		return q, errors.ErrBadRawQuery
	}
	q.MetricType = strings.Split(queryString, `/`)[1]
	if q.MetricType != "gauge" && q.MetricType != "counter" {
		return q, errors.ErrBadType
	}
	q.MetricName = strings.Split(queryString, `/`)[2]
	q.MetricRawValue = strings.Split(queryString, `/`)[3]

	return q, nil
}

/*
Writing data from the instance models.Query to storage.

Args:

	q Querier: object, implementing Querier interface
	s Storager: object, implementing Storager interface

Returns:

	error: nil or error, if occurred
*/
func UpdateMetrica(q Querier, s Storager) error {
	switch q.GetMetricaType() {
	case "gauge":
		g, err := strconv.ParseFloat(q.GetMetricaRawValue(), 64)
		if err != nil {
			return errors.ErrParseGauge
		}
		// In Storager save value as int64/float64 type
		s.UpdateGauge(q.GetMetricName(), g)
	case "counter":
		c, err := strconv.Atoi(q.GetMetricaRawValue())
		if err != nil {
			return errors.ErrParseCounter
		}
		s.AddCounter(q.GetMetricName(), int64(c))
	default:
		return errors.ErrBadType
	}
	return nil
}
