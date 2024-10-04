package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/middlewares"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
)

type ServerApp struct {
	logger  *slog.Logger
	storage *memstorage.MemStorage
	router  *chi.Mux
}

// Можно ли тут использовать интерфейсы? Как?
func NewServerApp(logger *slog.Logger, storage *memstorage.MemStorage, router *chi.Mux) *ServerApp {
	return &ServerApp{
		logger:  logger,
		storage: storage,
		router:  router,
	}
}

func (sa *ServerApp) Run() {
	sa.logger.Info("Server started")
	defer sa.logger.Info("Server stoped")

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		// Wait for closing requests
		time.Sleep(1 * time.Second)
		sa.logger.Info("Exit programm because Ctrl+C press")
		os.Exit(0)
	}()

	// Add middleware
	sa.router.Use(middlewares.LoggerMiddleware(sa.logger))
	sa.router.Use(middlewares.StatMiddleware(sa.logger, 10))

	// Add routes
	sa.router.Get(`/`, handlers.GetAllCurrentMetrics(sa.storage))
	sa.router.Get(`/value/{type}/{name}`, handlers.GetMetrica(sa.storage))
	sa.router.Post(`/update/*`, handlers.UpdateMemStorageHandler(sa.storage))

	// Start router
	sa.logger.Info("Start router")
	err := http.ListenAndServe(`:8080`, sa.router)
	if err != nil {
		sa.logger.Error(fmt.Sprintf("router error: %v", err))
		os.Exit(1)
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ms := memstorage.NewMemStorage()

	r := chi.NewRouter()

	app := NewServerApp(logger, ms, r)
	app.Run()

}
