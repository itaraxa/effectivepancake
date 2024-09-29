package handlers

import (
	"errors"
	"net/http"

	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/services"
)

func UpdateMemStorageHandler(s services.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Проверки запроса
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST request allowed", http.StatusMethodNotAllowed)
			return
		}
		if req.Header.Get("Content-Type") != "text/html" {
			http.Error(w, "Only text/html content allowed", http.StatusUnsupportedMediaType)
			return
		}

		// Выполнение логики
		q, err := services.ParseQueryString(req.URL.Path)
		if err != nil && errors.Is(err, myErrors.ErrBadRawQuery) {
			http.Error(w, "query string does not match the format", http.StatusNotFound)
			return
		}
		if err != nil && (errors.Is(err, myErrors.ErrBadType) || errors.Is(err, myErrors.ErrBadValue)) {
			http.Error(w, "invalid type or value", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "unknown parse query error", http.StatusInternalServerError)
		}

		services.ShowQuery(q)

		err = services.UpdateMetrica(q, s)
		if err != nil && (errors.Is(err, myErrors.ErrParseGauge) || errors.Is(err, myErrors.ErrParseCounter)) {
			http.Error(w, "the value is not of the specified type", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "unknown update metrica error", http.StatusInternalServerError)
		}

		services.ShowStorage(s)

		// Ответ
		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
