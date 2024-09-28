package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/itaraxa/effectivepancake/internal/models"
)

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
