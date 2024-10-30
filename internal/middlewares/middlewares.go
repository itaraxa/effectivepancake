package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/services"
)

type logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

type storager interface {
	GetMetrica(metricaType string, metricaName string) (interface{}, error)
	GetAllMetrics() interface{}
}

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

	l logger: implementation of logger interface

Returns:

	func(next http.Handler) http.Handler
*/
func LoggerMiddleware(l logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Request
			start := time.Now()
			wrappedWriter := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			l.Info("Request received", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)

			// Doing next middleware
			next.ServeHTTP(wrappedWriter, r)

			// Response
			l.Info("Request completed", "method", r.Method, "path", r.URL.Path, "status", wrappedWriter.statusCode, "duration", time.Since(start))
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

	l logger: implementation of logger interface
	logInterval int: the number of requests to output information after

Returns:

	func(next http.Handler) http.Handler
*/
func StatMiddleware(l logger, logInterval int) func(next http.Handler) http.Handler {
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
				l.Info("Request stat:", "counter", stat.counter, "Type stat", fmt.Sprintf("%v", stat.statType), "StatusCode stat", fmt.Sprintf("%v", stat.statCode))
			}
		})
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func CompressResponceMiddleware(l logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				l.Info("Responce will not compressed")
				next.ServeHTTP(w, r)
				return
			}

			// if r.Header.Get("Content-Type") != "text/html" && r.Header.Get("Content-Type") != "application/json" {
			// 	l.Debug("Responce will not compressed")
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

			// compressing responce
			gzw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
				l.Error("cannot compress responce", "error", err.Error())
				return
			}
			defer gzw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			// w.WriteHeader(http.StatusOK)
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzw}, r)
		})
	}
}

/*
DecompressRequestMiddleware decompress gziped request body and returned uncompressed data.
If request body is emty - nothing do

Args:

	l logger: object, implemented logger interface
*/
func DecompressRequestMiddleware(l logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") != "gzip" {
				l.Debug("Request isn't compressed")
				next.ServeHTTP(w, r)
				return
			}

			// Check for GET-requests with empty body
			if r.Body != nil {
				// decompress request
				gzr, err := gzip.NewReader(r.Body)
				if err != nil {
					// TO-DO: добавить обработку ошибки - изменение статус кода ответа
					l.Error("cannot decompress request", "error", err.Error())
					return
				}
				defer gzr.Close()
				r.Body = io.NopCloser(gzr)
			}

			next.ServeHTTP(w, r)
		})
	}
}

/*
SaveStorageToFile middleware save all storager data into dst
*/
func SaveStorageToFile(l logger, s storager, dst io.WriteCloser) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrappedWriter, r)
			if wrappedWriter.statusCode == http.StatusOK {
				err := services.WriteMetricsWithTimestamp(s, dst)
				if err != nil {
					l.Error("error writing to file", "error", err.Error())
					return
				}
				l.Debug("data writed")
			}
		})
	}
}
