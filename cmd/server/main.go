package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/handlers"
	"github.com/itaraxa/effectivepancake/internal/logger"
	"github.com/itaraxa/effectivepancake/internal/middlewares"
	"github.com/itaraxa/effectivepancake/internal/repositories/memstorage"
	"github.com/itaraxa/effectivepancake/internal/services"
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
	sa.logger.Info("server version",
		"Version", version.ServerVersion,
	)
	sa.logger.Info("server started",
		"Listen", sa.config.Endpoint,
		"Log level", sa.config.LogLevel,
		"Restore", sa.config.Restore,
		"Storing metrica file", sa.config.FileStoragePath,
		"Store interval", time.Duration(sa.config.StoreInterval)*time.Second,
	)
	defer sa.logger.Info("server stopped")

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		sa.logger.Info("stopping server", "cause", "Exit programm because Ctrl+C press")
		// Saving metric to a file before Exit
		if err := services.SaveMetricsToFile(sa.logger, sa.storage, sa.config.FileStoragePath); err == nil {
			sa.logger.Info("metric data has been saved to the file", "filename", sa.config.FileStoragePath)
		} else {
			sa.logger.Error("metric data hasn't been saved to the file", "error", err.Error(), "filename", sa.config.FileStoragePath)
		}
		os.Exit(0)
	}()

	// Restoring metric data from the file
	if sa.config.Restore {
		sa.logger.Info("try to load metrics from file", "filename", sa.config.FileStoragePath)
		err := services.LoadMetricsFromFile(sa.logger, sa.storage, sa.config.FileStoragePath)
		if err != nil {
			sa.logger.Error("metrics wasn't loaded from file", "error", err.Error(), "filename", sa.config.FileStoragePath)
		} else {
			sa.logger.Info("metrics have been loaded from file")
		}
	}

	// Writing metric data to the file periodically
	if sa.config.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(sa.config.StoreInterval))
			for {
				<-ticker.C
				if err := services.SaveMetricsToFile(sa.logger, sa.storage, sa.config.FileStoragePath); err != nil {
					sa.logger.Error("cannot save data to file", "error", err.Error())
				}
			}
		}()
	}

	// Add middlewares
	sa.router.Use(middlewares.LoggerMiddleware(sa.logger))
	// sa.router.Use(middleware.Compress(5, "application/json"))
	sa.router.Use(middlewares.CompressResponceMiddleware(sa.logger))
	sa.router.Use(middlewares.DecompressRequestMiddleware(sa.logger))
	sa.router.Use(middlewares.StatMiddleware(sa.logger, 10))
	if sa.config.StoreInterval == 0 {
		sa.logger.Info("synchronous file writing is used")
		file, err := os.OpenFile(sa.config.FileStoragePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			sa.logger.Error("cannot open file for writing", "error", err.Error())
			return
		}
		defer file.Close()
		sa.router.Use(middlewares.SaveStorageToFile(sa.logger, sa.storage, file))
	}

	// Add routes
	sa.router.Get(`/`, handlers.GetAllCurrentMetrics(sa.storage, sa.logger))
	sa.router.Get(`/value/{type}/{name}`, handlers.GetMetrica(sa.storage, sa.logger))
	sa.router.Post(`/value/`, handlers.JSONGetMetrica(sa.storage, sa.logger))
	sa.router.Post(`/update/`, handlers.JSONUpdateHandler(sa.logger, sa.storage))
	sa.router.Post(`/update/*`, handlers.UpdateHandler(sa.logger, sa.storage))

	// Start router
	sa.logger.Info("start router")
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
		fmt.Fprintf(os.Stderr, "error parsing flags: %v", err)
		os.Exit(1)
	}
	if serverConf.ShowVersion {
		fmt.Println(version.ServerVersion)
		os.Exit(0)
	}

	err = serverConf.ParseEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing environment variable: %v", err)
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
