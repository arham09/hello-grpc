package hello

import (
	"context"
	"fmt"

	"github.com/arham09/hello-grpc/pb/hello"
	hello_pb "github.com/arham09/hello-grpc/pb/hello"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HelloService struct {
	hello.UnimplementedGreeterServer
}

func newHelloService() *HelloService {
	return &HelloService{}
}

func RegisterService() func(srv *grpc.Server) error {
	return func(srv *grpc.Server) error {
		s := newHelloService()

		hello_pb.RegisterGreeterServer(srv, s)
		return nil
	}
}

func RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error {
	return hello_pb.RegisterGreeterHandlerFromEndpoint(ctx, mux, addr, opts)
}

func (s *HelloService) Hello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		Success: true,
		Message: fmt.Sprintf("Hello %s", req.Name),
		Name:    req.Name,
	}, nil
}

func (s *HelloService) Ping(ctx context.Context, _ *emptypb.Empty) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		Success: true,
		Message: "Pong",
	}, nil
}

func (s *HelloService) Ping2(ctx context.Context, _ *emptypb.Empty) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		Success: true,
		Message: "Pong2",
	}, nil
}
