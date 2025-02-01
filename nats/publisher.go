package nats

import "github.com/nats-io/nats.go"

// NewCommonPublisher creates *CommonPublisher.
func NewCommonPublisher(url string, opts ...nats.Option) (*CommonPublisher, error) {
	connection, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	return &CommonPublisher{
		connection: connection,
	}, nil
}

// CommonPublisher is a base NATS publisher.
type CommonPublisher struct {
	connection *nats.Conn
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
