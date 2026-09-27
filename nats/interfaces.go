package nats

// Consumer asynchronously processes NATS messages in goroutines.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/consumer.go -package=mocks -exclude_interfaces=Publisher,IDPublisher
type Consumer interface {
	Run() error
	Stop() error
}

// Publisher publishes messages to NATS broker.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/publisher.go -package=mocks -exclude_interfaces=Consumer,IDPublisher
type Publisher interface {
	Publish(subject string, content []byte) error
	Close() error
}

// IDPublisher publishes messages with a deduplication id.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/id_publisher.go -package=mocks -exclude_interfaces=Consumer,Publisher
type IDPublisher interface {
	Publisher

	// PublishWithID sends a message with a deduplication id. A repeated publish
	// with the same id inside the server deduplication window is dropped.
	PublishWithID(subject, msgID string, content []byte) error
}
