package http

import (
	"net/http"
	"time"

	"Online-queue-management-system/libs/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logger.From(r.Context())

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		log.Info(
			"incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.status,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
