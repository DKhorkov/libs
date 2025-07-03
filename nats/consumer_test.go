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
	consumerName             = "nats-consumer"
	goroutinesPoolSize       = 1
	messageChannelBufferSize = 1
)

func TestConsumer_Run(t *testing.T) {
	t.Run("consumer is already running", func(t *testing.T) {
		var resultStorage []string
		consumer, err := NewConsumer(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(consumerName)),
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

		consumer.isRunning = true
		err = consumer.Run()
		require.Error(t, err)
		assert.IsType(t, &ConsumerAlreadyRunningError{}, err)
		require.Empty(t, resultStorage)
	})

	t.Run("consumer successfully started", func(t *testing.T) {
		var resultStorage []string
		consumer, err := NewConsumer(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(consumerName)),
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

		err = consumer.Run()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		require.Equal(t, expected, resultStorage)

		err = consumer.Stop()
		require.NoError(t, err)
	})

	t.Run("consumer without options successfully started", func(t *testing.T) {
		var resultStorage []string
		consumer, err := NewConsumer(
			url,
			subject,
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
		err = publisher.Publish(subject, []byte(message))
		if err != nil {
			t.Fatal(err)
		}

		err = consumer.Run()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		require.Empty(t, resultStorage)

		err = consumer.Stop()
		require.NoError(t, err)
	})
}

func TestConsumer_Stop(t *testing.T) {
	t.Run("consumer is already stopped", func(t *testing.T) {
		var resultStorage []string
		consumer, err := NewConsumer(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(consumerName)),
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

		consumer.isStopped = true
		err = consumer.Stop()
		require.Error(t, err)
		assert.IsType(t, &ConsumerAlreadyStoppedError{}, err)
		require.Empty(t, resultStorage)
	})

	t.Run("consumer successfully stopped", func(t *testing.T) {
		var resultStorage []string
		consumer, err := NewConsumer(
			url,
			subject,
			WithNatsOptions(natsbroker.Name(consumerName)),
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

		err = consumer.Stop()
		require.NoError(t, err)
	})

	t.Run("consumer without options successfully stopped", func(t *testing.T) {
		consumer, err := NewConsumer(
			url,
			subject,
		)

		if err != nil {
			t.Fatal(err)
		}

		err = consumer.Stop()
		require.NoError(t, err)
	})
}
