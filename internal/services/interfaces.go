package services

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
