package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/services"
)

/*
Wrapper function for handler, what return all metrica values in HTML view

Args:

	s services.Storager: An object implementing the service.Storager interface

Returns:

	http.HandlerFunc
*/
func GetAllCurrentMetrics(s services.Storager, l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Uncorrect request type != GET", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("content-type", "text/html")
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
func GetMetrica(s services.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain")

		v, err := s.GetMetrica(chi.URLParam(req, "type"), chi.URLParam(req, "name"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			// TO-DO: add error reporting into logger
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
func UpdateMemStorageHandler(l *slog.Logger, s services.Storager) http.HandlerFunc {
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
			l.Error("query string does not match the format",
				slog.String("query string", req.URL.Path),
				slog.String("error", err.Error()),
			)
			return
		}
		if err != nil && (errors.Is(err, myErrors.ErrBadType) || errors.Is(err, myErrors.ErrBadValue)) {
			http.Error(w, "invalid type or value", http.StatusBadRequest)
			l.Error("invalid type or value",
				slog.String("query string", req.URL.Path),
				slog.String("error", err.Error()),
			)
			return
		}
		if err != nil {
			http.Error(w, "unknown parse query error", http.StatusInternalServerError)
			l.Error("unknown parse query error",
				slog.String("query string", req.URL.Path),
				slog.String("error", err.Error()),
			)
			return
		}

		err = services.UpdateMetrica(q, s)
		if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
			http.Error(w, "the value is not of the specified type", http.StatusBadRequest)
			l.Error("the value is not of the specified type",
				slog.String("query", q.String()),
				slog.String("error", err.Error()),
			)
			return
		}
		if err != nil {
			http.Error(w, "unknown update metrica error", http.StatusInternalServerError)
			l.Error("unknown update metrica error",
				slog.String("query", q.String()),
				slog.String("error", err.Error()),
			)
			return
		}

		// Response
		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
