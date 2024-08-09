package svc

import "google.golang.org/grpc"

func RegisterServices(srv *grpc.Server, services ...func(srv *grpc.Server) error) error {
	for _, svc := range services {
		if err := svc(srv); err != nil {
			return err
		}
	}
	return nil
}
