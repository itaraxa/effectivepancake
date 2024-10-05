package main

import (
	"os"
)

func parseEnv() {
	addressServer, ok := os.LookupEnv(`ADDRESS`)
	if ok {
		config.endpoint = addressServer
	}
}
