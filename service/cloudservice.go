package service

import (
	"net/http"

	"google.golang.org/grpc"
)

// CloudService represents an isolated component that serves http and/or grpc
// requests.
type CloudService interface {

	// Init initializes the service components. This method
	// should be called once on service start up.
	Init() error

	// REST returns the [http.Handler] that is registered for
	// this service as well as the config for the http server.
	// Returns (nil, nil) if the service is not serving http
	// requests.
	REST() (http.Handler, *RESTConfig)

	// GRPC returns the [grpc.Server] that is registered for
	// this service as well as the config for that server.
	// Returns (nil, nil) if the service is not serving gRPC
	// requests.
	GRPC() (*grpc.Server, *GRPCConfig)
}

// BaseService is a service implementation, which can be used as a base for
// other cloud services. The service provides a basic implementation of the
// [CloudService] interface methods.
type BaseService struct {
	// restHandler and restCfg are used to start an http server,
	// that serves http requests to the rest api of the service.
	restHandler http.Handler
	restCfg     *RESTConfig

	// grpcServer serves requests to the grpc api of the service.
	grpcServer *grpc.Server
	grpcCfg    *GRPCConfig
}

var _ CloudService = (*BaseService)(nil)

// Init implements the [CloudService] interface.
func (s *BaseService) Init() error { return nil }

// REST implements the [CloudService] interface.
func (s *BaseService) REST() (http.Handler, *RESTConfig) { return s.restHandler, s.restCfg }

// GRPC implements the [CloudService] interface.
func (s *BaseService) GRPC() (*grpc.Server, *GRPCConfig) { return s.grpcServer, s.grpcCfg }

// RegisterREST takes an [http.Handler] and registers that handler with the
// service. In addition, the config for the http server must also be provided.
// This function will simply set the corresponding fields of the [BaseService]
// so that they are available when calling `service.REST()`.
func (s *BaseService) RegisterREST(h http.Handler, cfg *RESTConfig) {
	s.restHandler = h
	s.restCfg = cfg
}

// RegisterGRPC takes a [grpc.Server] and registers that server with the
// service. In addition, the config for that server must also be provided.
// This function will simply set the corresponding fields of the [BaseService]
// so that they are available when calling `service.GRPC()`.
func (s *BaseService) RegisterGRPC(srv *grpc.Server, cfg *GRPCConfig) {
	s.grpcServer = srv
	s.grpcCfg = cfg
}
