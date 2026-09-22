package middleware

import (
	"URLShortener/internal/core/models"
	utilshttp "URLShortener/internal/core/transport/http/utils"
	"context"
	"net/http"
)

type Auth interface {
	Validate(tokenString string) (int, error)
}

type Middleware struct {
	auth Auth
}

func NewMiddleware(auth Auth) *Middleware {
	return &Middleware{
		auth: auth,
	}
}

func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				w.WriteHeader(http.StatusInternalServerError)

			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		userID, err := m.auth.Validate(token)
		if err != nil {
			utilshttp.SendError(w, "invalid access token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Chain(next http.Handler) http.Handler {

	middlewares := []func(http.Handler) http.Handler{m.Recover, m.AuthMiddleware}

	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}
