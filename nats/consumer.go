package nats

import (
	"sync"

	natsbroker "github.com/nats-io/nats.go"
)

// CommonConsumer is a base consumer for processing NATS messages.
type CommonConsumer struct {
	connection         *natsbroker.Conn
	subscription       *natsbroker.Subscription
	messageChannel     chan *natsbroker.Msg
	goroutinesPoolSize int
	messageHandler     func(message *natsbroker.Msg)
	isRunning          bool
	isStopped          bool
	wg                 *sync.WaitGroup
}

// NewConsumer creates *CommonConsumer with provided options.
func NewConsumer(
	url string,
	subject string,
	opts ...ConsumerOption,
) (*CommonConsumer, error) {
	options := newConsumerOptions()
	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	connection, err := natsbroker.Connect(url, options.natsOpts...)
	if err != nil {
		return nil, err
	}

	connection.SetErrorHandler(options.errorHandler)
	connection.SetDisconnectErrHandler(options.disconnectErrorHandler)
	connection.SetClosedHandler(options.closeHandler)

	messageChannel := make(chan *natsbroker.Msg, options.messageChannelBufferSize)

	subscription, err := connection.ChanSubscribe(subject, messageChannel)
	if err != nil {
		return nil, err
	}

	return &CommonConsumer{
		connection:         connection,
		subscription:       subscription,
		messageChannel:     messageChannel,
		messageHandler:     options.messageHandler,
		goroutinesPoolSize: options.goroutinesPoolSize,
		wg:                 new(sync.WaitGroup),
	}, nil
}

// Run starts goroutines for NATS messages processing.
func (c *CommonConsumer) Run() error {
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

// Stop stops launched goroutines, which processes NATS messages.
func (c *CommonConsumer) Stop() error {
	if c.isStopped {
		return &ConsumerAlreadyStoppedError{}
	}

	if err := c.subscription.Unsubscribe(); err != nil {
		return err
	}

	close(c.messageChannel)
	c.wg.Wait()

	c.connection.Close()
	c.isStopped = true

	return nil
}
