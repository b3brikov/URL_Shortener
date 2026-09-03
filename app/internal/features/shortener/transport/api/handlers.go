package shortenerapi

import (
	"URLShortener/internal/core/models"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// мувнуть в core потом
type contextKey string

const userIDKey contextKey = "user_id"

type Service interface {
	CreateNewCode(ctx context.Context, url string, userID int) (models.URLModel, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type Handlers struct {
	service Service
	logger  *slog.Logger
}

func ShortenerHandlers(service Service, logger *slog.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/newcode", h.createCode)
}

type createCodeInput struct {
	OriginalURL string `json:"original_url"`
}

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}

func (h *Handlers) createCode(w http.ResponseWriter, r *http.Request) {
	var input createCodeInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// handleError
		return
	}

	userID, ok := GetUserID(r.Context())
	if !ok {
		// 500
		return
	}

	res, err := h.service.CreateNewCode(
		r.Context(),
		input.OriginalURL,
		userID)

	if err != nil {
		// handleError
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error(
			"cannot encode response",
			slog.Any("error", err),
		)
	}
}
