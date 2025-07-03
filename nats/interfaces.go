package nats

// Consumer asynchronously processes NATS messages in goroutines.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/consumer.go -package=mocks -exclude_interfaces=Publisher
type Consumer interface {
	Run() error
	Stop() error
}

// Publisher publishes messages to NATS broker.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/publisher.go -package=mocks -exclude_interfaces=Consumer
type Publisher interface {
	Publish(subject string, content []byte) error
	Close() error
}
