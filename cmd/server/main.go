package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
)

// Интерфейс для описания взаимодействия с хранилищем метрик
type Storager interface {
	UpdateGauge(metric_name string, value float64) error
	AddCounter(metric_name string, value int64) error
	GetGauge(metric_name string) (float64, error)
	GetCounter(metric_name string) ([]int64, error)
}

func UpdateMemStorageHandler(ms *memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// ms.mu.Lock()
		// defer ms.mu.Unlock()

		// Проверки запроса
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			return
		}
		if req.Header.Get("Content-Type") != "text/html" {
			http.Error(w, "Only text/html content allowed", http.StatusUnsupportedMediaType)
			return
		}

		// Строка запроса в формате /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		q, err := models.Parse(req.URL.Path)
		if err != nil {
			http.Error(w, "Bad request: error in query", http.StatusBadRequest)
			return
		}
		fmt.Println(q)

		switch q.MetricType {
		case "gauge":
			g, err := strconv.ParseFloat(q.MetricRawValue, 64)
			if err != nil {
				http.Error(w, "Uncorrect value", http.StatusBadRequest)
				return
			}
			ms.Gauge[q.MetricName] = g
		case "counter":
			c, err := strconv.Atoi(q.MetricRawValue)
			if err != nil {
				http.Error(w, "Uncorrect value", http.StatusBadRequest)
				return
			}
			ms.Counter[q.MetricName] = append(ms.Counter[q.MetricName], int64(c))
		default:
			http.Error(w, "Uncorrect type of metric", http.StatusBadRequest)
			return
		}
		fmt.Println(ms)

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	ms := &memstorage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string][]int64),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, UpdateMemStorageHandler(ms))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
