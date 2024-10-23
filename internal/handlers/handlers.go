package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/logger"
	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/services"
)

/*
Wrapper function for handler, what return all metrica values in HTML view

Args:

	s services.Storager: An object implementing the service.Storager interface

Returns:

	http.HandlerFunc
*/
func GetAllCurrentMetrics(s services.Storager, l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Uncorrect request type != GET", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(s.HTML()))
		if err != nil {
			l.Error("Cannot write HTML to response body")
		}
	}
}

/*
Wrapper function for handler, which return metrica value

Args:

	s services.Storager: An object implementing the service.Storager interface

Returns:

	http.HandlerFunc
*/
func GetMetrica(s services.Storager, l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		l.Info("get metrica", "type", chi.URLParam(req, "type"), "name", chi.URLParam(req, "name"))
		v, err := s.GetMetrica(chi.URLParam(req, "type"), chi.URLParam(req, "name"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			l.Error("cannot get metrica", "type", chi.URLParam(req, "type"), "name", chi.URLParam(req, "name"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(v))
	}
}

/*
Wrapper function for handler: writing the metric value to the storage

Args:

	s services.Storager: An object implementing the service.Storager interface

Returns:

	http.HandlerFunc
*/
func UpdateMemStorageHandler(l logger.Logger, s services.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Checks
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			l.Error("Only POST request allowed")
			return
		}

		// Processing
		q, err := services.ParseQueryString(req.URL.Path)
		if err != nil && (errors.Is(err, myErrors.ErrBadRawQuery) ||
			errors.Is(err, myErrors.ErrEmptyMetricaName) ||
			errors.Is(err, myErrors.ErrEmptyMetricaRawValue)) {
			http.Error(w, "query string does not match the format", http.StatusNotFound)
			l.Error("query string does not match the format", "query string", req.URL.Path, "error", err.Error())
			return
		}
		if err != nil && (errors.Is(err, myErrors.ErrBadType) || errors.Is(err, myErrors.ErrBadValue)) {
			http.Error(w, "invalid type or value", http.StatusBadRequest)
			l.Error("invalid type or value", "query string", req.URL.Path, "error", err.Error())
			return
		}
		if err != nil {
			http.Error(w, "unknown parse query error", http.StatusInternalServerError)
			l.Error("unknown parse query error", "query string", req.URL.Path, "error", err.Error())
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

func JSONUpdateMemStorageHandler(l logger.Logger, s services.Storager) http.HandlerFunc {
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
		var jq models.JSONQuery
		if err = json.Unmarshal(buf.Bytes(), &jq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			l.Error("cannot unmarshal data", "data", buf.Bytes(), "error", err.Error())
			return
		}

		err = services.JSONUpdateMetrica(jq, s)
		if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Error("the value is not of the specified type", "json query", jq.String(), "error", err.Error())
		}
		if err != nil {
			http.Error(w, "unknown update metrica error", http.StatusInternalServerError)
			l.Error("unknown update metrica error", "json query", jq.String(), "error", err.Error())
			return
		}

		// Response
		value, err := s.GetMetricaValue(jq.GetMetricaType(), jq.GetMetricaName())
		if err != nil {
			http.Error(w, "get metrica from storage error", http.StatusInternalServerError)
			l.Error("get metrica from storage error", "json query", jq.String(), "error", err.Error())
		}

		resp := jq
		switch jq.GetMetricaType() {
		case "gauge":
			g, _ := value.(float64)
			resp.Value = &g
		case "counter":
			resp := jq
			c, _ := value.(int64)
			resp.Delta = &c
		}

		body, _ := json.Marshal(resp)

		w.Header().Set("Content-Type", "applicttion/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}
