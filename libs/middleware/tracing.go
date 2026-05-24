package middleware

import (
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "Online-queue-management-system/http"

type tracingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *tracingResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *tracingResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func TraceRequests(next http.Handler) http.Handler {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		spanName := r.Method + " " + r.URL.Path

		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("url.path", r.URL.Path),
				attribute.String("url.query", r.URL.RawQuery),
				attribute.String("server.address", r.Host),
				attribute.String("user_agent.original", r.UserAgent()),
			),
		)
		defer span.End()

		tw := &tracingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(tw, r.WithContext(ctx))

		status := tw.status
		if status == 0 {
			status = http.StatusOK
		}

		span.SetAttributes(
			attribute.Int("http.response.status_code", status),
			attribute.String("http.response.status_class", strconv.Itoa(status/100)+"xx"),
		)
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
	})
}
