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
	streamName    = "TEST_NOTIFICATIONS"
	streamSubject = "test-notifications.>"

	// Субъект на тест, а не один на всех, и это не косметика. Стрим держит
	// WorkQueuePolicy, а на нём фильтры консьюмеров обязаны не пересекаться:
	// второй durable на субъекте первого сервер отвергает с
	// `filtered consumer not unique on workqueue stream`.
	dedupSubject     = "test-notifications.verify_email.email"
	redeliverSubject = "test-notifications.forget_password.email"
	survivorSubject  = "test-notifications.friend_request.bell"

	dedupDurable     = "test-dedup-worker"
	redeliverDurable = "test-redeliver-worker"
	survivorDurable  = "test-survivor-worker"

	duplicatesWindow   = 2 * time.Minute
	streamMaxAge       = time.Hour
	streamMaxBytes     = 1 << 20
	jetStreamAckWait   = 2 * time.Second
	jetStreamMaxDelivs = 3
)

func testStreamConfig() StreamConfig {
	return StreamConfig{
		Name:       streamName,
		Subjects:   []string{streamSubject},
		MaxAge:     streamMaxAge,
		MaxBytes:   streamMaxBytes,
		Duplicates: duplicatesWindow,
	}
}

// TestJetStreamPublisher_PublishWithID проверяет дедупликацию по Nats-Msg-Id:
// публикатор, не дождавшийся подтверждения и отправивший задачу второй раз, не
// обязан порождать второе письмо.
func TestJetStreamPublisher_PublishWithID(t *testing.T) {
	publisher, err := NewJetStreamPublisher(url, testStreamConfig())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, publisher.Close())
	}()

	require.NoError(t, publisher.PublishWithID(dedupSubject, "msg-1", []byte(`{"n":1}`)))
	require.NoError(t, publisher.PublishWithID(dedupSubject, "msg-1", []byte(`{"n":1}`)))

	received := make(chan *natsbroker.Msg, 8)

	consumer, err := NewJetStreamConsumer(
		url,
		dedupSubject,
		WithJetStreamStream(streamName),
		WithJetStreamDurable(dedupDurable),
		WithJetStreamAckWait(jetStreamAckWait),
		WithJetStreamMaxDeliver(jetStreamMaxDelivs),
		WithJetStreamMessageHandler(func(message *natsbroker.Msg) {
			// assert, а не require: обработчик крутится в своей горутине, а
			// require зовёт FailNow, который за пределами тестовой горутины
			// тест не завершает — он роняет его непредсказуемо.
			assert.NoError(t, message.Ack())

			received <- message
		}),
	)
	require.NoError(t, err)
	require.NoError(t, consumer.Run())

	defer func() {
		require.NoError(t, consumer.Stop())
	}()

	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("first message was not delivered")
	}

	select {
	case <-received:
		t.Fatal("duplicate message passed the deduplication window")
	case <-time.After(time.Second):
	}
}

// TestJetStreamConsumer_RedeliversWithoutAck проверяет главное свойство, ради
// которого стрим и заводится: задача, которую не подтвердили, возвращается, а
// не исчезает.
func TestJetStreamConsumer_RedeliversWithoutAck(t *testing.T) {
	publisher, err := NewJetStreamPublisher(url, testStreamConfig())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, publisher.Close())
	}()

	require.NoError(t, publisher.PublishWithID(redeliverSubject, "msg-2", []byte(`{"n":2}`)))

	deliveries := make(chan struct{}, 8)

	consumer, err := NewJetStreamConsumer(
		url,
		redeliverSubject,
		WithJetStreamStream(streamName),
		WithJetStreamDurable(redeliverDurable),
		WithJetStreamAckWait(jetStreamAckWait),
		WithJetStreamMaxDeliver(jetStreamMaxDelivs),
		WithJetStreamMessageHandler(func(_ *natsbroker.Msg) {
			deliveries <- struct{}{}
		}),
	)
	require.NoError(t, err)
	require.NoError(t, consumer.Run())

	defer func() {
		require.NoError(t, consumer.Stop())
	}()

	for range 2 {
		select {
		case <-deliveries:
		case <-time.After(jetStreamAckWait * 3):
			t.Fatal("message was not redelivered after ack_wait")
		}
	}
}

// TestJetStreamConsumer_DurableSurvivesStop — консьюмер обязан пережить
// остановку воркера.
//
// И Unsubscribe, и Drain удаляют durable, созданный самой подпиской: позиция
// чтения уходит вместе с ним, а при нескольких репликах остановка одной
// ломает остальные. Поэтому консьюмер создаётся отдельно (AddConsumer), а
// подписка к нему только привязывается (Bind) — этот тест и есть
// доказательство, что так и сделано.
func TestJetStreamConsumer_DurableSurvivesStop(t *testing.T) {
	publisher, err := NewJetStreamPublisher(url, testStreamConfig())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, publisher.Close())
	}()

	consumer, err := NewJetStreamConsumer(
		url,
		survivorSubject,
		WithJetStreamStream(streamName),
		WithJetStreamDurable(survivorDurable),
		WithJetStreamAckWait(jetStreamAckWait),
		WithJetStreamMaxDeliver(jetStreamMaxDelivs),
	)
	require.NoError(t, err)
	require.NoError(t, consumer.Run())
	require.NoError(t, consumer.Stop())

	connection, err := natsbroker.Connect(url)
	require.NoError(t, err)

	defer connection.Close()

	jetStream, err := connection.JetStream()
	require.NoError(t, err)

	info, err := jetStream.ConsumerInfo(streamName, survivorDurable)
	require.NoError(t, err, "durable-консьюмер удалён остановкой воркера")
	assert.Equal(t, survivorDurable, info.Name)
}

// TestJetStreamConsumer_AppliesChangedSettings — правка сроков в конфиге
// обязана доезжать до уже существующего durable-консьюмера.
//
// Консьюмер переживает перезапуск воркера по построению (см. тест выше), и без
// обновления это оборачивается ловушкой: поменяв ack_wait или предел попыток в
// переменных окружения и перезапустив приложение, вы получаете прежние
// значения — молча, без ошибки и без следа. Стрим при этом обновляется
// (ensureStream зовёт UpdateStream), так что несимметричное поведение
// выглядело бы ещё и как опечатка.
func TestJetStreamConsumer_AppliesChangedSettings(t *testing.T) {
	const (
		settingsSubject = "test-notifications.battle_finished.web_push"
		settingsDurable = "test-settings-worker"

		initialMaxDeliver = 3
		updatedMaxDeliver = 7
		updatedAckWait    = 5 * time.Second
	)

	publisher, err := NewJetStreamPublisher(url, testStreamConfig())
	require.NoError(t, err)

	defer func() {
		require.NoError(t, publisher.Close())
	}()

	first, err := NewJetStreamConsumer(
		url,
		settingsSubject,
		WithJetStreamStream(streamName),
		WithJetStreamDurable(settingsDurable),
		WithJetStreamAckWait(jetStreamAckWait),
		WithJetStreamMaxDeliver(initialMaxDeliver),
	)
	require.NoError(t, err)
	require.NoError(t, first.Run())
	require.NoError(t, first.Stop())

	// Тот же durable, другие сроки: так выглядит перезапуск приложения после
	// правки переменных окружения.
	second, err := NewJetStreamConsumer(
		url,
		settingsSubject,
		WithJetStreamStream(streamName),
		WithJetStreamDurable(settingsDurable),
		WithJetStreamAckWait(updatedAckWait),
		WithJetStreamMaxDeliver(updatedMaxDeliver),
	)
	require.NoError(t, err)
	require.NoError(t, second.Run())

	defer func() {
		require.NoError(t, second.Stop())
	}()

	connection, err := natsbroker.Connect(url)
	require.NoError(t, err)

	defer connection.Close()

	jetStream, err := connection.JetStream()
	require.NoError(t, err)

	info, err := jetStream.ConsumerInfo(streamName, settingsDurable)
	require.NoError(t, err)

	assert.Equal(t, updatedMaxDeliver, info.Config.MaxDeliver, "предел попыток не применился")
	assert.Equal(t, updatedAckWait, info.Config.AckWait, "ack_wait не применился")
}
