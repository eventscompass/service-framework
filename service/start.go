package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

// Start takes a [CloudService], initializes it and starts a server that will
// accepts requests to the service. If the service exposes both rest and grpc
// apis, then two separate servers are started to serve each api.
// This is a blocking function that waits for the api server(s) to stop running.
func Start(s CloudService) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		if msg := recover(); msg != nil {
			log.Fatalf("Panic: %s\n", msg)
		}
	}()

	// Init the service components.
	if err := s.Init(ctx); err != nil {
		log.Fatalf("Failed to init service: %v", err)
	}

	// We will use an error group to start the server(s).
	// Start one goroutine that runs the server and another that waits to
	// perform a graceful shutdown. If any of the goroutines in the group
	// returns an error, the ctx is cancelled and the shutdown is triggered.
	g, ctx := errgroup.WithContext(ctx)

	if restHandler := s.REST(); restHandler != nil { // run the http server
		var cfg RESTConfig
		if err := env.Parse(&cfg); err != nil {
			log.Fatalf("Failed to parse rest environment variables: %v", err)
		}

		// The timeout values set on the server are used as TCP connection
		// deadlines. They will close the connection for read/write operations,
		// but will not stop the handler from processing the request. We wrap
		// the handler with a timeout in order to stop processing once it is too
		// late to write the result.
		// https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/
		h := http.TimeoutHandler(restHandler, cfg.WriteTimeout, "timeout")
		restSrv := &http.Server{
			// Increase the write timeout by a small margin (2s) to allow the
			// handler to write the timeout response in case of a timeout.
			WriteTimeout:      cfg.WriteTimeout + 2*time.Second,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			Addr:              cfg.Listen,
			Handler:           h,
		}
		g.Go(func() error { return restSrv.ListenAndServe() })
		// g.Go(func() error { return restSrv.ListenAndServeTLS("", "") })
		g.Go(func() error {
			<-ctx.Done() // block until context is cancelled
			return restSrv.Shutdown(context.Background())
		})
	}

	if grpcSrv := s.GRPC(); grpcSrv != nil { // run the grpc server
		var cfg GRPCConfig
		if err := env.Parse(&cfg); err != nil {
			log.Fatalf("Failed to parse grpc environment variables: %v", err)
		}

		lis, err := net.Listen("tcp", cfg.Listen)
		if err != nil {
			log.Fatalf("Failed to init grpc listener: %v", err)
		}
		defer lis.Close()
		g.Go(func() error { return grpcSrv.Serve(lis) })
		g.Go(func() error {
			<-ctx.Done() // block until context is cancelled
			grpcSrv.GracefulStop()
			return nil
		})
	}

	// In case the service is subscribed to a message broker, we will listen for
	// events inside the error group.
	if events := s.Events(); events != nil { // listen for events
		bus := s.Bus()
		if bus == nil {
			log.Fatalf("Message Bus not initialized")
		}
		for e, h := range events {
			event, handler := e, h
			g.Go(func() error { return bus.Subscribe(ctx, event, handler) })
		}
	}

	// Wait for interrupt signals. Upon receiving one of these signals, the ctx
	// will be cancelled, initiating a graceful shutdown of the server(s).
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, stopSignals...)
	g.Go(func() error {
		select {
		case sig := <-ch:
			log.Printf("Received stop signal: %q\n", sig)
			cancel()
		case <-ctx.Done():
			// Currently, the only way to cancel the context is by sending a
			// stop signal. However, if the context were to be cancelled in some
			// other way, then this goroutine would hang, blocking g.Wait().
			// For that reason we include this case here.
		}
		return nil
	})

	// Block until the service stops.
	if err := g.Wait(); err != nil {
		log.Fatalf("Received an error during serving: %v", err)
	}
}

var (
	// stopSignals are the interrupt and termination signals from the operating
	// system that the service listens for.
	stopSignals = []os.Signal{unix.SIGINT, unix.SIGTERM}
)
