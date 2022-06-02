package grpcversion

import (
	"context"

	"biller/lib/version"

	"google.golang.org/grpc"
)

type service struct {
	UnimplementedVersionServiceServer
}

func (s service) GetVersion(ctx context.Context, req *GetVersionRequest) (*GetVersionResponse, error) {
	return &GetVersionResponse{
		Version:   version.Version,
		BuildTime: version.BuildTime,
	}, nil
}

func Register(registrar grpc.ServiceRegistrar) {
	RegisterVersionServiceServer(registrar, service{})
}
