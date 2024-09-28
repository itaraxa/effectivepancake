package main

import (
	"net/http"

	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
)

func main() {
	ms := &memstorage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string][]int64),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, handlers.UpdateMemStorageHandler(ms))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
