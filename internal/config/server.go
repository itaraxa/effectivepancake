package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/itaraxa/effectivepancake/internal/version"
)

type ServerConfig struct {
	Endpoint        string
	LogLevel        string
	ShowVersion     bool
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

/*
Returns a configuration structure with default parameters

Args:

	None

Returns:

	*ServerConfig: pointer to an instance of the ServerConfig struct
*/
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Endpoint:    `localhost:8080`,
		LogLevel:    `INFO`,
		ShowVersion: false,
	}
}

/*
A method for reading parameters passed as flags. The read parameters override those already specified in the ServerConf structure

Args:

	None

Returns:

	error: nil or error of parsing flags
*/
func (sc *ServerConfig) ParseFlags() error {
	flag.BoolVar(&sc.ShowVersion, `v`, false, `Show version and exit`)
	flag.StringVar(&sc.Endpoint, `a`, `localhost:8080`, `HTTP-server endpoint address. Environment variable ADDRESS`)
	flag.StringVar(&sc.LogLevel, `log`, `INFO`, `Set log level: INFO, DEBUG, etc.`)
	flag.BoolVar(&sc.Restore, `r`, true, `Restore saved data from the file. Environment variable RESTORE`)
	flag.StringVar(&sc.FileStoragePath, `f`, `metrics.dat`, `File path for saving metrics. Environment variable FILE_STORAGE_PATH`)
	flag.StringVar(&sc.DatabaseDSN, `d`, ``, `database connection string. Environment variable DATABASE_DSN`)
	flag.IntVar(&sc.StoreInterval, `i`, 300, `Time interval after which the current metrics are saved to a file. If set to 0, data is saved synchronously. Environment variable STORE_INTERVAL`)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s\n", version.ServerVersion, os.Args[0])
		flag.PrintDefaults()
	}
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		return err
	}

	return nil
}

/*
A method for reading parameters passed as environment variables. The read parameters override those already specified in the ServerConf structure

Args:

	None

Returns:

	error: nil or error of parsing environment variables
*/
func (sc *ServerConfig) ParseEnv() error {
	addressServer, ok := os.LookupEnv(`ADDRESS`)
	if ok {
		sc.Endpoint = addressServer
	}
	if storeInterval, ok := os.LookupEnv(`STORE_INTERVAL`); ok {
		i, err := strconv.Atoi(storeInterval)
		if err != nil {
			return fmt.Errorf(`uncorrect value in environment variable: %v`, err)
		}
		sc.StoreInterval = i
	}
	if fileStoragePath, ok := os.LookupEnv(`FILE_STORAGE_PATH`); ok {
		sc.FileStoragePath = fileStoragePath
	}
	if restore, ok := os.LookupEnv(`RESTORE`); ok {
		r, err := strconv.ParseBool(restore)
		if err != nil {
			return fmt.Errorf(`uncorrect value in environment variable: %v`, err)
		}
		sc.Restore = r
	}
	if database, ok := os.LookupEnv(`DATABASE_DSN`); ok {
		sc.DatabaseDSN = database
	}
	return nil
}
