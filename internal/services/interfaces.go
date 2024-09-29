package services

// Интерфейс для описания взаимодействия с хранилищем метрик
type Storager interface {
	UpdateGauge(metricName string, value float64) error
	AddCounter(metricName string, value int64) error
	String() string
}

// Интерфейс для описания взаимодействия с запросом на обновление метрики
type Querier interface {
	GetMetricaType() string
	GetMetricName() string
	GetMetricaRawValue() string
	String() string
}
