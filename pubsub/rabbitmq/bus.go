package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/eventscompass/service-framework/service"
)

var (
	// ErrConnFailed is returned when we cannot establish the
	// connection.
	ErrConnFailed = errors.New("connection failed")

	// ErrConnBroken is returned when the connection that we are
	// trying to use is broken.
	ErrConnBroken = errors.New("connection broken")

	// ErrConnClosed is returned when the connection that we are
	// trying to use is closed.
	ErrConnClosed = errors.New("connection closed")

	// ErrChanBroken is returned when the server channel that we
	// are trying to use is broken.
	ErrChanBroken = errors.New("connection channel broken")
)

// Config holds configuration variables for connecting to a RabbitMQ broker.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Bus is a message bus backed by a RabbitMQ message broker.
type Bus struct {
	// conn is the connection to the RabbitMQ message broker.
	conn *amqp.Connection

	// exchange is the exchange associated with this Bus.
	exchange string
}

// NewAMQPBus creates a new [Bus] instance which can be used to publish events
// to the given exchange. If you want to publish to a different exchange then
// simply create a new [Bus] instance, the same broker connection will be
// re-used. This function returns [ErrConnBroken] in case the connection to the
// message broker is broken.
func NewAMQPBus(cfg *Config, exchange string) (*Bus, error) {
	connInfo := fmt.Sprintf(
		"amqp://%s:%s@%s:%d", cfg.Username, cfg.Password, cfg.Host, cfg.Port)

	// Use once.Do to make sure that a given micro-service creates only one
	// rabbitmq connection even if it calls this function multiple times.
	var err error
	once.Do(func() { conn, err = amqp.Dial(connInfo) })
	if err != nil {
		// TODO: maybe try using exponential backoff for connecting?
		return nil, fmt.Errorf("%w: dial broker:  %v", ErrConnFailed, err)
	}

	// Make sure the connection is working by opening a channel on it.
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close() //nolint:errcheck // intentional
		return nil, fmt.Errorf("%w: open channel: %v", ErrConnBroken, err)
	}
	defer ch.Close() //nolint:errcheck // intentional

	return &Bus{
		conn:     conn,
		exchange: exchange,
	}, nil
}

// Publish publishes a message to a given topic. This function returns
// [ErrConnClosed] in case the connection to the message broker is closed.
// This function returns [ErrConnBroken] in case the connection is broken.
// This function returns [ErrChanBroken] in case operations on the connection
// channel fail.
func (b *Bus) Publish(ctx context.Context, topic string, msg []byte) error {
	if b.conn.IsClosed() {
		return ErrConnClosed //nolint:wrapcheck // intentional
	}

	// Note that AMQP channels are not thread-safe. Thus, we will be creating a
	// new channel for every published message. By using separate AMQP channels
	// we can reuse the same AMQP connection concurrently.
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("%w: open channel: %v", ErrConnBroken, err)
	}
	defer ch.Close() //nolint:errcheck // intentional

	err = ch.ExchangeDeclare(b.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: declare exchange: %v", ErrChanBroken, err)
	}

	err = ch.PublishWithContext(
		ctx,
		b.exchange, // exchange
		topic,      // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
	if err != nil {
		// TODO: maybe we should retry publishing.
		// https://cloud.google.com/pubsub/docs/samples/pubsub-publish-with-error-handler
		return fmt.Errorf("%w: publish message: %v", ErrChanBroken, err)
	}
	return nil
}

// Subscribe subscribes to the given topic. The event handler callback will be
// executed on every received message. This function returns [ErrConnClosed] in
// case the connection to the message broker is closed. This function returns
// [ErrConnBroken] in case the connection is broken. This function returns
// [ErrChanBroken] in case operations on the connection channel fail. This is a
// blocking function. Canceling the context will cancel the subscription.
func (b *Bus) Subscribe(
	ctx context.Context,
	topic string,
	eventHandler service.EventHandler,
) error {
	if b.conn.IsClosed() {
		return ErrConnClosed //nolint:wrapcheck // intentional
	}

	// AMQP channels are not thread-safe, thus we need to use a separate channel
	// for every subscription, so that we can reuse the connection concurrently.
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("%w: open channel: %v", ErrConnBroken, err)
	}
	defer ch.Close() //nolint:errcheck // intentional

	// Before binding the queue, make sure the exchange exists.
	err = ch.ExchangeDeclare(b.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: declare exchange: %v", ErrChanBroken, err)
	}

	// Declare a queue with an arbitrary name and bind it to the exchange.
	q, err := ch.QueueDeclare("", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: declare queue: %v", ErrChanBroken, err)
	}
	defer ch.QueueDelete(q.Name, false, false, true) //nolint:errcheck // intentional

	err = ch.QueueBind(q.Name, topic, b.exchange, false, nil)
	if err != nil {
		return fmt.Errorf("%w: bind queue: %v", ErrChanBroken, err)
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // non-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("%w: consume queue: %v", ErrChanBroken, err)
	}

	for msg := range msgs {
		// Pass the message to the event handler.
		eventHandler(ctx, msg.Body)

		// Ack the message only after we have finished processing.
		_ = msg.Ack(false) //nolint:errcheck // intentional
	}

	return nil
}

// Close closes the connection to the message broker and releases all associated
// resources. This function returns [ErrConnBroken] if it fails to close the
// connection.
func (b *Bus) Close() error {
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("%w: close conn: %v", ErrConnBroken, err)
	}
	return nil
}

var (
	// Use a singleton to make sure only one connection is open.
	once sync.Once
	conn *amqp.Connection
)
