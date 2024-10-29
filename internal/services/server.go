package services

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейсы для описания взаимодействия с хранилищем метрик
type MetricStorager interface {
	MetricGetter
	MetricUpdater
	MetricPrinter
}

type MetricUpdater interface {
	UpdateGauge(metricName string, value float64) error
	AddCounter(metricName string, value int64) error
}

type MetricGetter interface {
	GetMetrica(metricaType string, metricaName string) (interface{}, error)
	GetAllMetrics() interface{}
}

type MetricPrinter interface {
	String() string
	HTML() string
}

// Интерфейс для описания взаимодействия с запросом на обновление метрики
// TO-DO: move from Qury to Metrica
type Querier interface {
	GetMetricaType() string
	SetMetricaType(string) error
	GetMetricName() string
	SetMetricaName(string) error
	GetMetricaRawValue() string
	SetMetricaRawValue(string) error
	String() string
}

type StringMetricaQuerier interface {
	SetMetricaRawValue(string) error
	String() string
}

type JSONMetricaQuerier interface {
	GetMetricaType() string
	GetMetricaName() string
	GetMetricaValue() *float64
	GetMetricaCounter() *int64
}

/*
Creating new instance of models.Query from request.URL string

Args:

	raw string: request.URL in string format. Example: "/update/gauge/test1/3.14"

Returns:

	q Querier: copy of instance, implemented Querier
	err error: nil or error occurred while parsing the raw string
*/
func ParseQueryString(raw string) (q Querier, err error) {
	queryString := raw[1:]
	if len(strings.Split(queryString, `/`)) != 4 {
		return nil, errors.ErrBadRawQuery
	}
	q = models.NewQuery()
	err = q.SetMetricaType(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
	err = q.SetMetricaName(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
	err = q.SetMetricaRawValue(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
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
func UpdateMetrica(q Querier, s MetricUpdater) error {
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

func JSONUpdateMetrica(jmq JSONMetricaQuerier, mu MetricUpdater) error {
	switch jmq.GetMetricaType() {
	case "gauge":
		err := mu.UpdateGauge(jmq.GetMetricaName(), *jmq.GetMetricaValue())
		if err != nil {
			return errors.ErrUpdateGauge
		}
	case "counter":
		err := mu.AddCounter(jmq.GetMetricaName(), *jmq.GetMetricaCounter())
		if err != nil {
			return errors.ErrAddCounter
		}
	default:
		return errors.ErrBadType
	}
	return nil
}

func SaveMetrics(mg MetricGetter, dst io.Writer) error {
	data, err := json.MarshalIndent(mg.GetAllMetrics(), "\t", "\t")
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	if err != nil {
		return err
	}
	return nil
}
