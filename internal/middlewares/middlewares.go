package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

/*
Helper structure for the middleware function
*/
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

/*
Overriding the writeHeader function to return the status code via a responseWriterWrapper structur

Args:

	statusCode int: status of processing request

Returns:

	None
*/
func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

/*
Middleware function for logging requests

Args:

	logger *slog.logger: pointer to logger

Returns:

	func(next http.Handler) http.Handler
*/
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

/*
Storage stat about requests
*/
type statResponse struct {
	statType map[string]int
	statCode map[int]int
	mu       sync.Mutex
	counter  int
}

/*
Singletone var for stat storage
*/
var (
	myStatResponse *statResponse
	once           sync.Once
)

/*
Creating stat storage. If a stat storage exist - return  the pointer to existing stat storage

Args:

	None

Returns:

	*statResponse: pointer to the statResponse struct
*/
func NewMyStatRes() *statResponse {
	// concarancy-safe realisation of singleton
	once.Do(func() {
		myStatResponse = &statResponse{
			statType: map[string]int{},
			statCode: map[int]int{},
			counter:  0,
		}
	})
	return myStatResponse
}

/*
Middleware function for gathering statistics about requests. Collect status and type of requests. Print to logger stats every logInterval requests

Args:

	logger *slog.Logger: the logger for outputting stat information
	logInterval int: the number of requests to output information after

Returns:

	func(next http.Handler) http.Handler
*/
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
