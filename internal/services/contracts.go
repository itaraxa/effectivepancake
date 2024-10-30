package services

import "github.com/itaraxa/effectivepancake/internal/models"

// Server sides interfaces
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

type JSONMetricaQuerier interface {
	GetMetricaType() string
	GetMetricaName() string
	GetMetricaValue() *float64
	GetMetricaCounter() *int64
}

// Agent-side interfaces
type MetricsAddGetter interface {
	MetricsAdderr
	MetricsGetter
}

type MetricsAdderr interface {
	AddData(data []models.JSONMetric) error
	AddPollCount(pollCount int64) error
}

type MetricsGetter interface {
	GetData() []models.JSONMetric
}

// Common interfaces
type logger interface {
	Error(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}
