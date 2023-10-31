package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/eventscompass/service-framework/service"
)

// AMQPBus is a message bus backed by RabbitMQ message broker.
type AMQPBus struct {
	// conn is the connection to the RabbitMQ message broker.
	conn *amqp.Connection

	// exchange is the exchange associated with this Bus.
	exchange string
}

var (
	_ service.MessageBus = (*AMQPBus)(nil)
	_ io.Closer          = (*AMQPBus)(nil)
)

// NewAMQPBus creates a new [AMQPBus] instance.
func NewAMQPBus(cfg *service.BusConfig, exchange string) (*AMQPBus, error) {
	connInfo := fmt.Sprintf(
		"amqp://%s:%s@%s:%d", cfg.Username, cfg.Password, cfg.Host, cfg.Port)

	// Use once.Do to make sure that a given micro-service creates only one
	// rabbitmq connection even if it calls this function multiple times.
	var err error
	once.Do(func() { conn, err = amqp.Dial(connInfo) })
	if err != nil {
		// TODO: maybe try using exponential backoff for connecting?
		return nil, fmt.Errorf("%w: amqp dial: %v", service.ErrUnexpected, err)
	}

	// Make sure the connection is working by opening a channel on it.
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w: pubsub channel: %v", service.ErrUnexpected, err)
	}
	defer ch.Close()

	return &AMQPBus{
		conn:     conn,
		exchange: exchange,
	}, nil
}

// Publish implements the [service.MessageBus] interface.
func (b *AMQPBus) Publish(ctx context.Context, p service.Payload) error {
	if b.conn.IsClosed() {
		return service.ErrConnectionClosed
	}

	// TODO: Maybe marshal before publishing ?
	// Publish(ctx context.Context, topic string, payload []byte) error
	topic := p.Topic()
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("%w: marshal payload: %v", service.ErrUnexpected, err)
	}

	// Note that AMQP channels are not thread-safe. Thus, we will be creating a
	// new channel for every published message. By using separate AMQP channels
	// we can reuse the same AMQP connection.
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("%w: pubsub channel: %v", service.ErrUnexpected, err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(b.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: exchange declare: %v", service.ErrUnexpected, err)
	}

	err = ch.PublishWithContext(
		ctx,
		b.exchange, // exchange
		topic,      // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		// TODO: maybe we should retry publishing.
		// https://cloud.google.com/pubsub/docs/samples/pubsub-publish-with-error-handler
		return service.Unexpected(ctx, fmt.Errorf("publish message: %w", err))
	}
	return nil
}

// Subscribe implements the [service.MessageBus] interface.
func (b *AMQPBus) Subscribe(
	ctx context.Context,
	topic string,
	eventHandler service.EventHandler,
) error {
	if b.conn.IsClosed() {
		return service.ErrConnectionClosed
	}

	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("%w: pubsub channel: %v", service.ErrUnexpected, err)
	}
	defer ch.Close()

	// Before binding the queue, make sure the exchange exists.
	err = ch.ExchangeDeclare(b.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: exchange declare: %v", service.ErrUnexpected, err)
	}

	// Declare a queue with an arbitrary name and bind it to the exchange.
	q, err := ch.QueueDeclare("", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%w: queue declare: %v", service.ErrUnexpected, err)
	}
	defer ch.QueueDelete(q.Name, false, false, true)

	err = ch.QueueBind(q.Name, topic, b.exchange, false, nil)
	if err != nil {
		return fmt.Errorf("%w: queue bind: %v", service.ErrUnexpected, err)
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
		return fmt.Errorf("%w: queue consume: %v", service.ErrUnexpected, err)
	}

	for msg := range msgs {
		// Pass the message to the event handler.
		eventHandler(ctx, msg.Body)

		// Ack the message only after we have successfully finished processing.
		_ = msg.Ack(false)
	}

	return nil
}

// Close implements the [io.Closer] interface.
func (b *AMQPBus) Close() error {
	return b.conn.Close()
}

var (
	// Use a singleton to make sure only one connection is open.
	once sync.Once
	conn *amqp.Connection
)
