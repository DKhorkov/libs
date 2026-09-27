package nats

import (
	"errors"
	"fmt"
	"sync"
	"time"

	natsbroker "github.com/nats-io/nats.go"
)

const (
	// drainTimeout is an upper bound of waiting for an asynchronous drain.
	drainTimeout = 30 * time.Second

	// drainPollInterval is how often a drain completion is checked.
	drainPollInterval = 10 * time.Millisecond
)

// StreamConfig describes a JetStream stream with file storage.
//
// Retention is always WorkQueuePolicy: a message leaves the stream when it is
// acknowledged, not when it expires. MaxAge and MaxBytes are the upper bounds,
// which keep an undeliverable task from growing the stream forever.
//
// Discard is always DiscardNew, and it is not a detail. With the default
// DiscardOld a stream, which hit MaxBytes, silently drops the OLDEST
// unacknowledged messages — that is, loses the very tasks the stream exists to
// keep. DiscardNew fails the publish instead, loudly and at the publisher.
//
// Mind the WorkQueuePolicy constraint on the consumer side: filter subjects of
// consumers of such a stream must not overlap. One consumer per channel
// (`notifications.*.email`, `.web_push`, `.bell`) is fine; a second one on a
// subset of those subjects is refused with `filtered consumer not unique on
// workqueue stream`.
type StreamConfig struct {
	// Name is a stream name. Stream is created on publisher start, if missing.
	Name string

	// Subjects is a list of subject patterns, captured by the stream.
	Subjects []string

	// MaxAge is an upper bound of a message lifetime in the stream.
	MaxAge time.Duration

	// MaxBytes is an upper bound of a stream size on disk. Reaching it fails
	// a publish (see DiscardNew above), and does not drop a stored task.
	MaxBytes int64

	// Duplicates is a deduplication window for Nats-Msg-Id header.
	Duplicates time.Duration
}

// JetStreamPublisher publishes messages to a JetStream stream.
//
// Unlike CommonPublisher it waits for a server acknowledgement: a publish,
// which returned nil, is stored on disk.
type JetStreamPublisher struct {
	connection *natsbroker.Conn
	jetStream  natsbroker.JetStreamContext
}

// NewJetStreamPublisher creates *JetStreamPublisher and ensures the stream.
func NewJetStreamPublisher(
	url string,
	stream StreamConfig,
	opts ...natsbroker.Option,
) (*JetStreamPublisher, error) {
	if stream.Name == "" || len(stream.Subjects) == 0 {
		return nil, &StreamNotConfiguredError{}
	}

	connection, err := natsbroker.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	jetStream, err := connection.JetStream()
	if err != nil {
		connection.Close()

		return nil, err
	}

	if err = ensureStream(jetStream, stream); err != nil {
		connection.Close()

		return nil, err
	}

	return &JetStreamPublisher{
		connection: connection,
		jetStream:  jetStream,
	}, nil
}

// ensureStream creates the stream, if it does not exist, and updates subjects
// and limits otherwise.
func ensureStream(jetStream natsbroker.JetStreamContext, stream StreamConfig) error {
	config := &natsbroker.StreamConfig{
		Name:      stream.Name,
		Subjects:  stream.Subjects,
		Retention: natsbroker.WorkQueuePolicy,
		Storage:   natsbroker.FileStorage,
		// Never drop a stored task to make room for a new one: a full stream
		// must break the publisher, not swallow the queue.
		Discard:    natsbroker.DiscardNew,
		MaxAge:     stream.MaxAge,
		MaxBytes:   stream.MaxBytes,
		Duplicates: stream.Duplicates,
	}

	_, err := jetStream.StreamInfo(stream.Name)

	switch {
	case errors.Is(err, natsbroker.ErrStreamNotFound):
		_, err = jetStream.AddStream(config)

		return err
	case err != nil:
		return err
	}

	_, err = jetStream.UpdateStream(config)

	return err
}

// Publish sends a message without deduplication.
//
// It matches Publisher by signature, but it is NOT a drop-in replacement for
// the core one, and assuming otherwise is the trap this comment exists for: a
// JetStream publish waits for a stream acknowledgement, so a subject, which no
// stream captures, fails with `no responders`. Live WebSocket events therefore
// keep going through the core publisher.
func (p *JetStreamPublisher) Publish(subject string, data []byte) error {
	_, err := p.jetStream.Publish(subject, data)

	return err
}

// PublishWithID sends a message with Nats-Msg-Id header.
//
// A repeated publish with the same id inside the deduplication window is
// dropped by the server. It closes the case, when a publisher did not get an
// acknowledgement and sent the task for the second time.
func (p *JetStreamPublisher) PublishWithID(subject, msgID string, data []byte) error {
	_, err := p.jetStream.Publish(subject, data, natsbroker.MsgId(msgID))

	return err
}

// Close closes NATS connection.
func (p *JetStreamPublisher) Close() error {
	p.connection.Close()

	return nil
}

// JetStreamConsumer processes messages of a durable JetStream consumer.
//
// Reading position lives on the server, so a restarted worker continues from
// the place, where it stopped, and not from the current moment. Acknowledgement
// is manual: the handler acks after a successful delivery, and an unacked
// message is redelivered after AckWait.
type JetStreamConsumer struct {
	connection         *natsbroker.Conn
	subscription       *natsbroker.Subscription
	messageChannel     chan *natsbroker.Msg
	goroutinesPoolSize int
	messageHandler     func(message *natsbroker.Msg)
	isRunning          bool
	isStopped          bool
	wg                 *sync.WaitGroup
}

// NewJetStreamConsumer creates *JetStreamConsumer with provided options.
func NewJetStreamConsumer(
	url string,
	subject string,
	opts ...JetStreamConsumerOption,
) (*JetStreamConsumer, error) {
	options := newJetStreamConsumerOptions()

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	if options.stream == "" || options.durable == "" {
		return nil, &StreamNotConfiguredError{
			Message: "jetstream consumer requires both stream and durable names",
		}
	}

	connection, err := natsbroker.Connect(url, options.natsOpts...)
	if err != nil {
		return nil, err
	}

	connection.SetErrorHandler(options.errorHandler)
	connection.SetDisconnectErrHandler(options.disconnectErrorHandler)
	connection.SetClosedHandler(options.closeHandler)

	jetStream, err := connection.JetStream()
	if err != nil {
		connection.Close()

		return nil, err
	}

	if err = ensureConsumer(jetStream, subject, options); err != nil {
		connection.Close()

		return nil, err
	}

	messageChannel := make(chan *natsbroker.Msg, options.messageChannelBufferSize)

	// Queue subscription by the durable name: a task must reach exactly one
	// replica, and the queue name must match the DeliverGroup of the consumer.
	//
	// Bind, and nothing else: the consumer already exists with its ack policy,
	// ack_wait and limits, and passing them here again is both pointless and
	// rejected by the client. Manual ack is the whole point of the stream, so
	// it is not an option, but a constant.
	subscription, err := jetStream.ChanQueueSubscribe(
		subject,
		options.durable,
		messageChannel,
		natsbroker.Bind(options.stream, options.durable),
		natsbroker.ManualAck(),
	)
	if err != nil {
		connection.Close()

		return nil, err
	}

	return &JetStreamConsumer{
		connection:         connection,
		subscription:       subscription,
		messageChannel:     messageChannel,
		messageHandler:     options.messageHandler,
		goroutinesPoolSize: options.goroutinesPoolSize,
		wg:                 new(sync.WaitGroup),
	}, nil
}

// ensureConsumer creates a durable push consumer, or updates an existing one.
//
// The consumer is created APART from the subscription, and this is the whole
// point of the function. A durable, created by the subscribe call, is deleted
// by both Unsubscribe and Drain: the reading position goes away with it, and
// with several replicas a shutdown of one breaks the others. A consumer,
// created here and only bound by the subscription, survives both.
//
// An existing consumer is UPDATED, not left alone, and that is the second
// point. A durable outlives the worker by design, so skipping the update would
// mean changed AckWait, MaxDeliver or MaxAckPending never reach the server: you
// edit the settings, restart the application and silently keep the old values —
// no error, no trace. ensureStream updates the stream for the same reason, and
// the asymmetry would read as an oversight.
//
// DeliverPolicy is left out of the update on purpose: it is immutable after
// creation, and sending it back unchanged is both pointless and rejected by
// some server versions.
func ensureConsumer(
	jetStream natsbroker.JetStreamContext,
	subject string,
	options *jetStreamConsumerOptions,
) error {
	config := &natsbroker.ConsumerConfig{
		Durable: options.durable,
		// DeliverSubject makes it a push consumer, DeliverGroup — a queue one:
		// a task must reach exactly one replica. The subject is derived from
		// the names instead of being a random inbox, so that a consumer, seen
		// in monitoring, says which worker it belongs to.
		DeliverSubject: fmt.Sprintf("_deliver.%s.%s", options.stream, options.durable),
		DeliverGroup:   options.durable,
		FilterSubject:  subject,
		AckPolicy:      natsbroker.AckExplicitPolicy,
		AckWait:        options.ackWait,
		MaxDeliver:     options.maxDeliver,
		MaxAckPending:  options.maxAckPending,
	}

	_, err := jetStream.ConsumerInfo(options.stream, options.durable)

	switch {
	case err == nil:
		_, err = jetStream.UpdateConsumer(options.stream, config)

		return err
	case !errors.Is(err, natsbroker.ErrConsumerNotFound):
		return err
	}

	// DeliverAllPolicy only on creation: a durable, created once, keeps its
	// reading position, and a restarted worker must continue from it rather
	// than replay the stream.
	config.DeliverPolicy = natsbroker.DeliverAllPolicy

	_, err = jetStream.AddConsumer(options.stream, config)

	return err
}

// Run starts goroutines for messages processing.
func (c *JetStreamConsumer) Run() error {
	if c.isRunning {
		return &ConsumerAlreadyRunningError{}
	}

	c.wg.Add(c.goroutinesPoolSize)

	for range c.goroutinesPoolSize {
		go func() {
			defer c.wg.Done()

			for msg := range c.messageChannel {
				c.messageHandler(msg)
			}
		}()
	}

	c.isRunning = true

	return nil
}

// Stop drains the subscription and waits for launched goroutines.
//
// Drain instead of Unsubscribe — but neither of them is what keeps the durable
// consumer alive: both delete a consumer, created by the subscribe call. The
// consumer survives, because ensureConsumer creates it apart and the
// subscription only binds to it.
//
// Drain is asynchronous: it returns at once, and the library keeps delivering
// buffered messages into the channel. Closing the channel right away would
// panic with a send on a closed channel, so the drain is awaited first. The
// timeout insures against a hung connection and loses nothing: an unacked task
// comes back after ack_wait.
func (c *JetStreamConsumer) Stop() error {
	if c.isStopped {
		return &ConsumerAlreadyStoppedError{}
	}

	if err := c.subscription.Drain(); err != nil {
		return err
	}

	deadline := time.Now().Add(drainTimeout)
	for c.subscription.IsValid() && time.Now().Before(deadline) {
		time.Sleep(drainPollInterval)
	}

	close(c.messageChannel)
	c.wg.Wait()

	c.connection.Close()
	c.isStopped = true

	return nil
}
