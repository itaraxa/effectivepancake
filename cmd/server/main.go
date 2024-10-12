package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/middlewares"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
	"github.com/itaraxa/effectivepancake/internal/version"
)

type ServerApp struct {
	logger  *slog.Logger
	storage *memstorage.MemStorage
	router  *chi.Mux
	config  *config.ServerConfig
}

func NewServerApp(logger *slog.Logger, storage *memstorage.MemStorage, router *chi.Mux, config *config.ServerConfig) *ServerApp {
	return &ServerApp{
		logger:  logger,
		storage: storage,
		router:  router,
		config:  config,
	}
}

func (sa *ServerApp) Run() {
	sa.logger.Info("Server started", slog.String(`Listen`, sa.config.Endpoint))
	defer sa.logger.Info("Server stoped")

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		sa.logger.Info("Stopping server", slog.String("cause", "Exit programm because Ctrl+C press"))
		os.Exit(0)
	}()

	// Add middleware
	sa.router.Use(middlewares.LoggerMiddleware(sa.logger))
	sa.router.Use(middlewares.StatMiddleware(sa.logger, 10))

	// Add routes
	sa.router.Get(`/`, handlers.GetAllCurrentMetrics(sa.storage, sa.logger))
	sa.router.Get(`/value/{type}/{name}`, handlers.GetMetrica(sa.storage))
	sa.router.Post(`/update/*`, handlers.UpdateMemStorageHandler(sa.storage))

	// Start router
	sa.logger.Info("Start router")
	err := http.ListenAndServe(sa.config.Endpoint, sa.router)
	if err != nil {
		sa.logger.Error(fmt.Sprintf("router error: %v", err))
		os.Exit(1)
	}
}

func main() {
	serverConf := config.NewServerConfig()
	err := serverConf.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		os.Exit(1)
	}
	if serverConf.ShowVersion {
		fmt.Println(version.ServerVersion)
		os.Exit(0)
	}

	err = serverConf.ParseEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing environment variable: %v", err)
		os.Exit(1)
	}

	var level slog.Level
	switch serverConf.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	ms := memstorage.NewMemStorage()

	r := chi.NewRouter()

	app := NewServerApp(logger, ms, r, serverConf)
	app.Run()
}
