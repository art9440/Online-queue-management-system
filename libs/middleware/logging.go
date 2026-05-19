package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"time"

	"Online-queue-management-system/libs/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytesWritten += n
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := requestIDFromRequest(r)
		w.Header().Set("X-Request-ID", requestID)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
		}

		requestLog := logger.From(r.Context()).With(
			"request_id", requestID,
			"http_method", r.Method,
			"http_path", r.URL.Path,
		)
		ctx := logger.With(r.Context(), requestLog)

		next.ServeHTTP(lrw, r.WithContext(ctx))

		status := lrw.status
		if status == 0 {
			status = http.StatusOK
		}
		duration := time.Since(start)

		requestLog.Info(
			"http_request",
			"event", "http_request",
			"http_method", r.Method,
			"http_path", r.URL.Path,
			"http_query", r.URL.RawQuery,
			"http_status", status,
			"http_status_class", strconv.Itoa(status/100)+"xx",
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
			"bytes_written", lrw.bytesWritten,
			"remote_ip", remoteIP(r),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"referer", r.Referer(),
			"host", r.Host,
			"content_length", r.ContentLength,
		)
	})
}

func requestIDFromRequest(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	return hex.EncodeToString(raw[:])
}

func remoteIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if host, _, err := net.SplitHostPort(forwardedFor); err == nil {
			return host
		}
		return forwardedFor
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
