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

	// Bus returns the [MessageBus] that is used for publishing
	// and subscribing to messages. Returns nil if the service is
	// not publishing/subscribing messages.
	Bus() MessageBus

	// Events returns a map of events for which the service is
	// listening, and their associated handlers. Returns nil if
	// the service is not listening to events.
	Events() map[string]EventHandler
}
