package services

import (
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Функция для создания query
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

// Функция для обновления метрики из запроса в репозитории
func UpdateMetrica(q Querier, s Storager) error {
	switch q.GetMetricaType() {
	case "gauge":
		g, err := strconv.ParseFloat(q.GetMetricaRawValue(), 64)
		if err != nil {
			return errors.ErrParseGauge
		}
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
