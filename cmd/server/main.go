package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/logger"
	"github.com/itaraxa/effectivepancake/internal/middlewares"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
	"github.com/itaraxa/effectivepancake/internal/version"
)

// Structure for embedding dependencies into the server app
type ServerApp struct {
	logger  logger.Logger
	storage *memstorage.MemStorage
	router  *chi.Mux
	config  *config.ServerConfig
}

/*
NewServerApp creates an empty instance of the serverApp structure

Args:

	logger logger.Logger: object, implementing the logger.Logger interface
	storage *memstorage.MemStorage: pointer to memstorage.MemStorage object
	router *chi.Mux: http router
	config  *config.ServerConfig: pointer to config.ServerConfig instance

Returns:

	*ServerApp: pointer to the ServerApp instance
*/
func NewServerApp(logger logger.Logger, storage *memstorage.MemStorage, router *chi.Mux, config *config.ServerConfig) *ServerApp {
	return &ServerApp{
		logger:  logger,
		storage: storage,
		router:  router,
		config:  config,
	}
}

/*
Run function start a logging and http-routing processes

Args:

	sa *ServerApp: pointer to ServerApp structure with injected dependencies
*/
func (sa *ServerApp) Run() {
	sa.logger.Info("Server version",
		"Version", version.ServerVersion,
	)
	sa.logger.Info("Server started",
		"Listen", sa.config.Endpoint,
		"Log level", sa.config.LogLevel,
	)
	defer sa.logger.Info("Server stoped")

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		sa.logger.Info("Stopping server", zap.String("cause", "Exit programm because Ctrl+C press"))
		os.Exit(0)
	}()

	// Add middleware
	sa.router.Use(middlewares.LoggerMiddleware(sa.logger))
	// sa.router.Use(middlewares.StatMiddleware(sa.logger, 10))

	// Add routes
	sa.router.Get(`/`, handlers.GetAllCurrentMetrics(sa.storage, sa.logger))
	sa.router.Get(`/value/{type}/{name}`, handlers.GetMetrica(sa.storage, sa.logger))
	sa.router.Post(`/value/`, handlers.JSONGetMetrica(sa.storage, sa.logger))
	sa.router.Post(`/update/`, handlers.JSONUpdateHandler(sa.logger, sa.storage))
	sa.router.Post(`/update/*`, handlers.UpdateHandler(sa.logger, sa.storage))

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

	logger, err := logger.NewZapLogger(serverConf.LogLevel)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ms := memstorage.NewMemStorage()

	r := chi.NewRouter()

	app := NewServerApp(logger, ms, r, serverConf)
	app.Run()
}
