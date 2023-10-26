package service

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

// CloudService represents an isolated component that serves http and/or grpc
// requests.
type CloudService interface {

	// Init initializes the service components. This method
	// should be called once on service start up.
	Init(_ context.Context) error

	// REST returns the [http.Handler] that is registered for
	// this service. Returns nil if the service is not serving
	// http requests.
	REST() http.Handler

	// GRPC returns the [grpc.Server] that is registered for
	// this service. Returns nil if the service is not serving
	// grpc requests.
	GRPC() *grpc.Server
}

// BaseService is a service implementation, which can be used as a base for
// other cloud services. The service provides a basic implementation of the
// [CloudService] interface methods.
type BaseService struct {
	// restHandler is used to start an http server, that serves
	// http requests to the rest api of the service.
	restHandler http.Handler

	// grpcServer serves requests to the grpc api of the service.
	grpcServer *grpc.Server
}

var _ CloudService = (*BaseService)(nil)

// Init implements the [CloudService] interface.
func (s *BaseService) Init(_ context.Context) error { return nil }

// REST implements the [CloudService] interface.
func (s *BaseService) REST() http.Handler { return s.restHandler }

// GRPC implements the [CloudService] interface.
func (s *BaseService) GRPC() *grpc.Server { return s.grpcServer }

// RegisterREST takes an [http.Handler] and registers that handler with the
// service. This function will simply set the corresponding field of the
// [BaseService] so that this field is available when calling `service.REST()`.
func (s *BaseService) RegisterREST(h http.Handler) { s.restHandler = h }

// RegisterGRPC takes a [grpc.Server] and registers that server with the
// service. This function will simply set the corresponding field of the
// [BaseService] so that this field is available when calling `service.GRPC()`.
func (s *BaseService) RegisterGRPC(srv *grpc.Server) { s.grpcServer = srv }
