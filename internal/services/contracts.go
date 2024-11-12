package services

import (
	"context"

	"github.com/itaraxa/effectivepancake/internal/models"
)

// Server sides interfaces
type MetricStorager interface {
	MetricGetter
	MetricUpdater
	MetricPrinter
	PingContext(context.Context) error
	Close() error
}

type MetricUpdater interface {
	UpdateGauge(context.Context, string, float64) error
	AddCounter(context.Context, string, int64) error
}

type MetricGetter interface {
	GetMetrica(context.Context, string, string) (interface{}, error)
	GetAllMetrics(context.Context) (interface{}, error)
}

type MetricPrinter interface {
	String(ctx context.Context) string
	HTML(ctx context.Context) string
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
