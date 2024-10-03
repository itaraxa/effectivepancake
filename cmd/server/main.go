package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
)

func main() {
	ms := &memstorage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string][]int64),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get(`/`, handlers.GetAllCurrentMetrics(ms))
	r.Get(`/value/{type}/{name}`, handlers.GetMetrica(ms))
	r.Post(`/update/*`, handlers.UpdateMemStorageHandler(ms))

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
