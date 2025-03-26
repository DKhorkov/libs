//go:build integration

package nats

import (
	"testing"
	"time"

	natsbroker "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	url                      = natsbroker.DefaultURL
	subject                  = "test"
	workerName               = "nats-worker"
	goroutinesPoolSize       = 1
	messageChannelBufferSize = 1
)

func TestWorker_Run(t *testing.T) {
	t.Run("worker is already running", func(t *testing.T) {
		var resultStorage []string
		worker, err := NewWorker(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(workerName)),
			WithGoroutinesPoolSize(goroutinesPoolSize),
			WithMessageChannelBufferSize(messageChannelBufferSize),
			WithCloseHandler(func(_ *natsbroker.Conn) {
				t.Log("close handler called")
			}),
			WithErrorHandler(func(_ *natsbroker.Conn, _ *natsbroker.Subscription, err error) {
				t.Logf("error handler called. Error: %v", err)
			}),
			WithMessageHandler(func(m *natsbroker.Msg) {
				data := string(m.Data)
				t.Logf("message handler called. Message: %s", data)
				resultStorage = append(resultStorage, data)
			}),
			WithDisconnectErrorHandler(func(_ *natsbroker.Conn, err error) {
				t.Logf("disconnect handler called. Error: %v", err)
			}),
		)

		if err != nil {
			t.Fatal(err)
		}

		worker.isRunning = true
		err = worker.Run()
		require.Error(t, err)
		assert.IsType(t, &WorkerAlreadyRunningError{}, err)
		require.Empty(t, resultStorage)
	})

	t.Run("worker successfully started", func(t *testing.T) {
		var resultStorage []string
		worker, err := NewWorker(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(workerName)),
			WithGoroutinesPoolSize(goroutinesPoolSize),
			WithMessageChannelBufferSize(messageChannelBufferSize),
			WithCloseHandler(func(_ *natsbroker.Conn) {
				t.Log("close handler called")
			}),
			WithErrorHandler(func(_ *natsbroker.Conn, _ *natsbroker.Subscription, err error) {
				t.Logf("error handler called. Error: %v", err)
			}),
			WithMessageHandler(func(m *natsbroker.Msg) {
				data := string(m.Data)
				t.Logf("message handler called. Message: %s", data)
				resultStorage = append(resultStorage, data)
			}),
			WithDisconnectErrorHandler(func(_ *natsbroker.Conn, err error) {
				t.Logf("disconnect handler called. Error: %v", err)
			}),
		)

		if err != nil {
			t.Fatal(err)
		}

		publisher, err := NewPublisher(url)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			require.NoError(t, publisher.Close())
		}()

		message := "test"
		expected := []string{message}
		err = publisher.Publish(subject, []byte(message))
		if err != nil {
			t.Fatal(err)
		}

		err = worker.Run()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		require.Equal(t, expected, resultStorage)

		err = worker.Stop()
		require.NoError(t, err)
	})
}

func TestWorker_Stop(t *testing.T) {
	t.Run("worker is already stopped", func(t *testing.T) {
		var resultStorage []string
		worker, err := NewWorker(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(workerName)),
			WithGoroutinesPoolSize(goroutinesPoolSize),
			WithMessageChannelBufferSize(messageChannelBufferSize),
			WithCloseHandler(func(_ *natsbroker.Conn) {
				t.Log("close handler called")
			}),
			WithErrorHandler(func(_ *natsbroker.Conn, _ *natsbroker.Subscription, err error) {
				t.Logf("error handler called. Error: %v", err)
			}),
			WithMessageHandler(func(m *natsbroker.Msg) {
				data := string(m.Data)
				t.Logf("message handler called. Message: %s", data)
				resultStorage = append(resultStorage, data)
			}),
			WithDisconnectErrorHandler(func(_ *natsbroker.Conn, err error) {
				t.Logf("disconnect handler called. Error: %v", err)
			}),
		)

		if err != nil {
			t.Fatal(err)
		}

		worker.isStopped = true
		err = worker.Stop()
		require.Error(t, err)
		assert.IsType(t, &WorkerAlreadyStoppedError{}, err)
		require.Empty(t, resultStorage)
	})

	t.Run("worker successfully stopped", func(t *testing.T) {
		var resultStorage []string
		worker, err := NewWorker(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(workerName)),
			WithGoroutinesPoolSize(1),
			WithMessageChannelBufferSize(1),
			WithCloseHandler(func(_ *natsbroker.Conn) {
				t.Log("close handler called")
			}),
			WithErrorHandler(func(_ *natsbroker.Conn, _ *natsbroker.Subscription, err error) {
				t.Logf("error handler called. Error: %v", err)
			}),
			WithMessageHandler(func(m *natsbroker.Msg) {
				data := string(m.Data)
				t.Logf("message handler called. Message: %s", data)
				resultStorage = append(resultStorage, data)
			}),
			WithDisconnectErrorHandler(func(_ *natsbroker.Conn, err error) {
				t.Logf("disconnect handler called. Error: %v", err)
			}),
		)

		if err != nil {
			t.Fatal(err)
		}

		err = worker.Stop()
		require.NoError(t, err)
	})
}
