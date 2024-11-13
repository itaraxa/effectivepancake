package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/services"
)

const (
	gauge                = `gauge`
	counter              = `counter`
	maxQueryStringLength = 256
)

type logger interface {
	Error(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

type metricStorager interface {
	metricGetter
	metricUpdater
	metricBatchUpdater
}

type metricGetter interface {
	GetMetrica(context.Context, string, string) (interface{}, error)
}

type metricUpdater interface {
	UpdateGauge(ctx context.Context, metricName string, value float64) error
	AddCounter(ctx context.Context, metricName string, value int64) error
}

type metricBatchUpdater interface {
	UpdateBatchGauge(context.Context, map[string]*float64) error
	AddBatchCounter(context.Context, map[string]*int64) error
}

type metricPrinter interface {
	HTML(ctx context.Context) string
}

/*
GetAllCurrentMetrics creates handler that return all metrica values in HTML view

Args:

	s metricPrinter: An object implementing the service.Storager interface
	l logger: a logger for printing messages

Returns:

	http.HandlerFunc
*/
func GetAllCurrentMetrics(s metricPrinter, l logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "uncorrect request type != GET", http.StatusMethodNotAllowed)
			return
		}
		l.Info("received a request to retrieve the current value of all metrics")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(s.HTML(context.TODO())))
		if err != nil {
			l.Error("cannot write HTML to response body")
		}
	}
}

/*
GetMetrica creates a handler that returns the metric value

Args:

	s metricGetter: An object implementing the service.Storager interface
	l logger: a logger for printing messages

Returns:

	http.HandlerFunc
*/
func GetMetrica(s metricGetter, l logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		l.Info("received a request to get metrica", "type", chi.URLParam(req, "type"), "name", chi.URLParam(req, "name"))
		v, err := s.GetMetrica(context.TODO(), chi.URLParam(req, "type"), chi.URLParam(req, "name"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			l.Error("cannot get metrica", "type", chi.URLParam(req, "type"), "name", chi.URLParam(req, "name"), "error", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)

		res := ""
		switch chi.URLParam(req, "type") {
		case gauge:
			res = fmt.Sprint(v.(float64))
		case counter:
			res = fmt.Sprint(v.(int64))
		}
		_, _ = w.Write([]byte(res))
	}
}

/*
JSONGetMetrica creates a handler that return metrica value in JSON.
Timeout for getting metrica from storage is 5 seconds.

Args:

	s metricGetter: a storage that allows getting metric
	l logger: a logger for printing messages

Returns:

	http.HandlerFunc
*/
func JSONGetMetrica(s metricGetter, l logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Checks
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			l.Error("Only POST request allowed")
			return
		}

		// Processing
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot read from request body")
			return
		}
		var jm models.JSONMetric
		if err = json.Unmarshal(buf.Bytes(), &jm); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot unmarshal data", "data", buf.Bytes(), "error", err.Error())
			return
		}
		l.Info("received a request to get metrica in JSON", "type", jm.GetMetricaType(), "name", jm.GetMetricaName())
		valueFromStorage, err := s.GetMetrica(ctx, jm.GetMetricaType(), jm.GetMetricaName())
		if err != nil && errors.Is(err, myErrors.ErrMetricaNotFaund) {
			w.WriteHeader(http.StatusNotFound)
			l.Error("cannot get metrica", "type", jm.GetMetricaType(), "name", jm.GetMetricaName(), "error", err.Error())
			return
		}
		if err != nil {
			http.Error(w, "unknown getting metrica error", http.StatusNotFound)
			l.Error("cannot get metrica", "type", jm.GetMetricaType(), "name", jm.GetMetricaName(), "error", err.Error())
			return
		}

		// Type switching
		switch jm.GetMetricaType() {
		case gauge:
			t := valueFromStorage.(float64)
			jm.Value = &t
		case counter:
			t := valueFromStorage.(int64)
			jm.Delta = &t
		}

		// Write response
		jsonData, err := json.Marshal(jm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.Error("cannot marshal data", "error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.Error("cannot write data to body", "error", err.Error())
			return
		}
	}
}

/*
UpdateHandler creates handler that writes the metric value to the storage

Args:

	l logger: a logger for printing messages
	s metricUpdater: a storage that allows update metric data

Returns:

	http.HandlerFunc
*/
func UpdateHandler(l logger, s metricUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		queryString := req.URL.Path
		// Checks
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			l.Error("Only POST request allowed")
			return
		}
		if len(queryString) > maxQueryStringLength {
			http.Error(w, "Query sting too long", http.StatusBadRequest)
			l.Error("query string too long", "query string", queryString[:maxQueryStringLength], "query string length", len(queryString))
			return
		}

		// Processing
		q, err := services.ParseQueryString(queryString)
		if err != nil && (errors.Is(err, myErrors.ErrBadRawQuery) ||
			errors.Is(err, myErrors.ErrEmptyMetricaName) ||
			errors.Is(err, myErrors.ErrEmptyMetricaRawValue)) {
			http.Error(w, "query string does not match the format", http.StatusNotFound)
			l.Error("query string does not match the format", "query string", queryString, "error", err.Error())
			return
		}
		if err != nil && (errors.Is(err, myErrors.ErrBadType) || errors.Is(err, myErrors.ErrBadValue)) {
			http.Error(w, "invalid type or value", http.StatusBadRequest)
			l.Error("invalid type or value", "query string", queryString, "error", err.Error())
			return
		}
		if err != nil {
			http.Error(w, "unknown parse query error", http.StatusInternalServerError)
			l.Error("unknown parse query error", "query string", queryString, "error", err.Error())
			return
		}

		err = services.UpdateMetrica(q, s)
		if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
			http.Error(w, "the value is not of the specified type", http.StatusBadRequest)
			l.Error("the value is not of the specified type", "query", q.String(), "error", err.Error())
			return
		}
		if err != nil {
			http.Error(w, "unknown update metrica error", http.StatusInternalServerError)
			l.Error("unknown update metrica error", "query", q.String(), "error", err.Error())
			return
		}

		// Response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}

/*
JSONUpdateHandler creates handler that updates metric values received in JSON format to the storage

Args:

	l logger: a logger for printing messages
	s metricUpdater: a storage that allows update metric data

Returns:

	http.HandlerFunc
*/
func JSONUpdateHandler(l logger, s metricStorager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Checks
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			l.Error("Only POST request allowed")
			return
		}

		// Processing
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot read from request body")
			return
		}
		var jm models.JSONMetric
		if err = json.Unmarshal(buf.Bytes(), &jm); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot unmarshal data", "data", buf.String(), "error", err.Error())
			return
		}
		// validating request
		// Check nil values
		if jm.ID == "" {
			http.Error(w, "metric name not found", http.StatusNotFound)
			l.Error("cannot update metrica", "data", buf.String(), "error", "metric name not found")
			return
		}
		if jm.Delta == nil && jm.Value == nil {
			http.Error(w, "any metric value is not set", http.StatusBadRequest)
			l.Error("cannot update metrica", "data", buf.String(), "error", "any metric value is not set")
			return
		}
		if jm.Delta == nil && jm.MType == counter {
			http.Error(w, "the counter delta is not set", http.StatusBadRequest)
			l.Error("cannot update metrica", "data", buf.String(), "error", "the counter delta is not set")
			return
		}
		if jm.Value == nil && jm.MType == gauge {
			http.Error(w, "the gauge value is not set", http.StatusBadRequest)
			l.Error("cannot update metrica", "data", buf.String(), "error", "the gauge value is not set")
			return
		}

		// updating metrica in storage
		err = services.JSONUpdateMetrica(jm, s)
		if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("the value is not of the specified type", "json query", jm.String(), "error", err.Error())
			return
		}
		if err != nil && errors.Is(err, myErrors.ErrBadType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("unknown metrica type update error", "json query", jm.String(), "error", err.Error())
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Error("metrica update error", "json query", jm.String(), "error", err.Error())
			return
		}

		// Response
		value, err := s.GetMetrica(context.TODO(), jm.GetMetricaType(), jm.GetMetricaName())
		if err != nil {
			http.Error(w, "get metrica from storage error", http.StatusInternalServerError)
			l.Error("get metrica from storage error", "json query", jm.String(), "error", err.Error())
		}

		resp := jm
		switch jm.GetMetricaType() {
		case gauge:
			g, _ := value.(float64)
			resp.Value = &g
		case counter:
			c, _ := value.(int64)
			resp.Delta = &c
		}

		body, _ := json.Marshal(resp)

		w.Header().Set("Content-Type", "applicttion/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

func JSONUpdateBatchHandler(l logger, s metricStorager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Checks
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			l.Error("Only POST request allowed")
			return
		}
		// Processing
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot read from request body")
			return
		}
		var jms []models.JSONMetric
		if err = json.Unmarshal(buf.Bytes(), &jms); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot unmarshal data", "data", buf.String(), "error", err.Error())
			return
		}

		// convert slice of models.JSONMetric into slice of services.JSONMetricaQuerier
		var jmqs []services.JSONMetricaQuerier
		for _, jmq := range jms {
			jmqs = append(jmqs, jmq)
		}

		// updating metrica in storage
		err = services.JSONUpdateBatchMetrica(l, jmqs, s)
		l.Info("request batch update", "body", fmt.Sprint(jmqs))
		// if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	l.Error("the value is not of the specified type", "json query", jms.String(), "error", err.Error())
		// 	return
		// }
		// if err != nil && errors.Is(err, myErrors.ErrBadType) {
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	l.Error("unknown metrica type update error", "json query", jm.String(), "error", err.Error())
		// 	return
		// }
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Error("metrica update error", "json query", buf.String(), "error", err.Error())
			return
		}

		w.Header().Set("Content-Type", "applicttion/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
	}
}
