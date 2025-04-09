package nats

// Worker asynchronously processes NATS messages in goroutines.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/worker.go -package=mocks -exclude_interfaces=Publisher
type Worker interface {
	Run() error
	Stop() error
}

// Publisher publishes messages to NATS broker.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/publisher.go -package=mocks -exclude_interfaces=Worker
type Publisher interface {
	Publish(subject string, content []byte) error
	Close() error
}
