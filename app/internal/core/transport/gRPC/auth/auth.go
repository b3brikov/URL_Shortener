package auth

import (
	"URLShortener/internal/core/transport/gRPC/proto"
	"URLShortener/internal/core/models"
	"context"
	"errors"
)

var EmptyCreds = errors.New("empty credentials")
var NotFoundUser = errors.New("user not found")

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
	auth Auth
	proto.UnimplementedShortenerServer
}

func NewHandler(auth Auth) *Handler {
	return &Handler{
		auth:                         auth,
		UnimplementedShortenerServer: proto.UnimplementedShortenerServer{},
	}
}

func (h *Handler) Login(ctx context.Context,
	lr *proto.LoginReq) (*proto.LoginResp, error) {
	e := lr.Email
	p := lr.Password
	if e == "" || p == "" {
		return nil, EmptyCreds
	}

	tp, err := h.auth.Authorize(ctx, e, p)
	if err != nil {
		return nil, NotFoundUser
	}

	return &proto.LoginResp{
		Id:         int64(tp.UserID),
		Access:     tp.Access,
		Refresh:    tp.Refresh,
		AccessTtl:  int64(tp.AccessTTL),
		RefreshTtl: int64(tp.RefreshTTL),
	}, nil
}
