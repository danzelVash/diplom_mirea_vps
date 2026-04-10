package internal

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func (a *App) init() error {
	lis, err := net.Listen("tcp", a.grpcAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.grpcListener = lis
	a.grpcServer = grpc.NewServer()
	return nil
}
