package service

import "time"

// RESTConfig encapsulates the configuration for the rest component of the service.
type RESTConfig struct {
	// Listen is the port on which the REST endpoints of this
	// service will be registered.
	Listen string `env:"HTTP_SERVER_LISTEN" envDefault:":10080"`

	ReadHeaderTimeout time.Duration `env:"HTTP_SERVER_READ_HEADER_TIMEOUT" envDefault:"10s"`
	ReadTimeout       time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout      time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT" envDefault:"30s"`

	DumpRequests bool `env:"HTTP_SERVER_DUMP_REQUESTS"`
}

// GRPCConfig encapsulates the configuration for the rest component of the service.
type GRPCConfig struct {
	// Listen is the port on which the grpc endpoints of this
	// service will be registered.
	Listen string `env:"GRPC_SERVER_LISTEN" envDefault:":10090"`

	// ClientTimeout is a timeout used for RPC HTTP clients. #courier
	ClientTimeout time.Duration `enc:"RPC_CLIENT_TIMEOUT"`
}
