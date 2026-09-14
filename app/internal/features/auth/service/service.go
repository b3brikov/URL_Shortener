package authservice

import (
	"URLShortener/internal/core/models"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	EncryptError  = errors.New("encrypting password internal error")
	EmptyPassword = errors.New("empty password")
	EmptyEmail    = errors.New("empty email")
)

type Cache interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	GetValue(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type Repository interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, username, email, hash string) error
}

type Service struct {
	accessTTL, refreshTTL time.Duration
	cache                 Cache
	repo                  Repository
	secret                []byte
}

func NewService(repo Repository, cache Cache, accessTTL, refreshTTL time.Duration, secret []byte) *Service {
	return &Service{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		repo:       repo,
		secret:     secret,
		cache:      cache,
	}
}

type Claims struct {
	UserID string
	jwt.RegisteredClaims
}

func (t *Service) generateAccessToken(id string) (string, error) {
	jti := uuid.New().String()

	claims := &Claims{UserID: id,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.accessTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()), ID: jti},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *Service) generateRefreshToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (t *Service) CreateNewUser(ctx context.Context, username, email, password string) error {
	if password == "" {
		return EmptyPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return EncryptError
	}
	return t.repo.CreateUser(ctx, username, email, string(hash))
}

func (t *Service) Authorize(ctx context.Context, email, password string) (models.TokenPair, error) {
	if email == "" {
		return models.TokenPair{}, EmptyEmail
	}
	if password == "" {
		return models.TokenPair{}, EmptyPassword
	}
	user, err := t.repo.GetUser(ctx, email)
	if err != nil {
		return models.TokenPair{}, err
	}

	if err := user.CompareHash([]byte(password)); err != nil {
		return models.TokenPair{}, errors.New("incorrect pasword")
	}
	userID := strconv.Itoa(user.ID)
	access, err := t.generateAccessToken(userID)
	if err != nil {
		return models.TokenPair{}, err
	}
	refresh := t.generateRefreshToken()
	ctxRefreshToken, cancelS := context.WithTimeout(context.Background(), 2*time.Second)
	t.cache.Set(ctxRefreshToken, models.RefreshNameSpace+refresh, userID, t.refreshTTL)
	cancelS()

	return models.TokenPair{
		UserID:     user.ID,
		Access:     access,
		Refresh:    refresh,
		AccessTTL:  t.accessTTL,
		RefreshTTL: t.refreshTTL,
	}, nil
}

func (t *Service) validateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return t.secret, nil
	})

	if err != nil {
		return nil, errors.New("token not accepted")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserID == "" {
		return nil, errors.New("missing user id")
	}
	log.Println(claims.UserID)
	return claims, nil
}

func (t *Service) Validate(tokenString string) (int, error) {
	claims, err := t.validateAccessToken(tokenString)
	if err != nil {
		return 0, err
	}
	res, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return 0, errors.New("invalid user id value")
	}
	return res, nil
}

func (t *Service) RefreshAccessToken(ctx context.Context, refresh string) (models.TokenPair, error) {
	key := models.RefreshNameSpace + refresh
	userID, err := t.cache.GetValue(ctx, key)

	if err != nil {
		return models.TokenPair{}, err
	}

	access, err := t.generateAccessToken(userID)
	if err != nil {
		return models.TokenPair{}, err
	}

	return models.TokenPair{
		Access:     access,
		Refresh:    refresh,
		AccessTTL:  t.accessTTL,
		RefreshTTL: t.refreshTTL,
	}, nil
}
