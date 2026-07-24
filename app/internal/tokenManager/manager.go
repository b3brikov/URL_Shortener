package tokenmanager

import (
	"URLShortener/internal/models"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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

type Repository interface {
	SetToken(ctx context.Context, key, value string, ttl time.Duration) error
	GetToken(ctx context.Context, key string) (string, error)
	DeleteToken(ctx context.Context, key string) error
	GetUser(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, username, email, hash string) error
}

type TokenManager struct {
	accessTTL, refreshTTL time.Duration
	repo                  Repository
	secret                []byte
}

func NewTokenManager(repo Repository, accessTTL, refreshTTL time.Duration, secret []byte) *TokenManager {
	return &TokenManager{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		repo:       repo,
		secret:     secret,
	}
}

type Claims struct {
	UserID string
	jwt.RegisteredClaims
}

func (t *TokenManager) generateAccessToken(id int) (string, error) {
	jti := uuid.New().String()

	claims := &Claims{UserID: strconv.Itoa(id),
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.accessTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()), ID: jti},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenManager) generateRefreshToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (t *TokenManager) CreateNewUser(ctx context.Context, username, email, password string) error {
	if password == "" {
		return EmptyPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return EncryptError
	}
	return t.repo.CreateUser(ctx, username, email, string(hash))
}

// -------------------------------------------------------------------------------------------------------

func (t *TokenManager) Authorize(ctx context.Context, email, password string) (models.TokenPair, error) {
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
	access, err := t.generateAccessToken(user.ID)
	if err != nil {
		return models.TokenPair{}, err
	}
	refresh := t.generateRefreshToken()
	go func() {
		userID := strconv.Itoa(user.ID)
		ctxRefreshToken, cancelS := context.WithTimeout(context.Background(), 2*time.Second)
		t.repo.SetToken(ctxRefreshToken, models.RefreshNameSpace+userID, refresh, t.refreshTTL)
		cancelS()
	}()
	return models.TokenPair{
		UserID:     user.ID,
		Access:     access,
		Refresh:    refresh,
		AccessTTL:  t.accessTTL,
		RefreshTTL: t.refreshTTL,
	}, nil
}

func (t *TokenManager) validateAccessToken(tokenString string) (*Claims, error) {
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

	return claims, nil
}

func (t *TokenManager) Validate(tokenString string) (int, error) {
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

func (t *TokenManager) RefreshTokens(ctx context.Context, userID int, refresh string) (models.TokenPair, error) {
	key := models.RefreshNameSpace + strconv.Itoa(userID)
	token, err := t.repo.GetToken(ctx, key)

	if err != nil {
		return models.TokenPair{}, err
	}

	if token != refresh {
		return models.TokenPair{}, errors.New("invalid refresh token")
	}
	access, err := t.generateAccessToken(userID)
	if err != nil {
		return models.TokenPair{}, err
	}

	newRefresh := t.generateRefreshToken()

	ctxRefreshToken, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := t.repo.SetToken(ctxRefreshToken, key, newRefresh, t.refreshTTL); err != nil {
		return models.TokenPair{}, errors.New("cannot create new refresh token")
	}
	return models.TokenPair{
		Access:     access,
		Refresh:    newRefresh,
		AccessTTL:  t.accessTTL,
		RefreshTTL: t.refreshTTL,
	}, nil
}
