package api

import (
	"URLShortener/internal/models"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
		fmt.Println("USER_ID", uid)
		c.Next()
	}
}

func (h *Handler) Register(c *gin.Context) {
	var login models.Login
	err := c.BindJSON(&login)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request error"})
		return
	}

	if err := h.AuthService.CreateNewUser(c.Request.Context(), "", login.Email, login.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Login(c *gin.Context) {
	var login models.Login
	err := c.BindJSON(&login)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request error"})
		return
	}

	user, err := h.AuthService.Authorize(c.Request.Context(), login.Email, login.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "accepted",
		"credentials": user,
	})
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
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	code := h.Service.GenerateCode()
	c.JSON(200, gin.H{
		"short_url": code,
	})
}

func (h *Handler) CreateNewCode(c *gin.Context) {
	var url models.OriginalURL
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := c.BindJSON(&url); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	code, err := h.Service.CreateNewCode(c.Request.Context(), url.Original, userID)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, code)
}

func (h *Handler) GoToOriginal(c *gin.Context) {
	code := c.Param("code")

	original, err := h.Service.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, original)
}
