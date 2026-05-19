package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userKey contextKey = "user"

func Middleware(parser *TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("access_token")
			if err != nil || cookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := parser.ParseAccessToken(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) *AccessClaims {
	user, _ := ctx.Value(userKey).(*AccessClaims)
	return user
}

// ContextWithUser is a test helper function to create a context with user claims
func ContextWithUser(ctx context.Context, user *AccessClaims) context.Context {
	return context.WithValue(ctx, userKey, user)
}
