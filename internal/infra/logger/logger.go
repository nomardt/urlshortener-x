package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// This middleware logs every server response duration, status code, response size
// and some request data: URI, method, request size, IP, datetime
func WithLogging() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			Log.Info("Handled request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("duration", duration.String()),
				zap.Int("size", responseData.size),
				zap.Int("status", responseData.status),
				zap.String("IP", r.RemoteAddr),
				zap.String("date", time.Now().Format("2006/01/02")),
				zap.String("time", time.Now().Format("15:04:05")),
			)
		}

		return http.HandlerFunc(logFn)
	}
}
