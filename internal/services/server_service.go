package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейс для описания взаимодействия с хранилищем метрик
type Storager interface {
	UpdateGauge(metric_name string, value float64) error
	AddCounter(metric_name string, value int64) error
	String() string
}

type Querier interface {
	GetMetricaType() string
	GetMetricName() string
	GetMetricaRawValue() string
	String() string
}

// Функция для создания query
func ParseQueryString(raw string) (q models.Query, err error) {
	q.MetricType = strings.Split(raw, `/`)[2]
	q.MetricName = strings.Split(raw, `/`)[3]
	q.MetricRawValue = strings.Split(raw, `/`)[4]

	return q, nil
}

func UpdateMetrica(q Querier, s Storager) error {
	switch q.GetMetricaType() {
	case "gauge":
		g, err := strconv.ParseFloat(q.GetMetricaRawValue(), 64)
		if err != nil {
			return fmt.Errorf("parse gauge value error: %w", err)
		}
		s.UpdateGauge(q.GetMetricName(), g)
	case "counter":
		c, err := strconv.Atoi(q.GetMetricaRawValue())
		if err != nil {
			return fmt.Errorf("parse counter value error: %w", err)
		}
		s.AddCounter(q.GetMetricName(), int64(c))
		// ms.Counter[q.MetricName] = append(ms.Counter[q.MetricName], int64(c))
	default:
		return fmt.Errorf("uncorrect type of metrica")
	}
	return nil
}
