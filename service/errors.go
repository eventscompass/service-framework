package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// All exported functions, are only allowed to return the following errors or
// wrapped versions of those errors. The function should mention which errors it
// could possibly return. If your function returns one of the following errors
// you should mention so in the function's comment. The error [ErrUnexpected]
// has a different nature and therefore can be returned by any function without
// explicitly mentioning it. Context errors can be returned by any function that
// accepts a context as its first argument, without explicitly mentioning it.
//
// When returning one of these errors, it should be wrapped as follows:
//
//	fmt.Errorf("%w: info message: %v", ErrUnexpected, err)
//
// The order should be "%w: %v". Then error messages look like this:
//
//	# caller N-1: caller N: unexpected: info message: 'foo' not found
//	#   ^            ^         ^            ^                  ^
//	#   ` call stack ´         |        the problem     source of the problem
//	#                          |
//	#                classification of the problem
//	#
//
// This hierarchical structure of the error message is helpful when debugging.
// If you append the "unexpected" to the end, e.g fmt.Errorf("%v: %w", ...),
// then the error messages don't give away where in the stack something went
// wrong.
var (
	// ErrAlreadyExists is returned when the client requests to
	// create a resource that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrBadRequest is returned when the client submits a
	// request that cannot be understood and processed by the
	// service. Usually when the request body cannot be decoded,
	// or the request URL parameters cannot be handled, then this
	// error is returned.
	ErrBadRequest = errors.New("bad request")

	// ErrConnectionClosed is returned when the connection we are
	// trying to use is closed.
	ErrConnectionClosed = errors.New("connection closed")

	// ErrNotAllowed is returned when the requested action is not
	// allowed to be executed.
	ErrNotAllowed = errors.New("not allowed")

	// ErrNotFound is returned when the requested resource is not
	// found.
	ErrNotFound = errors.New("not found")

	// ErrSpaceFull is returned when the storage of the service
	// is full.
	ErrSpaceFull = errors.New("no space")

	// ErrTimeOut is returned when an operation performed by the
	// service is taking longer than the allowed time limit.
	ErrTimeOut = errors.New("time out")

	// ErrUnexpected is reserved for errors that look like they
	// would never happen. Instead of panicking use
	// ErrUnexpected. This error can be returned by any function
	// even if not explicitly mentioned.
	ErrUnexpected = errors.New("unexpected")
)

// Unexpected returns err if it's the error of ctx, otherwise it logs err and
// returns err wrapped in [ErrUnexpected].
func Unexpected(ctx context.Context, err error) error {
	if errors.Is(err, ErrUnexpected) {
		return err
	}
	if errors.Is(err, ctx.Err()) {
		slog.Info("context was cancelled or timed out")
		return err
	}

	slog.Error("unexpected error occurred", slog.String("error", err.Error()))
	return fmt.Errorf("%w: %s", ErrUnexpected, err)
}

// StatusClientClosedConnection is the http response status returned to the
// ingress if the client closed the connection. The client won't see the
// response, but the ingress might use it for analysis.
const StatusClientClosedConnection = 499

// HTTPError maps the provided error to the correct http status code and writes
// that status code to the response writer `w`. It does not end the request; the
// caller should ensure no further writes are done to w.
func HTTPError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {

	// The client closed the connection and cancelled the context.
	case errors.Is(err, context.Canceled):
		http.Error(w, err.Error(), StatusClientClosedConnection) // 499
		slog.Info(
			"request interrupted due to ctx cancellation",
			slog.String("error", err.Error()),
		)

	// The context deadline was exceeded and the connection had a timeout.
	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, err.Error(), http.StatusServiceUnavailable) // 503
		slog.Info(
			"request interrupted due to ctx timeout",
			slog.String("error", err.Error()),
		)

	// The client made a bad request that cannot be fulfilled.
	case errors.Is(err, ErrBadRequest),
		errors.Is(err, ErrSpaceFull):

		http.Error(w, err.Error(), http.StatusBadRequest) // 400
		slog.Info("client made a bad request", slog.String("error", err.Error()))

	// The client requested an action that is not allowed.
	case errors.Is(err, ErrNotAllowed):
		http.Error(w, err.Error(), http.StatusForbidden) // 403
		slog.Info(
			"client requested an action that is not allowed",
			slog.String("error", err.Error()),
		)

	// The client requested a non-existing resource or action.
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound) // 404
		slog.Info(
			"client requested a missing resource or action",
			slog.String("error", err.Error()),
		)

	// The client requested to create a resource that already exists.
	case errors.Is(err, ErrAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict) // 409
		slog.Info(
			"client requested to create a resource that already exists",
			slog.String("error", err.Error()),
		)

	// The service is not correctly configured or another unexpected error
	// occurred. Errors like [ErrTimeout], [ErrUnexpected] and other unhandled
	// errors can end up here.
	case errors.Is(err, ErrUnexpected):
		fallthrough
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		slog.Error(
			"unexpected error while handling request",
			slog.String("error", err.Error()),
		)
	}
}
