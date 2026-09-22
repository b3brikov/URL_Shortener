package grpcapi

import (
	"URLShortener/internal/core/models"
	"URLShortener/internal/core/transport/gRPC/proto"
	"context"
	"fmt"
)

type Service interface {
	CreateNewCode(ctx context.Context, url string, userID *int) (models.URLModel, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type Server struct {
	proto.UnimplementedShortenerServer
	service Service
}

func NewServer(service Service) *Server {
	return &Server{
		UnimplementedShortenerServer: proto.UnimplementedShortenerServer{},
		service:                      service,
	}
}

func (s *Server) CreateNewCode(ctx context.Context, in *proto.CreateNewCodeReq) (*proto.CreateNewCodeResp, error) {
	if in.Url == "" {
		return nil, fmt.Errorf("url value is empty")
	}

	res, err := s.service.CreateNewCode(ctx, in.Url, nil)
	if err != nil {
		return nil, err
	}

	return &proto.CreateNewCodeResp{
		Url:  res.Original_url,
		Code: res.Short_code,
	}, nil
}
func (s *Server) GetOriginalUrl(ctx context.Context, in *proto.GetOriginalUrlReq) (*proto.GetOriginalUrlResp, error) {
	if in.Code == "" {
		return nil, fmt.Errorf("url value is empty")
	}

	res, err := s.service.GetOriginalURL(ctx, in.Code)
	if err != nil {
		return nil, err
	}

	return &proto.GetOriginalUrlResp{
		Url: res,
	}, nil
}
