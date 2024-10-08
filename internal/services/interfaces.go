package services

import "github.com/itaraxa/effectivepancake/internal/models"

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
	GetMetricName() string
	GetMetricaRawValue() string
	String() string
}

// Интерфейс для работы с метриками на агенте
type Metricer interface {
	AddData(data []models.Metric) error
	AddPollCount(pollCount uint64) error

	String() string
}
