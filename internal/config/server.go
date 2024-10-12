package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/itaraxa/effectivepancake/internal/version"
)

type ServerConfig struct {
	Endpoint    string
	LogLevel    string
	ShowVersion bool
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
	flag.StringVar(&sc.Endpoint, `a`, `localhost:8080`, `HTTP-server endpoint address`)
	flag.StringVar(&sc.LogLevel, `log`, `INFO`, `Set log level: INFO, DEBUG, etc.`)

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
	return nil
}
