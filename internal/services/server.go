package services

import (
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейс для описания взаимодействия с хранилищем метрик
type Storager interface {
	UpdateGauge(metricName string, value float64) error
	AddCounter(metricName string, value int64) error
	GetMetrica(metricaType string, metricaName string) (string, error)
	String() string
	HTML() string
}

// Интерфейс для описания взаимодействия с запросом на обновление метрики
type Querier interface {
	GetMetricaType() string
	SetMetricaType(string) (Querier, error)
	GetMetricName() string
	SetMetricaName(string) (Querier, error)
	GetMetricaRawValue() string
	SetMetricaRawValue(string) (Querier, error)
	String() string
}

/*
Creating new instance of models.Query from request.URL string

Args:

	raw string: request.URL in string format. Example: "/update/gauge/test1/3.14"

Returns:

	q models.Query: copy of instance models.Query
	err error: nil or error occurred while parsing the raw string
*/
func ParseQueryString(raw string) (q Querier, err error) {
	queryString := raw[1:]
	if len(strings.Split(queryString, `/`)) != 4 {
		return nil, errors.ErrBadRawQuery
	}
	q = models.NewQuery()
	q, err = q.SetMetricaType(queryString)
	if err != nil {
		return q, err
	}
	// q, _ = q.SetMetricaName(queryString)
	q, _ = q.SetMetricaRawValue(queryString)
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
		err = s.UpdateGauge(q.GetMetricName(), g)
		if err != nil {
			return errors.ErrUpdateGauge
		}
	case "counter":
		c, err := strconv.Atoi(q.GetMetricaRawValue())
		if err != nil {
			return errors.ErrParseCounter
		}
		err = s.AddCounter(q.GetMetricName(), int64(c))
		if err != nil {
			return errors.ErrAddCounter
		}
	default:
		return errors.ErrBadType
	}
	return nil
}
