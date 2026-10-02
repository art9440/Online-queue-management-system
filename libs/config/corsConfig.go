package config

import (
	"context"
	"strings"
)

// DefaultCORSAllowedOrigins covers the Vite dev server reached via localhost, IPv4 and IPv6 loopback.
const DefaultCORSAllowedOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://[::1]:5173"

// LoadCORSAllowedOrigins reads CORS_ALLOWED_ORIGINS as a comma-separated list of origins.
func LoadCORSAllowedOrigins(ctx context.Context) []string {
	raw := Get(ctx, "CORS_ALLOWED_ORIGINS", DefaultCORSAllowedOrigins)

	var origins []string
	for origin := range strings.SplitSeq(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}
