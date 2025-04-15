package nats

import (
	"sync"

	natsbroker "github.com/nats-io/nats.go"
)

// CommonWorker is a base worker for processing NATS messages.
type CommonWorker struct {
	connection         *natsbroker.Conn
	subscription       *natsbroker.Subscription
	messageChannel     chan *natsbroker.Msg
	goroutinesPoolSize int
	messageHandler     func(message *natsbroker.Msg)
	isRunning          bool
	isStopped          bool
	wg                 *sync.WaitGroup
}

// NewWorker creates *CommonWorker with provided options.
func NewWorker(
	url string,
	subject string,
	opts ...WorkerOption,
) (*CommonWorker, error) {
	options := newWorkerOptions()
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

	return &CommonWorker{
		connection:         connection,
		subscription:       subscription,
		messageChannel:     messageChannel,
		messageHandler:     options.messageHandler,
		goroutinesPoolSize: options.goroutinesPoolSize,
		wg:                 new(sync.WaitGroup),
	}, nil
}

// Run starts goroutines for NATS messages processing.
func (w *CommonWorker) Run() error {
	if w.isRunning {
		return &WorkerAlreadyRunningError{}
	}

	w.wg.Add(w.goroutinesPoolSize)

	for range w.goroutinesPoolSize {
		go func() {
			defer w.wg.Done()

			for msg := range w.messageChannel {
				w.messageHandler(msg)
			}
		}()
	}

	w.isRunning = true

	return nil
}

// Stop stops launched goroutines, which processes NATS messages.
func (w *CommonWorker) Stop() error {
	if w.isStopped {
		return &WorkerAlreadyStoppedError{}
	}

	if err := w.subscription.Unsubscribe(); err != nil {
		return err
	}

	close(w.messageChannel)
	w.wg.Wait()

	w.connection.Close()
	w.isStopped = true

	return nil
}
