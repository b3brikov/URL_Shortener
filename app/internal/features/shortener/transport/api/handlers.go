package shortenerapi

import (
	"URLShortener/internal/core/models"
	"URLShortener/internal/core/postgres"
	utilshttp "URLShortener/internal/core/transport/http/utils"
	shortener "URLShortener/internal/features/shortener/service"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postgres.ErrCodeNotFound):
		utilshttp.SendError(w, "url not found", http.StatusNotFound)
	case errors.Is(err, shortener.CannotCreateNewUnique):
		utilshttp.SendError(w, "internal error", http.StatusInternalServerError)
	case errors.Is(err, context.DeadlineExceeded):
		utilshttp.SendError(w, "time out", http.StatusInternalServerError)
	default:
		utilshttp.SendError(w, "unexpected error", http.StatusInternalServerError)
	}
}

const shortCode = "shortCode"

type Middleware interface {
	Chain(next http.Handler) http.Handler
}

type Service interface {
	CreateNewCode(ctx context.Context, url string, userID *int) (models.URLModel, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type Handlers struct {
	middleware Middleware
	service    Service
	logger     *slog.Logger
}

func ShortenerHandlers(service Service, middleware Middleware, logger *slog.Logger) *Handlers {
	return &Handlers{
		middleware: middleware,
		service:    service,
		logger:     logger,
	}
}

func (h *Handlers) Register(mux *http.ServeMux) {
	submux := http.NewServeMux()

	submux.HandleFunc("POST /api/newcode", h.createCode)
	submux.HandleFunc("GET /api/{shortCode}", h.redirectCode)

	mux.Handle("/", h.middleware.Chain(submux))
}

type createCodeInput struct {
	OriginalURL string `json:"original_url"`
}

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(models.UserIDKey).(int)
	return id, ok
}

func (h *Handlers) createCode(w http.ResponseWriter, r *http.Request) {
	var input createCodeInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utilshttp.SendError(w, "bad request", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r.Context())

	if !ok {
		utilshttp.SendError(w, "user id not found", http.StatusInternalServerError)
		return
	}

	res, err := h.service.CreateNewCode(
		r.Context(),
		input.OriginalURL,
		&userID)

	if err != nil {
		handleError(w, err)
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

func (h *Handlers) redirectCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue(shortCode)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	url, err := h.service.GetOriginalURL(ctx, code)

	if err != nil {
		handleError(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}
