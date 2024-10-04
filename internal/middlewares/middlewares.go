package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Logging requests via slog logger
func LoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Request
			start := time.Now()
			wrappedWriter := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			logger.Debug("Request received",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			// Doing next middleware
			next.ServeHTTP(wrappedWriter, r)

			// Response
			logger.Debug("Request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrappedWriter.statusCode),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

type statResponse struct {
	statType map[string]int
	statCode map[int]int
	mu       sync.Mutex
	counter  int
}

var (
	myStatResponse *statResponse
	once           sync.Once
)

// Function for creating singleton of statResponse
func NewMyStatRes() *statResponse {
	once.Do(func() {
		myStatResponse = &statResponse{
			statType: map[string]int{},
			statCode: map[int]int{},
			counter:  0,
		}
	})
	return myStatResponse
}

// Show stats into logger
func StatMiddleware(logger *slog.Logger, logInterval int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			stat := NewMyStatRes()
			next.ServeHTTP(wrappedWriter, r)
			stat.mu.Lock()
			defer stat.mu.Unlock()
			stat.counter += 1
			stat.statType[r.Method] += 1
			stat.statCode[wrappedWriter.statusCode] += 1
			if stat.counter%logInterval == 0 {
				logger.Info("Request stat:",
					slog.Int("counter", stat.counter),
					slog.String("Type stat", fmt.Sprintf("%v", stat.statType)),
					slog.String("StatusCode stat", fmt.Sprintf("%v", stat.statCode)))
			}
		})
	}
}
