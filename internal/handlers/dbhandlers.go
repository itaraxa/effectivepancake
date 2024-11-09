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
func PingDB(l logger, s storagChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "wrong request type != GET", http.StatusMethodNotAllowed)
			return
		}
		l.Info("received a request to ping db-storage")
		w.Header().Set("Content-Type", "text/html")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := services.CheckConnectionStorage(ctx, l, s); err != nil {
			l.Error("error connection to storage", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			l.Info("succesful ping storage")
			w.WriteHeader(http.StatusOK)
		}
	}
}

// func Ping(l logger, dsn string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		if req.Method != http.MethodGet {
// 			http.Error(w, "wrong request type != GET", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		l.Info("received a request to ping db-storage")
// 		w.Header().Set("Content-Type", "text/html")
// 		db, err := sql.Open("pgx", dsn)
// 		if err != nil {
// 			l.Error("error connection to storage", "error", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		l.Info("connection openned", "dsn", dsn)

// 		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 		defer cancel()
// 		if err := db.PingContext(ctx); err != nil {
// 			l.Error("error ping storage database", "error", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		} else {
// 			l.Info("succesful ping storage")
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}
// 	}
// }
