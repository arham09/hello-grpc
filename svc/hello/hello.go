package hello

import (
	"context"
	"fmt"
	"log"

	"github.com/arham09/hello-grpc/pb/hello"
	hello_pb "github.com/arham09/hello-grpc/pb/hello"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/open-feature/go-sdk/openfeature"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HelloService struct {
	hello.UnimplementedGreeterServer

	ff *openfeature.Client
}

func newHelloService(ff *openfeature.Client) *HelloService {
	return &HelloService{
		ff: ff,
	}
}

func RegisterService(ff *openfeature.Client) func(srv *grpc.Server) error {
	return func(srv *grpc.Server) error {
		s := newHelloService(ff)

		hello_pb.RegisterGreeterServer(srv, s)
		return nil
	}
}

func RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error {
	return hello_pb.RegisterGreeterHandlerFromEndpoint(ctx, mux, addr, opts)
}

func (s *HelloService) Hello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	val, err := s.ff.BooleanValue(context.Background(), "enable_feature_a", false, openfeature.EvaluationContext{})
	if err != nil {
		return nil, err
	}

	log.Printf("FF %+v \n", val)

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
