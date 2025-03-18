package jetstream_wrapper

import (
	"context"

	nats "github.com/nats-io/nats.go"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamBase struct {
		Inner nats_jetstream.JetStream
	}
)

func (s *JetStreamBase) SetInner(inner nats_jetstream.JetStream) {
	s.Inner = inner
}
func (s *JetStreamBase) AccountInfo(ctx context.Context) (*nats_jetstream.AccountInfo, error) {
	return s.Inner.AccountInfo(ctx)
}

// Conn returns the underlying NATS connection.
func (s *JetStreamBase) Conn() *nats.Conn {
	return s.Inner.Conn()
}

// Options returns read-only JetStreamOptions used
// when making requests to JetStream.
func (s *JetStreamBase) Options() nats_jetstream.JetStreamOptions {
	return s.Inner.Options()
}

// CreateOrUpdateConsumer creates a consumer on a given stream with
// given config. If consumer already exists, it will be updated (if
// possible). Consumer interface is returned, allowing to operate on a
// consumer (e.g. fetch messages).
func (s *JetStreamBase) CreateOrUpdateConsumer(ctx context.Context, stream string, cfg nats_jetstream.ConsumerConfig) (nats_jetstream.Consumer, error) {
	return s.Inner.CreateOrUpdateConsumer(ctx, stream, cfg)
}

// CreateConsumer creates a consumer on a given stream with given
// config. If consumer already exists and the provided configuration
// differs from its configuration, ErrConsumerExists is returned. If the
// provided configuration is the same as the existing consumer, the
// existing consumer is returned. Consumer interface is returned,
// allowing to operate on a consumer (e.g. fetch messages).
func (s *JetStreamBase) CreateConsumer(ctx context.Context, stream string, cfg nats_jetstream.ConsumerConfig) (nats_jetstream.Consumer, error) {
	return s.Inner.CreateConsumer(ctx, stream, cfg)
}

// UpdateConsumer updates an existing consumer. If consumer does not
// exist, ErrConsumerDoesNotExist is returned. Consumer interface is
// returned, allowing to operate on a consumer (e.g. fetch messages).
func (s *JetStreamBase) UpdateConsumer(ctx context.Context, stream string, cfg nats_jetstream.ConsumerConfig) (nats_jetstream.Consumer, error) {
	return s.Inner.UpdateConsumer(ctx, stream, cfg)
}

// OrderedConsumer returns an OrderedConsumer instance. OrderedConsumer
// are managed by the library and provide a simple way to consume
// messages from a stream. Ordered consumers are ephemeral in-memory
// pull consumers and are resilient to deletes and restarts.
func (s *JetStreamBase) OrderedConsumer(ctx context.Context, stream string, cfg nats_jetstream.OrderedConsumerConfig) (nats_jetstream.Consumer, error) {
	return s.Inner.OrderedConsumer(ctx, stream, cfg)
}

// Consumer returns an interface to an existing consumer, allowing processing
// of messages. If consumer does not exist, ErrConsumerNotFound is
// returned.
func (s *JetStreamBase) Consumer(ctx context.Context, stream string, consumer string) (nats_jetstream.Consumer, error) {
	return s.Inner.Consumer(ctx, stream, consumer)
}

// DeleteConsumer removes a consumer with given name from a stream.
// If consumer does not exist, ErrConsumerNotFound is returned.
func (s *JetStreamBase) DeleteConsumer(ctx context.Context, stream string, consumer string) error {
	return s.Inner.DeleteConsumer(ctx, stream, consumer)
}

// CreateStream creates a new stream with given config and returns an
// interface to operate on it. If stream with given name already exists
// and its configuration differs from the provided one,
// ErrStreamNameAlreadyInUse is returned.
func (s *JetStreamBase) CreateStream(ctx context.Context, cfg nats_jetstream.StreamConfig) (nats_jetstream.Stream, error) {
	return s.Inner.CreateStream(ctx, cfg)
}

// UpdateStream updates an existing stream. If stream does not exist,
// ErrStreamNotFound is returned.
func (s *JetStreamBase) UpdateStream(ctx context.Context, cfg nats_jetstream.StreamConfig) (nats_jetstream.Stream, error) {
	return s.Inner.UpdateStream(ctx, cfg)
}

// CreateOrUpdateStream creates a stream with given config. If stream
// already exists, it will be updated (if possible).
func (s *JetStreamBase) CreateOrUpdateStream(ctx context.Context, cfg nats_jetstream.StreamConfig) (nats_jetstream.Stream, error) {
	return s.Inner.CreateOrUpdateStream(ctx, cfg)
}

// Stream fetches [StreamInfo] and returns a [Stream] interface for a given stream name.
// If stream does not exist, ErrStreamNotFound is returned.
func (s *JetStreamBase) Stream(ctx context.Context, stream string) (nats_jetstream.Stream, error) {
	return s.Inner.Stream(ctx, stream)
}

// StreamNameBySubject returns a stream name stream listening on given
// subject. If no stream is bound to given subject, ErrStreamNotFound
// is returned.
func (s *JetStreamBase) StreamNameBySubject(ctx context.Context, subject string) (string, error) {
	return s.Inner.StreamNameBySubject(ctx, subject)
}

// DeleteStream removes a stream with given name. If stream does not
// exist, ErrStreamNotFound is returned.
func (s *JetStreamBase) DeleteStream(ctx context.Context, stream string) error {
	return s.Inner.DeleteStream(ctx, stream)
}

// ListStreams returns StreamInfoLister, enabling iterating over a
// channel of stream infos.
func (s *JetStreamBase) ListStreams(context.Context, ...nats_jetstream.StreamListOpt) nats_jetstream.StreamInfoLister {
	return s.Inner.ListStreams(context.Background())
}

// StreamNames returns a  StreamNameLister, enabling iterating over a
// channel of stream names.
func (s *JetStreamBase) StreamNames(context.Context, ...nats_jetstream.StreamListOpt) nats_jetstream.StreamNameLister {
	return s.Inner.StreamNames(context.Background())
}

// Publish performs a synchronous publish to a stream and waits for ack
// from server. It accepts subject name (which must be bound to a stream)
// and message payload.
func (s *JetStreamBase) Publish(ctx context.Context, subject string, payload []byte, opts ...nats_jetstream.PublishOpt) (*nats_jetstream.PubAck, error) {
	return s.Inner.Publish(ctx, subject, payload, opts...)
}

// PublishMsg performs a synchronous publish to a stream and waits for
// ack from server. It accepts subject name (which must be bound to a
// stream) and nats.Message.
func (s *JetStreamBase) PublishMsg(ctx context.Context, msg *nats.Msg, opts ...nats_jetstream.PublishOpt) (*nats_jetstream.PubAck, error) {
	return s.Inner.PublishMsg(ctx, msg, opts...)
}

// PublishAsync performs a publish to a stream and returns
// [PubAckFuture] interface, not blocking while waiting for an
// acknowledgement. It accepts subject name (which must be bound to a
// stream) and message payload.
//
// PublishAsync does not guarantee that the message has been
// received by the server. It only guarantees that the message has been
// sent to the server and thus messages can be stored in the stream
// out of order in case of retries.
func (s *JetStreamBase) PublishAsync(subject string, payload []byte, opts ...nats_jetstream.PublishOpt) (nats_jetstream.PubAckFuture, error) {
	return s.Inner.PublishAsync(subject, payload, opts...)
}

// PublishMsgAsync performs a publish to a stream and returns
// [PubAckFuture] interface, not blocking while waiting for an
// acknowledgement. It accepts subject name (which must
// be bound to a stream) and nats.Message.
//
// PublishMsgAsync does not guarantee that the message has been
// sent to the server and thus messages can be stored in the stream
// received by the server. It only guarantees that the message has been
// out of order in case of retries.
func (s *JetStreamBase) PublishMsgAsync(msg *nats.Msg, opts ...nats_jetstream.PublishOpt) (nats_jetstream.PubAckFuture, error) {
	return s.Inner.PublishMsgAsync(msg, opts...)
}

// PublishAsyncPending returns the number of async publishes outstanding
// for this context. An outstanding publish is one that has been
// sent by the publisher but has not yet received an ack.
func (s *JetStreamBase) PublishAsyncPending() int {
	return s.Inner.PublishAsyncPending()
}

// PublishAsyncComplete returns a channel that will be closed when all
// outstanding asynchronously published messages are acknowledged by the
// server.
func (s *JetStreamBase) PublishAsyncComplete() <-chan struct{} {
	return s.Inner.PublishAsyncComplete()
}

// CleanupPublisher will cleanup the publishing side of JetStreamContext.
//
// This will unsubscribe from the internal reply subject if needed.
// All pending async publishes will fail with ErrJetStreamContextClosed.
//
// If an error handler was provided, it will be called for each pending async
// publish and PublishAsyncComplete will be closed.
//
// After completing JetStreamContext is still usable - internal subscription
// will be recreated on next publish, but the acks from previous publishes will
// be lost.
func (s *JetStreamBase) CleanupPublisher() {
	s.Inner.CleanupPublisher()
}

// KeyValue will lookup and bind to an existing KeyValue store.
//
// If the KeyValue store with given name does not exist,
// ErrBucketNotFound will be returned.
func (s *JetStreamBase) KeyValue(ctx context.Context, bucket string) (nats_jetstream.KeyValue, error) {
	return s.Inner.KeyValue(ctx, bucket)
}

// CreateKeyValue will create a KeyValue store with the given
// configuration.
//
// If a KeyValue store with the same name already exists and the
// configuration is different, ErrBucketExists will be returned.
func (s *JetStreamBase) CreateKeyValue(ctx context.Context, cfg nats_jetstream.KeyValueConfig) (nats_jetstream.KeyValue, error) {
	return s.Inner.CreateKeyValue(ctx, cfg)
}

// UpdateKeyValue will update an existing KeyValue store with the given
// configuration.
//
// If a KeyValue store with the given name does not exist, ErrBucketNotFound
// will be returned.
func (s *JetStreamBase) UpdateKeyValue(ctx context.Context, cfg nats_jetstream.KeyValueConfig) (nats_jetstream.KeyValue, error) {
	return s.Inner.UpdateKeyValue(ctx, cfg)
}

// CreateOrUpdateKeyValue will create a KeyValue store if it does not
// exist or update an existing KeyValue store with the given
// configuration (if possible).
func (s *JetStreamBase) CreateOrUpdateKeyValue(ctx context.Context, cfg nats_jetstream.KeyValueConfig) (nats_jetstream.KeyValue, error) {
	return s.Inner.CreateOrUpdateKeyValue(ctx, cfg)
}

// DeleteKeyValue will delete this KeyValue store.
//
// If the KeyValue store with given name does not exist,
// ErrBucketNotFound will be returned.
func (s *JetStreamBase) DeleteKeyValue(ctx context.Context, bucket string) error {
	return s.Inner.DeleteKeyValue(ctx, bucket)
}

// KeyValueStoreNames is used to retrieve a list of key value store
// names. It returns a KeyValueNamesLister exposing a channel to read
// the names from. The lister will always close the channel when done
// (either all names have been read or an error occurred) and therefore
// can be used in range loops.
func (s *JetStreamBase) KeyValueStoreNames(ctx context.Context) nats_jetstream.KeyValueNamesLister {
	return s.Inner.KeyValueStoreNames(ctx)
}

// KeyValueStores is used to retrieve a list of key value store
// statuses. It returns a KeyValueLister exposing a channel to read the
// statuses from. The lister will always close the channel when done
// (either all statuses have been read or an error occurred) and
// therefore can be used in range loops.
func (s *JetStreamBase) KeyValueStores(ctx context.Context) nats_jetstream.KeyValueLister {
	return s.Inner.KeyValueStores(ctx)
}

// ObjectStore will look up and bind to an existing object store
// instance.
//
// If the object store with given name does not exist, ErrBucketNotFound
// will be returned.
func (s *JetStreamBase) ObjectStore(ctx context.Context, bucket string) (nats_jetstream.ObjectStore, error) {
	return s.Inner.ObjectStore(ctx, bucket)
}

// CreateObjectStore will create a new object store with the given
// configuration.
//
// If the object store with given name already exists, ErrBucketExists
// will be returned.
func (s *JetStreamBase) CreateObjectStore(ctx context.Context, cfg nats_jetstream.ObjectStoreConfig) (nats_jetstream.ObjectStore, error) {
	return s.Inner.CreateObjectStore(ctx, cfg)
}

// UpdateObjectStore will update an existing object store with the given
// configuration.
//
// If the object store with given name does not exist, ErrBucketNotFound
// will be returned.
func (s *JetStreamBase) UpdateObjectStore(ctx context.Context, cfg nats_jetstream.ObjectStoreConfig) (nats_jetstream.ObjectStore, error) {
	return s.Inner.UpdateObjectStore(ctx, cfg)
}

// CreateOrUpdateObjectStore will create a new object store with the given
// configuration if it does not exist, or update an existing object store
// with the given configuration.
func (s *JetStreamBase) CreateOrUpdateObjectStore(ctx context.Context, cfg nats_jetstream.ObjectStoreConfig) (nats_jetstream.ObjectStore, error) {
	return s.Inner.CreateOrUpdateObjectStore(ctx, cfg)
}

// DeleteObjectStore will delete the provided object store.
//
// If the object store with given name does not exist, ErrBucketNotFound
// will be returned.
func (s *JetStreamBase) DeleteObjectStore(ctx context.Context, bucket string) error {
	return s.Inner.DeleteObjectStore(ctx, bucket)
}

// ObjectStoreNames is used to retrieve a list of bucket names.
// It returns an ObjectStoreNamesLister exposing a channel to receive
// the names of the object stores.
//
// The lister will always close the channel when done (either all names
// have been read or an error occurred) and therefore can be used in a
// for-range loop.
func (s *JetStreamBase) ObjectStoreNames(ctx context.Context) nats_jetstream.ObjectStoreNamesLister {
	return s.Inner.ObjectStoreNames(ctx)
}

// ObjectStores is used to retrieve a list of bucket statuses.
// It returns an ObjectStoresLister exposing a channel to receive
// the statuses of the object stores.
//
// The lister will always close the channel when done (either all statuses
// have been read or an error occurred) and therefore can be used in a
// for-range loop.
func (s *JetStreamBase) ObjectStores(ctx context.Context) nats_jetstream.ObjectStoresLister {
	return s.Inner.ObjectStores(ctx)
}
