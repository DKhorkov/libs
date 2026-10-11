// Package nats provides tools for comfort work during event processing with
// NATS broker.
//
// Core mode — CommonPublisher and CommonConsumer — is an at-most-once
// broadcast: it has neither memory, nor acknowledgements. JetStream mode —
// JetStreamPublisher and JetStreamConsumer — is an at-least-once queue: the
// stream keeps messages on disk, a durable consumer holds the reading position
// on the server, and an unacknowledged message is redelivered.
//
// Dropped — сколько сообщений подписка выбросила из-за полного канала.
// Асинхронный обработчик ошибок узнаёт о slow consumer один раз на эпизод:
// nats.go снимает флаг на следующем доставленном сообщении. Масштаб потери
// виден только этим счётчиком. После Stop подписка закрыта, и счётчик
// отвечает ошибкой.
package nats
