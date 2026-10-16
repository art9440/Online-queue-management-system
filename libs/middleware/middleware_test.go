package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	allowed := []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://[::1]:5173"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := CORSMiddleware(allowed, next)

	tests := []struct {
		name       string
		method     string
		origin     string
		wantOrigin string
		wantStatus int
	}{
		{"localhost", http.MethodPost, "http://localhost:5173", "http://localhost:5173", http.StatusTeapot},
		{"ipv4 loopback", http.MethodPost, "http://127.0.0.1:5173", "http://127.0.0.1:5173", http.StatusTeapot},
		{"ipv6 loopback", http.MethodPost, "http://[::1]:5173", "http://[::1]:5173", http.StatusTeapot},
		{"preflight allowed", http.MethodOptions, "http://[::1]:5173", "http://[::1]:5173", http.StatusOK},
		{"unknown origin", http.MethodPost, "http://evil.example", "", http.StatusTeapot},
		{"preflight unknown origin", http.MethodOptions, "http://evil.example", "", http.StatusOK},
		{"no origin", http.MethodGet, "", "", http.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Vary"); got != "Origin" {
				t.Errorf("Vary = %q, want %q", got, "Origin")
			}
		})
	}
}
