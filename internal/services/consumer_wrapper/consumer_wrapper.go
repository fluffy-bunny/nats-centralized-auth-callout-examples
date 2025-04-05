package consumer_wrapper

import (
	contracts_nats "natsauth/internal/contracts/nats"

	// Import NATS Go client
	"github.com/nats-io/nats.go"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
)

type (
	service struct {
		ConsumerBase
		tracerProvider contracts_nats.ITracerProvider
	}
)

var stemService = (*service)(nil)
var _ contracts_nats.IConsumer = stemService

func (s *service) Ctor(tracerProvider contracts_nats.ITracerProvider) (contracts_nats.IConsumer, error) {
	return &service{
		tracerProvider: tracerProvider,
	}, nil
}

func AddTransientIConsumer(builder di.ContainerBuilder) {
	di.AddTransient[contracts_nats.IConsumer](
		builder,
		stemService.Ctor,
	)
}
func (s *service) ConsumeWithContext(handler contracts_nats.MessageHandler, opts ...nats_jetstream.PullConsumeOpt) (nats_jetstream.ConsumeContext, error) {
	return s.Consume(func(msg nats_jetstream.Msg) {
		ctx := s.tracerProvider.ContextFromMessage(msg)

		nMsg := &nats.Msg{
			Subject: msg.Subject(),
			Data:    msg.Data(),
			Reply:   msg.Reply(),
			Header:  msg.Headers(),
		}
		response, _ := s.tracerProvider.StartNewSpan(ctx,
			&contracts_nats.StartNewSpanRequest{
				Msg:           nMsg,
				OperationName: "PublishMsgAsyncWithContext",
				Tags: map[string]string{
					ext.Component: "nats",
					ext.SpanType:  ext.SpanTypeMessageConsumer,
				},
			})
		defer response.Span.Finish()
		handler(ctx, msg)
	}, opts...)
}
