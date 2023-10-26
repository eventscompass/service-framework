package service

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

// BaseService is a service implementation, which can be used as a base for
// other cloud services. The service provides a dummy implementations of the
// [CloudService] interface methods.
type BaseService struct{}

var _ CloudService = (*BaseService)(nil)

// Init implements the [CloudService] interface.
func (s *BaseService) Init(_ context.Context) error { return nil }

// REST implements the [CloudService] interface.
func (s *BaseService) REST() http.Handler { return nil }

// GRPC implements the [CloudService] interface.
func (s *BaseService) GRPC() *grpc.Server { return nil }

// Bus implements the [CloudService] interface.
func (s *BaseService) Bus() MessageBus { return nil }

// Events implements the [CloudService] interface.
func (s *BaseService) Events() map[string]EventHandler { return nil }
