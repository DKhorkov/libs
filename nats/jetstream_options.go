package nats

import (
	"time"

	natsbroker "github.com/nats-io/nats.go"
)

const (
	defaultJetStreamAckWait       = 30 * time.Second
	defaultJetStreamMaxDeliver    = 5
	defaultJetStreamMaxAckPending = 64
)

// newJetStreamConsumerOptions creates *jetStreamConsumerOptions with defaults.
func newJetStreamConsumerOptions() *jetStreamConsumerOptions {
	return &jetStreamConsumerOptions{
		messageChannelBufferSize: defaultMessageChannelBufferSize,
		goroutinesPoolSize:       defaultGoroutinesPoolSize,
		messageHandler:           defaultMessageHandler,
		errorHandler:             defaultErrorHandler,
		disconnectErrorHandler:   defaultDisconnectErrorHandler,
		closeHandler:             defaultCloseHandler,
		ackWait:                  defaultJetStreamAckWait,
		maxDeliver:               defaultJetStreamMaxDeliver,
		maxAckPending:            defaultJetStreamMaxAckPending,
	}
}

// jetStreamConsumerOptions represents options for JetStreamConsumer.
type jetStreamConsumerOptions struct {
	stream                   string
	durable                  string
	ackWait                  time.Duration
	maxDeliver               int
	maxAckPending            int
	messageChannelBufferSize int
	goroutinesPoolSize       int
	messageHandler           func(message *natsbroker.Msg)
	errorHandler             func(connection *natsbroker.Conn, subscription *natsbroker.Subscription, err error)
	disconnectErrorHandler   func(connection *natsbroker.Conn, err error)
	closeHandler             func(connection *natsbroker.Conn)
	natsOpts                 []natsbroker.Option
}

// JetStreamConsumerOption represents functional option for JetStreamConsumer.
type JetStreamConsumerOption func(options *jetStreamConsumerOptions) error

// WithJetStreamStream binds consumer to an existing stream.
func WithJetStreamStream(name string) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.stream = name

		return nil
	}
}

// WithJetStreamDurable sets durable consumer name, which holds reading position
// on the server.
func WithJetStreamDurable(name string) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.durable = name

		return nil
	}
}

// WithJetStreamAckWait sets how long server waits for an acknowledgement before
// redelivering the message.
func WithJetStreamAckWait(ackWait time.Duration) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.ackWait = ackWait

		return nil
	}
}

// WithJetStreamMaxDeliver limits number of delivery attempts.
func WithJetStreamMaxDeliver(maxDeliver int) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.maxDeliver = maxDeliver

		return nil
	}
}

// WithJetStreamMaxAckPending limits number of unacknowledged messages in flight.
func WithJetStreamMaxAckPending(maxAckPending int) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.maxAckPending = maxAckPending

		return nil
	}
}

// WithJetStreamMessageChannelBufferSize sets buffer of the messages channel.
func WithJetStreamMessageChannelBufferSize(size int) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.messageChannelBufferSize = size

		return nil
	}
}

// WithJetStreamGoroutinesPoolSize sets number of processing goroutines.
func WithJetStreamGoroutinesPoolSize(size int) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.goroutinesPoolSize = size

		return nil
	}
}

// WithJetStreamMessageHandler sets handler for received message.
func WithJetStreamMessageHandler(handler func(message *natsbroker.Msg)) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.messageHandler = handler

		return nil
	}
}

// WithJetStreamNatsOptions sets NATS connection options.
func WithJetStreamNatsOptions(opts ...natsbroker.Option) JetStreamConsumerOption {
	return func(options *jetStreamConsumerOptions) error {
		options.natsOpts = append(options.natsOpts, opts...)

		return nil
	}
}
