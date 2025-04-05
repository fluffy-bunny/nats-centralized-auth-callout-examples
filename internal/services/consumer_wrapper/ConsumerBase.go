package consumer_wrapper

import (
	"context"

	nats_jetstream "github.com/nats-io/nats.go/jetstream"
)

type (
	ConsumerBase struct {
		Inner nats_jetstream.Consumer
	}
)

func (s *ConsumerBase) SetInner(inner nats_jetstream.Consumer) {
	s.Inner = inner
}

// Fetch is used to retrieve up to a provided number of messages from a
// stream. This method will send a single request and deliver either all
// requested messages unless time out is met earlier. Fetch timeout
// defaults to 30 seconds and can be configured using FetchMaxWait
// option.
//
// By default, Fetch uses a 5s idle heartbeat for requests longer than
// 10 seconds. For shorter requests, the idle heartbeat is disabled.
// This can be configured using FetchHeartbeat option. If a client does
// not receive a heartbeat message from a stream for more than 2 times
// the idle heartbeat setting, Fetch will return [ErrNoHeartbeat].
//
// Fetch is non-blocking and returns MessageBatch, exposing a channel
// for delivered messages.
//
// Messages channel is always closed, thus it is safe to range over it
// without additional checks. After the channel is closed,
// MessageBatch.Error() should be checked to see if there was an error
// during message delivery (e.g. missing heartbeat).
func (s ConsumerBase) Fetch(batch int, opts ...nats_jetstream.FetchOpt) (nats_jetstream.MessageBatch, error) {
	return s.Inner.Fetch(batch, opts...)
}

// FetchBytes is used to retrieve up to a provided bytes from the
// stream. This method will send a single request and deliver the
// provided number of bytes unless time out is met earlier. FetchBytes
// timeout defaults to 30 seconds and can be configured using
// FetchMaxWait option.
//
// By default, FetchBytes uses a 5s idle heartbeat for requests longer than
// 10 seconds. For shorter requests, the idle heartbeat is disabled.
// This can be configured using FetchHeartbeat option. If a client does
// not receive a heartbeat message from a stream for more than 2 times
// the idle heartbeat setting, Fetch will return ErrNoHeartbeat.
//
// FetchBytes is non-blocking and returns MessageBatch, exposing a channel
// for delivered messages.
//
// Messages channel is always closed, thus it is safe to range over it
// without additional checks. After the channel is closed,
// MessageBatch.Error() should be checked to see if there was an error
// during message delivery (e.g. missing heartbeat).
func (s ConsumerBase) FetchBytes(maxBytes int, opts ...nats_jetstream.FetchOpt) (nats_jetstream.MessageBatch, error) {
	return s.Inner.FetchBytes(maxBytes, opts...)
}

// FetchNoWait is used to retrieve up to a provided number of messages
// from a stream. Unlike Fetch, FetchNoWait will only deliver messages
// that are currently available in the stream and will not wait for new
// messages to arrive, even if batch size is not met.
//
// FetchNoWait is non-blocking and returns MessageBatch, exposing a
// channel for delivered messages.
//
// Messages channel is always closed, thus it is safe to range over it
// without additional checks. After the channel is closed,
// MessageBatch.Error() should be checked to see if there was an error
// during message delivery (e.g. missing heartbeat).
func (s ConsumerBase) FetchNoWait(batch int) (nats_jetstream.MessageBatch, error) {
	return s.Inner.FetchNoWait(batch)
}

// Consume will continuously receive messages and handle them
// with the provided callback function. Consume can be configured using
// PullConsumeOpt options:
//
//   - Error handling and monitoring can be configured using ConsumeErrHandler
//     option, which provides information about errors encountered during
//     consumption (both transient and terminal)
//   - Consume can be configured to stop after a certain number of
//     messages is received using StopAfter option.
//   - Consume can be optimized for throughput or memory usage using
//     PullExpiry, PullMaxMessages, PullMaxBytes and PullHeartbeat options.
//     Unless there is a specific use case, these options should not be used.
//
// Consume returns a ConsumeContext, which can be used to stop or drain
// the consumer.
func (s ConsumerBase) Consume(handler nats_jetstream.MessageHandler, opts ...nats_jetstream.PullConsumeOpt) (nats_jetstream.ConsumeContext, error) {
	return s.Inner.Consume(handler, opts...)
}

// Messages returns MessagesContext, allowing continuously iterating
// over messages on a stream. Messages can be configured using
// PullMessagesOpt options:
//
//   - Messages can be optimized for throughput or memory usage using
//     PullExpiry, PullMaxMessages, PullMaxBytes and PullHeartbeat options.
//     Unless there is a specific use case, these options should not be used.
//   - WithMessagesErrOnMissingHeartbeat can be used to enable/disable
//     erroring out on MessagesContext.Next when a heartbeat is missing.
//     This option is enabled by default.
func (s ConsumerBase) Messages(opts ...nats_jetstream.PullMessagesOpt) (nats_jetstream.MessagesContext, error) {
	return s.Inner.Messages(opts...)
}

// Next is used to retrieve the next message from the consumer. This
// method will block until the message is retrieved or timeout is
// reached.
func (s ConsumerBase) Next(opts ...nats_jetstream.FetchOpt) (nats_jetstream.Msg, error) {
	return s.Inner.Next(opts...)
}

// Info fetches current ConsumerInfo from the server.
func (s ConsumerBase) Info(ctx context.Context) (*nats_jetstream.ConsumerInfo, error) {
	return s.Inner.Info(ctx)
}

// CachedInfo returns ConsumerInfo currently cached on this consumer.
// This method does not perform any network requests. The cached
// ConsumerInfo is updated on every call to Info and Update.
func (s ConsumerBase) CachedInfo() *nats_jetstream.ConsumerInfo {
	return s.Inner.CachedInfo()
}
