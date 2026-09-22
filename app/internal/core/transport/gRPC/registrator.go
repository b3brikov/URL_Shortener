package registratorgrpc

import (
	"net"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	Server *grpc.Server
}

func (g *GRPCServer) Run() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}
	return g.Server.Serve(listener)
}
