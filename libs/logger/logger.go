package logger

import (
	"context"
	"log/slog"
	"os"
)

type Config struct {
	Level   slog.Level
	JSON    bool
	Source  bool
	Service string
}

// New writes logs to stdout only; collecting and routing them is the job of the environment.
func New(cfg Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.Source,
	}

	if cfg.Service == "" {
		cfg.Service = os.Getenv("SERVICE_NAME")
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

	return slog.New(h)
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
