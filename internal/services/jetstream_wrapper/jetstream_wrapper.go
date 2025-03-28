package jetstream_wrapper

import (
	"context"
	contracts_nats "natsauth/internal/contracts/nats"
	propagators "natsauth/internal/propagators"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	nats "github.com/nats-io/nats.go" // Import NATS Go client
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type (
	service struct {
		JetStreamBase
		tracerProvider contracts_nats.ITracerProvider
	}
)

var stemService = (*service)(nil)
var _ contracts_nats.IJetStream = (*service)(nil)

func (s *service) Ctor(tracerProvider contracts_nats.ITracerProvider) (contracts_nats.IJetStream, error) {
	return &service{
		tracerProvider: tracerProvider,
	}, nil
}

func AddTransientIJetStream(builder di.ContainerBuilder) {
	di.AddTransient[contracts_nats.IJetStream](
		builder,
		stemService.Ctor,
	)
}

func (s *service) Publish(ctx context.Context, subject string, payload []byte, opts ...nats_jetstream.PublishOpt) (*nats_jetstream.PubAck, error) {
	msg := &nats.Msg{
		Subject: subject,
		Data:    payload,
	}
	return s.PublishMsg(ctx, msg, opts...)
}

func (s *service) PublishMsg(ctx context.Context, msg *nats.Msg, opts ...nats_jetstream.PublishOpt) (*nats_jetstream.PubAck, error) {
	ctx = propagators.EnsureCorrelationId(ctx)
	response, _ := s.tracerProvider.StartNewSpan(
		ctx,
		&contracts_nats.StartNewSpanRequest{
			Msg:           msg,
			OperationName: "PublishMsg",
			Tags: map[string]string{
				ext.Component: "nats",
				ext.SpanType:  ext.SpanTypeMessageProducer,
			},
		})
	defer response.Span.Finish()

	return s.Inner.PublishMsg(response.Context, msg, opts...)
}

func (s *service) PublishMsgAsyncWithContext(ctx context.Context, msg *nats.Msg, opts ...nats_jetstream.PublishOpt) (nats_jetstream.PubAckFuture, error) {
	ctx = propagators.EnsureCorrelationId(ctx)
	response, _ := s.tracerProvider.StartNewSpan(
		ctx,
		&contracts_nats.StartNewSpanRequest{
			Msg:           msg,
			OperationName: "PublishMsgAsyncWithContext",
			Tags: map[string]string{
				ext.Component: "nats",
				ext.SpanType:  ext.SpanTypeMessageProducer,
			},
		})
	defer response.Span.Finish()

	return s.Inner.PublishMsgAsync(msg, opts...)
}

func (s *service) PublishAsyncWithContext(ctx context.Context, subject string, payload []byte, opts ...nats_jetstream.PublishOpt) (nats_jetstream.PubAckFuture, error) {
	msg := &nats.Msg{
		Subject: subject,
		Data:    payload,
	}
	return s.PublishMsgAsyncWithContext(ctx, msg, opts...)
}
