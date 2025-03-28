package datadog

import (
	"context"
	contracts_nats "natsauth/internal/contracts/nats"

	services_consumer_wrapper "natsauth/internal/services/consumer_wrapper"
	services_jetstream_wrapper "natsauth/internal/services/jetstream_wrapper"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	otel "go.opentelemetry.io/otel"
	tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type (
	service struct {
		config *contracts_nats.TraceProviderConfig
	}
	spanCarrier struct {
		span tracer.Span
	}
)

var stemService = (*service)(nil)
var _ contracts_nats.ITracerProvider = (*service)(nil)
var _ contracts_nats.ISpan = (*spanCarrier)(nil)

func (s *service) Ctor(config *contracts_nats.TraceProviderConfig) (contracts_nats.ITracerProvider, error) {
	return &service{
		config: config,
	}, nil
}
func AddSingletonITracerProvider(builder di.ContainerBuilder) {
	di.AddSingleton[contracts_nats.ITracerProvider](
		builder,
		stemService.Ctor,
	)
}
func (s *spanCarrier) Finish() {
	if s.span != nil {
		s.span.Finish()
	}
}
func (s *service) StartNewSpan(ctx context.Context, request *contracts_nats.StartNewSpanRequest) (*contracts_nats.StartNewSpanResponse, error) {
	span, _ := tracer.StartSpanFromContext(ctx, request.OperationName)
	for k, v := range request.Tags {
		span.SetTag(k, v)
	}
	carrier := services_jetstream_wrapper.NewNATSMessageCarrier(request.Msg)
	err := tracer.Inject(span.Context(), carrier)
	if err != nil {
		return nil, err
	}
	ctx = tracer.ContextWithSpan(ctx, span)
	return &contracts_nats.StartNewSpanResponse{
		Span:    &spanCarrier{span: span},
		Context: ctx,
	}, nil
}
func (s *service) ContextFromMessage(msg nats_jetstream.Msg) context.Context {
	propagator := otel.GetTextMapPropagator()
	ctx := context.Background()
	carrier := services_consumer_wrapper.NewNATSMessageCarrier(msg)
	ctx = propagator.Extract(ctx, carrier)
	return ctx
}
