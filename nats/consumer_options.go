package nats

import (
	"fmt"

	natsbroker "github.com/nats-io/nats.go"
)

const (
	defaultMessageChannelBufferSize = 1
	defaultGoroutinesPoolSize       = 1
)

var (
	defaultMessageHandler = func(message *natsbroker.Msg) {
		fmt.Printf("nats message: %s\n", string(message.Data))
	}

	defaultErrorHandler = func(_ *natsbroker.Conn, _ *natsbroker.Subscription, err error) {
		fmt.Printf("nats error: %v\n", err)
	}

	defaultDisconnectErrorHandler = func(_ *natsbroker.Conn, err error) {
		if err != nil {
			fmt.Printf("nats disconnect error: %v\n", err)
		}
	}

	defaultCloseHandler = func(connection *natsbroker.Conn) {
		fmt.Printf("nats close connection. Status: %d\n", connection.Status())
	}
)

// newConsumerOptions creates *consumerOptions with default values.
func newConsumerOptions() *consumerOptions {
	return &consumerOptions{
		messageChannelBufferSize: defaultMessageChannelBufferSize,
		goroutinesPoolSize:       defaultGoroutinesPoolSize,
		messageHandler:           defaultMessageHandler,
		errorHandler:             defaultErrorHandler,
		disconnectErrorHandler:   defaultDisconnectErrorHandler,
		closeHandler:             defaultCloseHandler,
	}
}

// consumerOptions represents options for Consumer configuration.
type consumerOptions struct {
	messageChannelBufferSize int
	goroutinesPoolSize       int
	messageHandler           func(message *natsbroker.Msg)
	errorHandler             func(connection *natsbroker.Conn, subscription *natsbroker.Subscription, err error)
	disconnectErrorHandler   func(connection *natsbroker.Conn, err error)
	closeHandler             func(connection *natsbroker.Conn)
	natsOpts                 []natsbroker.Option
}

// ConsumerOption represents golang functional option pattern func for Consumer configuration.
type ConsumerOption func(options *consumerOptions) error

// WithMessageChannelBufferSize sets buffer for channel, where NATS will store messages for processing.
func WithMessageChannelBufferSize(size int) ConsumerOption {
	return func(options *consumerOptions) error {
		options.messageChannelBufferSize = size

		return nil
	}
}

// WithGoroutinesPoolSize sets number of goroutines for process messages from NATS via message channel.
func WithGoroutinesPoolSize(size int) ConsumerOption {
	return func(options *consumerOptions) error {
		options.goroutinesPoolSize = size

		return nil
	}
}

// WithMessageHandler sets handler for received message.
func WithMessageHandler(handler func(message *natsbroker.Msg)) ConsumerOption {
	return func(options *consumerOptions) error {
		options.messageHandler = handler

		return nil
	}
}

// WithErrorHandler sets handler for processing error during message processing.
func WithErrorHandler(
	handler func(connection *natsbroker.Conn, subscription *natsbroker.Subscription, err error),
) ConsumerOption {
	return func(options *consumerOptions) error {
		options.errorHandler = handler

		return nil
	}
}

// WithDisconnectErrorHandler sets handler for disconnection from server.
func WithDisconnectErrorHandler(handler func(connection *natsbroker.Conn, err error)) ConsumerOption {
	return func(options *consumerOptions) error {
		options.disconnectErrorHandler = handler

		return nil
	}
}

// WithCloseHandler sets handler for connection with NATS closure.
func WithCloseHandler(handler func(connection *natsbroker.Conn)) ConsumerOption {
	return func(options *consumerOptions) error {
		options.closeHandler = handler

		return nil
	}
}

// WithNatsOptions sets NATS option for connection with broker configuration.
func WithNatsOptions(opts ...natsbroker.Option) ConsumerOption {
	return func(options *consumerOptions) error {
		options.natsOpts = append(options.natsOpts, opts...)

		return nil
	}
}
