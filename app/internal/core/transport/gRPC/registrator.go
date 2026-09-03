package registratorgrpc

import (
	"URLShortener/internal/core/transport/gRPC/auth"
	"URLShortener/internal/core/transport/gRPC/proto"

	"net"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	Server *grpc.Server
}

func NewGRPCServer(auth *auth.Handler) *GRPCServer {
	srv := &grpc.Server{}
	proto.RegisterShortenerServer(srv, auth)

	return &GRPCServer{
		Server: srv,
	}
}

func (g *GRPCServer) Run() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}
	return g.Server.Serve(listener)
}
