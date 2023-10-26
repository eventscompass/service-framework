package service

import (
	"context"
)

// EventHandler is a callback function, which is executed when a subscriber
// receives a message.
type EventHandler func(ctx context.Context, payload any) error

// Payload is an interface that all message payloads must implement.
type Payload interface {

	// Topic returns the topic of the payload.
	Topic() string
}

// MessageBus defines the interface for publishing messages to a topic and
// subscribing for receiving messages from a topic.
type MessageBus interface {

	// Publish publishes a message to a given topic.
	// The payload must support `json.Unmarshal`.
	Publish(_ context.Context, p Payload) error

	// Subscribe subscribes to the given topic. The event handler
	// callback will be executed on every received message.
	// This is a blocking function.
	Subscribe(_ context.Context, topic string, h EventHandler) error
}
