package main

import (
	"context"
	"log"
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
	"github.com/itaraxa/effectivepancake/internal/repositories/postgres"
	"github.com/itaraxa/effectivepancake/internal/services"
	"github.com/itaraxa/effectivepancake/internal/version"
)

// Structure for embedding dependencies into the server app
type ServerApp struct {
	logger  logger.Logger
	storage services.MetricStorager
	router  *chi.Mux
	config  *config.ServerConfig
}

/*
NewServerApp creates an empty instance of the serverApp structure

Args:

	logger logger.Logger: object, implementing the logger.Logger interface
	storage services.MetricStorager: object, implementing the services.MetricStorager interface
	router *chi.Mux: http router
	config  *config.ServerConfig: pointer to config.ServerConfig instance

Returns:

	*ServerApp: pointer to the ServerApp instance
*/
func NewServerApp(logger logger.Logger, storage services.MetricStorager, router *chi.Mux, config *config.ServerConfig) *ServerApp {
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
		"Database DSN", sa.config.DatabaseDSN,
	)
	defer sa.logger.Info("server stopped")

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	stopServerChan := make(chan bool, 1)
	signal.Notify(signalChan, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(cancel context.CancelFunc) {
		defer cancel()
		defer func() {
			_ = sa.storage.Close()
		}()
		<-signalChan
		sa.logger.Info("stopping server", "cause", "Exit programm because Ctrl+C press")
		// Saving metric to a file before Exit
		err := services.SaveMetricsToFile(ctx, sa.logger, sa.storage, sa.config.FileStoragePath)
		if err != nil {
			sa.logger.Error("metric data hasn't been saved to the file", "error", err.Error(), "filename", sa.config.FileStoragePath)
			return
		}
		sa.logger.Info("metric data has been saved to the file", "filename", sa.config.FileStoragePath)
		stopServerChan <- true
		// _ = sa.storage.Close()
		// cancel()
	}(cancel)

	// Restoring metric data from the file
	// если воостанавливаем метрики из файла, то предварительно очищаем хранилище
	if sa.config.Restore {
		sa.logger.Info("clear storage")
		ctxWithTimeout, cancelWithTimeout := context.WithTimeout(ctx, 3*time.Second)
		defer cancelWithTimeout()
		err := sa.storage.Clear(ctxWithTimeout)
		if err != nil {
			sa.logger.Error("cleaning storage before metrics loading from file", "error", err.Error())
		}
		sa.logger.Info("try to load metrics from file", "filename", sa.config.FileStoragePath)
		err = services.LoadMetricsFromFile(sa.logger, sa.storage, sa.config.FileStoragePath)
		if err != nil {
			sa.logger.Error("metrics wasn't loaded from file", "error", err.Error(), "filename", sa.config.FileStoragePath)
		} else {
			sa.logger.Info("metrics have been loaded from file")
		}
	}

	// Writing metric data to the file periodically if don't use postgres storage
	if sa.config.StoreInterval > 0 && sa.config.DatabaseDSN == "" {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(sa.config.StoreInterval))
			defer ticker.Stop()
			for {
				<-ticker.C
				if err := services.SaveMetricsToFile(ctx, sa.logger, sa.storage, sa.config.FileStoragePath); err != nil {
					sa.logger.Error("cannot save data to file", "error", err.Error())
				}
			}
		}()
	}

	// Add middlewares
	sa.router.Use(middlewares.LoggerMiddleware(sa.logger))
	sa.router.Use(middlewares.CompressResponceMiddleware(sa.logger))
	sa.router.Use(middlewares.DecompressRequestMiddleware(sa.logger))
	sa.router.Use(middlewares.StatMiddleware(sa.logger, 10))
	if sa.config.StoreInterval == 0 && sa.config.DatabaseDSN == "" {
		sa.logger.Info("synchronous file writing is used")
		file, err := os.OpenFile(sa.config.FileStoragePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			sa.logger.Error("cannot open file for writing", "error", err.Error())
			return
		}
		defer file.Close()
		sa.router.Use(middlewares.SaveStorageToFile(ctx, sa.logger, sa.storage, file))
	}

	// Add routes
	// health-checks
	sa.router.Get(`/ping`, handlers.PingDB(ctx, sa.logger, sa.storage))
	sa.router.Get(`/ping/`, handlers.PingDB(ctx, sa.logger, sa.storage))
	// query-row routs
	sa.router.Get(`/value/{type}/{name}`, handlers.GetMetrica(ctx, sa.storage, sa.logger))
	sa.router.Post(`/update/*`, handlers.PostUpdateHandler(ctx, sa.logger, sa.storage))
	// json routs
	sa.router.Post(`/value`, handlers.JSONGetMetrica(ctx, sa.storage, sa.logger))
	sa.router.Post(`/value/`, handlers.JSONGetMetrica(ctx, sa.storage, sa.logger))
	sa.router.Post(`/update/`, handlers.PostJSONUpdateHandler(ctx, sa.logger, sa.storage))
	sa.router.Post(`/updates/`, handlers.PostJSONUpdateBatchHandler(ctx, sa.logger, sa.storage))
	// get all metrics
	sa.router.Get(`/`, handlers.GetAllCurrentMetrics(ctx, sa.storage, sa.logger))

	// Start router
	server := &http.Server{
		Addr:    sa.config.Endpoint,
		Handler: sa.router,
	}

	go func() {
		sa.logger.Info("start router")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			sa.logger.Fatal("router error", "err", err.Error())
		}
	}()

	// stopping http server
	<-stopServerChan
	ctxWithTimeout, cancelWithTimeout := context.WithTimeout(ctx, 3*time.Second)
	defer cancelWithTimeout()
	err := server.Shutdown(ctxWithTimeout)
	if err != nil {
		sa.logger.Fatal("stopping server", "error", err.Error())
	}
	sa.logger.Info("server stopped gracefully")
}

func main() {
	serverConf := config.NewServerConfig()
	err := serverConf.ParseFlags()
	if err != nil {
		log.Fatalf("error parsing flags: %v", err)
	}
	if serverConf.ShowVersion {
		log.Print(version.ServerVersion)
		return
	}

	err = serverConf.ParseEnv()
	if err != nil {
		log.Fatalf("error parsing environment variable: %v", err)
	}

	logger, err := logger.NewZapLogger(serverConf.LogLevel)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}
	defer logger.Sync()

	r := chi.NewRouter()

	var s services.MetricStorager
	if serverConf.DatabaseDSN != "" {
		s, err = postgres.NewPostgresRepository(context.Background(), serverConf.DatabaseDSN)
		if err != nil {
			log.Fatalf("error connecting to database: %v", err)
		}
		defer s.Close()
	} else {
		s = memstorage.NewMemStorage()
	}
	app := NewServerApp(logger, s, r, serverConf)
	app.Run()

}
