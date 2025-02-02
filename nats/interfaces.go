package nats

// Worker asynchronously processes NATS messages in goroutines.
type Worker interface {
	Run() error
	Stop() error
}

// Publisher publishes messages to NATS broker.
type Publisher interface {
	Publish(subject string, content []byte) error
	Close() error
}
