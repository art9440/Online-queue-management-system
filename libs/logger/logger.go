package logger

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

type Config struct {
	Level        slog.Level
	JSON         bool
	Source       bool
	Service      string
	LogstashAddr string
}

func New(cfg Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.Source,
	}

	if cfg.Service == "" {
		cfg.Service = os.Getenv("SERVICE_NAME")
	}
	if cfg.LogstashAddr == "" {
		cfg.LogstashAddr = os.Getenv("LOGSTASH_TCP_ADDR")
	}

	var h slog.Handler
	if cfg.JSON {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	if cfg.Service != "" {
		h = h.WithAttrs([]slog.Attr{slog.String("service", cfg.Service)})
	}

	if cfg.LogstashAddr != "" {
		var logstashHandler slog.Handler = slog.NewJSONHandler(newTCPWriter(cfg.LogstashAddr), opts)
		if cfg.Service != "" {
			logstashHandler = logstashHandler.WithAttrs([]slog.Attr{slog.String("service", cfg.Service)})
		}
		h = multiHandler{handlers: []slog.Handler{h, logstashHandler}}
	}

	return slog.New(h)
}

type multiHandler struct {
	handlers []slog.Handler
}

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			_ = handler.Handle(ctx, record.Clone())
		}
	}
	return nil
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return multiHandler{handlers: handlers}
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return multiHandler{handlers: handlers}
}

type tcpWriter struct {
	addr      string
	mu        sync.Mutex
	conn      net.Conn
	nextRetry time.Time
}

func newTCPWriter(addr string) io.Writer {
	return &tcpWriter{addr: addr}
}

func (w *tcpWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn == nil {
		if time.Now().Before(w.nextRetry) {
			return len(p), nil
		}
		conn, err := net.DialTimeout("tcp", w.addr, 200*time.Millisecond)
		if err != nil {
			w.nextRetry = time.Now().Add(5 * time.Second)
			return len(p), nil
		}
		w.conn = conn
	}

	_ = w.conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := w.conn.Write(p); err != nil {
		_ = w.conn.Close()
		w.conn = nil
		w.nextRetry = time.Now().Add(5 * time.Second)
	}

	return len(p), nil
}

type ctxKey struct{}

func With(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func From(ctx context.Context) *slog.Logger {
	if v := ctx.Value(ctxKey{}); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
