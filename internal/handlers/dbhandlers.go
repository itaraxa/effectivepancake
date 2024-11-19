package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/itaraxa/effectivepancake/internal/services"
)

type storagChecker interface {
	PingContext(context.Context) error
}

/*
PingDB creates handler that check connection to storage with 3 sec. timeout

Args:

	l logger: a logger for printing messages
	s storageChecker: a storage, that implemeting storageChecker interface

Returns:

	http.HandlerFunc
*/
func PingDB(ctx context.Context, l logger, s storagChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "wrong request type != GET", http.StatusMethodNotAllowed)
			return
		}
		l.Info("received a request to ping db-storage")
		w.Header().Set("Content-Type", "text/html")
		ctxWithTimeout, cancelWithTimeout := context.WithTimeout(ctx, 3*time.Second)
		defer cancelWithTimeout()
		if err := services.CheckConnectionStorage(ctxWithTimeout, l, s); err != nil {
			l.Error("error connection to storage", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			l.Info("succesful ping storage")
			w.WriteHeader(http.StatusOK)
		}
	}
}
