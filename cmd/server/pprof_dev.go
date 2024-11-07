//go:build dev
// +build dev

package main

import (
	"log"
	_ "net/http/pprof"

	"net/http"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
