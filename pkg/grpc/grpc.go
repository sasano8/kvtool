package grpc

import (
	"context"

	"google.golang.org/grpc"
)

type GetFooRequest struct {
}

type GetFooResponse struct {
}

type FooServiceClient interface {
	GetFoo(ctx context.Context, in *GetFooRequest, opts ...grpc.CallOption) (*GetFooResponse, error)
}

// type LocalFooClient struct {
// 	svc pb.FooServiceServer
// }
