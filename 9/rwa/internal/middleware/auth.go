package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"rwa/internal/model"
	"strings"
)

type Service interface {
	Check(ctx context.Context, token string) (model.Session, error)
}

var (
	noAuthUrls = map[string]struct{}{
		"/api/users":       struct{}{},
		"/api/users/login": struct{}{},
		// "/api/articles":    struct{}{},
	}
)

type ctxKey int

const sessionKey ctxKey = 1

var (
	ErrNoAuth = errors.New("No session found")
)

func SessionFromContext(ctx context.Context) (model.Session, error) {
	sess, ok := ctx.Value(sessionKey).(model.Session)
	if !ok {
		return model.Session{}, ErrNoAuth
	}
	return sess, nil
}

func getTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	return strings.TrimPrefix(auth, "Token ")
}

func AuthMiddleware(s Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if _, ok := noAuthUrls[r.URL.Path]; ok && r.URL.RawQuery == "" {

				log.Printf("Url: %s, Values: %s no token\n", r.URL.Path, r.URL.RawQuery)
				next.ServeHTTP(w, r)
				return
			}

			token := getTokenFromHeader(r)
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sess, err := s.Check(r.Context(), token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
