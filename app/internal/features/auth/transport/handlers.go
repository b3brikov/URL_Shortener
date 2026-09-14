package transport

import (
	"URLShortener/internal/core/cache"
	"URLShortener/internal/core/models"
	utilshttp "URLShortener/internal/core/transport/http/utils"
	authservice "URLShortener/internal/features/auth/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type contextKey string

const userIDKey contextKey = "user_id"

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authservice.EmptyEmail):
		utilshttp.SendError(w, "empty email", http.StatusBadRequest)
	case errors.Is(err, authservice.EmptyPassword):
		utilshttp.SendError(w, "empty password", http.StatusBadRequest)
	case errors.Is(err, cache.ErrMissingValue):
		utilshttp.SendError(w, "invalid refresh token", http.StatusUnauthorized)
	default:
		utilshttp.SendError(w, "unexpected error", http.StatusInternalServerError)
	}
}

type Service interface {
	CreateNewUser(ctx context.Context, username, email, password string) error
	Authorize(ctx context.Context, email, password string) (models.TokenPair, error)
	RefreshAccessToken(ctx context.Context, refresh string) (models.TokenPair, error)
}

type Middleware interface {
	Chain(next http.Handler) http.Handler
}

type Handlers struct {
	middleware Middleware
	service    Service
}

func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Create)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/refresh", h.RefreshAccess)
}

func NewHandlers(service Service, middleware Middleware) *Handlers {
	return &Handlers{
		service:    service,
		middleware: middleware,
	}
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var userCreds models.Login
	err := json.NewDecoder(r.Body).Decode(&userCreds)
	if err != nil {
		utilshttp.SendError(w, "bad request", http.StatusBadRequest)
		return
	}
	tokPair, err := h.service.Authorize(r.Context(), userCreds.Email, userCreds.Password)

	if err != nil {
		handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokPair)
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var user models.NewUser

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		utilshttp.SendError(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := h.service.CreateNewUser(ctx, user.UserName, user.Email, user.Password); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}

func (h *Handlers) RefreshAccess(w http.ResponseWriter, r *http.Request) {
	var tokens models.TokenPair

	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		utilshttp.SendError(w, "bad request", http.StatusBadRequest)
		return
	}

	tokenPair, err := h.service.RefreshAccessToken(r.Context(), tokens.Refresh)

	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenPair)
}
