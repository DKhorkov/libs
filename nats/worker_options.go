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

// newWorkerOptions creates *workerOptions with default values.
func newWorkerOptions() *workerOptions {
	return &workerOptions{
		messageChannelBufferSize: defaultMessageChannelBufferSize,
		goroutinesPoolSize:       defaultGoroutinesPoolSize,
		messageHandler:           defaultMessageHandler,
		errorHandler:             defaultErrorHandler,
		disconnectErrorHandler:   defaultDisconnectErrorHandler,
		closeHandler:             defaultCloseHandler,
	}
}

// workerOptions represents options for Worker configuration.
type workerOptions struct {
	messageChannelBufferSize int
	goroutinesPoolSize       int
	messageHandler           func(message *natsbroker.Msg)
	errorHandler             func(connection *natsbroker.Conn, subscription *natsbroker.Subscription, err error)
	disconnectErrorHandler   func(connection *natsbroker.Conn, err error)
	closeHandler             func(connection *natsbroker.Conn)
	natsOpts                 []natsbroker.Option
}

// WorkerOption represents golang functional option pattern func for Worker configuration.
type WorkerOption func(options *workerOptions) error

// WithMessageChannelBufferSize sets buffer for channel, where NATS will store messages for processing.
func WithMessageChannelBufferSize(size int) WorkerOption {
	return func(options *workerOptions) error {
		options.messageChannelBufferSize = size
		return nil
	}
}

// WithGoroutinesPoolSize sets number of goroutines for process messages from NATS via message channel.
func WithGoroutinesPoolSize(size int) WorkerOption {
	return func(options *workerOptions) error {
		options.goroutinesPoolSize = size
		return nil
	}
}

// WithMessageHandler sets handler for received message.
func WithMessageHandler(handler func(message *natsbroker.Msg)) WorkerOption {
	return func(options *workerOptions) error {
		options.messageHandler = handler
		return nil
	}
}

// WithErrorHandler sets handler for processing error during message processing.
func WithErrorHandler(
	handler func(connection *natsbroker.Conn, subscription *natsbroker.Subscription, err error),
) WorkerOption {
	return func(options *workerOptions) error {
		options.errorHandler = handler
		return nil
	}
}

// WithDisconnectErrorHandler sets handler for disconnection from server.
func WithDisconnectErrorHandler(handler func(connection *natsbroker.Conn, err error)) WorkerOption {
	return func(options *workerOptions) error {
		options.disconnectErrorHandler = handler
		return nil
	}
}

// WithCloseHandler sets handler for connection with NATS closure.
func WithCloseHandler(handler func(connection *natsbroker.Conn)) WorkerOption {
	return func(options *workerOptions) error {
		options.closeHandler = handler
		return nil
	}
}

// WithNatsOptions sets NATS option for connection with broker configuration.
func WithNatsOptions(opts ...natsbroker.Option) WorkerOption {
	return func(options *workerOptions) error {
		options.natsOpts = append(options.natsOpts, opts...)
		return nil
	}
}
