package service

import (
	"time"
)

// RESTConfig encapsulates the configuration for the rest component of the service.
type RESTConfig struct {
	// Listen is the port on which the REST endpoints of this
	// service will be registered.
	Listen string `env:"HTTP_SERVER_LISTEN" envDefault:":8080"`

	ReadHeaderTimeout time.Duration `env:"HTTP_SERVER_READ_HEADER_TIMEOUT" envDefault:"10s"`
	ReadTimeout       time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout      time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT" envDefault:"30s"`

	DumpRequests bool `env:"HTTP_SERVER_DUMP_REQUESTS"`
}

// GRPCConfig encapsulates the configuration for the rest component of the service.
type GRPCConfig struct {
	// Listen is the port on which the grpc endpoints of this
	// service will be registered.
	Listen string `env:"GRPC_SERVER_LISTEN" envDefault:":8081"`

	// ClientTimeout is a timeout used for RPC HTTP clients. #courier
	ClientTimeout time.Duration `env:"GRPC_CLIENT_TIMEOUT"`
}

// BusConfig encapsulates the configuration for the message bus used by the service.
type BusConfig struct {
	Host     string `env:"MESSAGE_BUS_HOST" envDefault:"rabbitmq"`
	Port     int    `env:"MESSAGE_BUS_PORT" envDefault:":5672"`
	Username string `env:"MESSAGE_BUS_USERNAME" envDefault:"user"`
	Password string `env:"MESSAGE_BUS_PASSWORD" envDefault:"password"`
}
