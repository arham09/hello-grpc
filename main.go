package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arham09/hello-grpc/svc"
	"github.com/arham09/hello-grpc/svc/hello"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func main() {
	// header match function for accepting custom headers
	incomingHeaderfn := func(key string) (string, bool) {
		switch key {
		case "tracestate":
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}

	outgoingHeaderfn := func(key string) (string, bool) {
		switch key {
		case "X-Request-Id":
			return key, true
		case "x-request-id":
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}

	grpcSrv, err := newGRPCServer()
	if err != nil {
		log.Fatalln("Failed to create gRPC server:", err)
	}

	srv := Setup(":55211", grpcSrv,
		AddGatewayFunc(hello.RegisterGateway),
		SetHTTPHost(":8000"),
		AddGatewayMuxOption(runtime.WithIncomingHeaderMatcher(incomingHeaderfn), runtime.WithOutgoingHeaderMatcher(outgoingHeaderfn)),
	)

	srv.Run()
}

func enableCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, HEAD, OPTIONS")
			if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
			}
			h.ServeHTTP(w, r)
	})
}

func newGRPCServer() (*grpc.Server, error) {
	opts := []grpc.ServerOption{
		grpc.ConnectionTimeout(300 * time.Second),
	}

	srv := grpc.NewServer(opts...)
	if err := svc.RegisterServices(
		srv,
		hello.RegisterService(),
	); err != nil {
		return nil, errors.Wrap(err, "failed to register gRPC service")
	}

	return srv, nil
}

type Server struct {
	host       string
	httpHost   string
	metric     *http.Server
	grpcSrv    *grpc.Server
	muxOptions []runtime.ServeMuxOption
	gateways   []RegisterGatewayFunc
}

func Setup(host string, grpcSrv *grpc.Server, opts ...func(*Server)) *Server {
	s := &Server{
		host:    host,
		grpcSrv: grpcSrv,
		muxOptions: []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
		},
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Server) Run(closeFn ...func()) {
	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	go func() {
		log.Println("Serving gRPC connection on", s.host)
		if err := s.grpcSrv.Serve(lis); err != nil {
			log.Fatalln("Failed to listen grpc server:", err)
		}
	}()

	pbMessageOpt := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions:   protojson.MarshalOptions{UseProtoNames: true},
		UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: true},
	})

	s.muxOptions = append(s.muxOptions, pbMessageOpt)

	mux := runtime.NewServeMux(s.muxOptions...)
	for _, gw := range s.gateways {
		if err := gw.RegisterGateway(context.Background(), mux, s.host, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}); err != nil {
			log.Fatalln("Failed to register gateway:", err)
		}
	}

	gwServer := &http.Server{
		Addr:    s.httpHost,
		Handler: enableCors(mux),
	}

	if s.httpHost != "" {
		go func() {
			log.Println("Serving gRPC-Gateway connection on", s.httpHost)
			if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln("Failed to listen grpc gateway server:", err)
			}
		}()
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Println("server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		for _, fn := range closeFn {
			fn()
		}

		if s.metric != nil {
			s.metric.Shutdown(ctx)
		}

		s.grpcSrv.GracefulStop()
		cancel()
	}()

	if s.httpHost != "" {
		if err := gwServer.Shutdown(ctx); err != nil {
			log.Fatalln("Failed to shutdown server:", err)
		}
	}

	log.Println("web app exit properly")
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set http status code
	if vals := md.HeaderMD.Get("x-request-id"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-request-id")
		delete(w.Header(), "Grpc-Metadata-X-Request-Id")
		w.WriteHeader(code)
	}

	return nil
}

type RegisterGatewayFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

func (f RegisterGatewayFunc) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, gwAddr string, opts []grpc.DialOption) error {
	return f(ctx, mux, gwAddr, opts)
}

func AddGatewayFunc(gateways ...RegisterGatewayFunc) func(h *Server) {
	return func(h *Server) { h.gateways = append(h.gateways, gateways...) }
}

func AddGatewayMuxOption(muxOpts ...runtime.ServeMuxOption) func(h *Server) {
	return func(h *Server) { h.muxOptions = append(h.muxOptions, muxOpts...) }
}

func SetHTTPHost(httpHost string) func(h *Server) {
	return func(h *Server) { h.httpHost = httpHost }
}

func SetMetricHost(metricHost string) func(h *Server) {
	return func(h *Server) { h.metric = &http.Server{Addr: metricHost} }
}
