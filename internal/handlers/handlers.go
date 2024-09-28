package handlers

import (
	"net/http"

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
		if err != nil {
			http.Error(w, "Bad request: error in query", http.StatusBadRequest)
			return
		}
		services.ShowQuery(q)

		err = services.UpdateMetrica(q, s)
		if err != nil {
			http.Error(w, "Server error: cannot update metrica", http.StatusBadRequest)
			return
		}
		services.ShowStorage(s)

		// Ответ
		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
