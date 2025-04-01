package nats

import natsbroker "github.com/nats-io/nats.go"

// NewPublisher creates *CommonPublisher.
func NewPublisher(url string, opts ...natsbroker.Option) (*CommonPublisher, error) {
	connection, err := natsbroker.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	return &CommonPublisher{
		connection: connection,
	}, nil
}

// CommonPublisher is a base NATS publisher.
type CommonPublisher struct {
	connection *natsbroker.Conn
}

// Publish sends message to provided topic (subject).
func (p *CommonPublisher) Publish(topic string, data []byte) error {
	return p.connection.Publish(topic, data)
}

// Close closes NATS connection.
func (p *CommonPublisher) Close() error {
	p.connection.Close()

	return nil
}
