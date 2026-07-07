package api

import (
	"URLShortener/internal/models"
	"URLShortener/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	InvalidIDType = errors.New("invalid user_id type")
)

type Service interface {
	GenerateCode() string
	CreateNewCode(ctx context.Context, url string, userID int) (models.URLModel, error)
	GetOriginalURL(ctx context.Context, url string) (string, error)
}

type Auth interface {
	Authorize(ctx context.Context, email string, password string) (models.TokenPair, error)
	CreateNewUser(ctx context.Context, username string, email string, password string) error
	RefreshTokens(ctx context.Context, userID int, refresh string) (models.TokenPair, error)
	Validate(tokenString string) (int, error)
}

type Handler struct {
	Service     Service
	AuthService Auth
}

func NewHandler(service Service, auth Auth) *Handler {
	return &Handler{
		Service:     service,
		AuthService: auth,
	}
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}

func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCodeNotFound):
		Fail(c, http.StatusNotFound, "not found")
	case errors.Is(err, service.CannotCreateNewUnique):
		Fail(c, http.StatusInternalServerError, "cannot create new code")
	case errors.Is(err, service.ErrUnexpectedError):
		Fail(c, http.StatusInternalServerError, "unexpected error")
	case errors.Is(err, service.ErrTimeOut):
		Fail(c, http.StatusRequestTimeout, "timeout")
	default:
		Fail(c, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		log.Println(header)
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing autorization header"})
			c.Abort()
			return
		}
		parts := strings.Split(header, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Println(1)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}
		uid, err := h.AuthService.Validate(parts[1])
		if err != nil {
			log.Println(err)
			log.Println(2)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", uid)
		c.Next()
	}
}

func (h *Handler) Register(c *gin.Context) {
	var login models.Login
	err := c.BindJSON(&login)

	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.AuthService.CreateNewUser(c.Request.Context(), "", login.Email, login.Password); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Login(c *gin.Context) {
	var login models.Login
	err := c.BindJSON(&login)

	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.AuthService.Authorize(c.Request.Context(), login.Email, login.Password)
	if err != nil {
		Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	Success(c, http.StatusAccepted, user)
}

func GetUserID(c *gin.Context) (int, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user_id not found in context")
	}

	id, ok := userID.(int)
	if !ok {
		return 0, errors.New("invalid user_id type")
	}

	return id, nil
}

func (h *Handler) TestShorten(c *gin.Context) {
	var url models.OriginalURL

	if err := c.BindJSON(&url); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	code := h.Service.GenerateCode()
	Success(c, http.StatusCreated, code)
}

func (h *Handler) CreateNewCode(c *gin.Context) {
	var url models.OriginalURL
	userID, err := GetUserID(c)
	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.BindJSON(&url); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}

	code, err := h.Service.CreateNewCode(c.Request.Context(), url.Original, userID)
	if err != nil {
		HandleError(c, err)
		return
	}
	Success(c, http.StatusCreated, code)
}

func (h *Handler) GoToOriginal(c *gin.Context) {
	code := c.Param("code")

	original, err := h.Service.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		Fail(c, http.StatusNotFound, "not found")
		return
	}

	c.Redirect(http.StatusFound, original)
}
